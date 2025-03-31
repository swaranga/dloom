package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SourceDir string                    `yaml:"sourceDir"`
	TargetDir string                    `yaml:"targetDir"`
	BackupDir string                    `yaml:"backupDir"`
	Force     bool                      `yaml:"force"`
	Verbose   bool                      `yaml:"verbose"`
	DryRun    bool                      `yaml:"dryRun"`
	Packages  map[string]*PackageConfig `yaml:"packages"`
}

type PackageConfig struct {
	SourceDir  string                 `yaml:"sourceDir"`
	TargetDir  string                 `yaml:"targetDir"`
	BackupDir  string                 `yaml:"backupDir"`
	Force      *bool                  `yaml:"force"`
	Verbose    *bool                  `yaml:"verbose"`
	DryRun     bool                   `yaml:"dryRun"`
	Conditions *ConditionSet          `yaml:"conditions"`
	Files      map[string]*FileConfig `yaml:"files"`
}

type FileConfig struct {
	TargetDir  string        `yaml:"targetDir"`
	TargetName string        `yaml:"targetName"`
	BackupDir  string        `yaml:"backupDir"`
	Force      *bool         `yaml:"force"`
	Verbose    *bool         `yaml:"verbose"`
	DryRun     bool          `yaml:"dryRun"`
	Conditions *ConditionSet `yaml:"conditions"`
}

type ConditionSet struct {
	OS                []string          `yaml:"os"`
	Distro            []string          `yaml:"distro"`
	Executable        []string          `yaml:"executable"`
	ExecutableVersion map[string]string `yaml:"executable_version"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "~"
	}

	return &Config{
		SourceDir: ".",     // Current directory by default
		TargetDir: homeDir, // User's home directory by default
		BackupDir: filepath.Join(homeDir, ".dloom/backups"),
		Force:     false,
		Verbose:   false,
		DryRun:    false,
		Packages:  make(map[string]*PackageConfig),
	}
}

// LoadConfig loads configuration from the specified file
// If the file doesn't exist, returns default config
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// If no config path specified, look in default locations
	// First try to find a
	if configPath == "" {
		// First, try current directory and see if a dloom/config.yaml exists
		// If not, try ~/.config/dloom/config.yaml
		currentDir, err := os.Getwd()
		if err == nil {
			configPath = filepath.Join(currentDir, "dloom", "config.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Next, try ~/.config/dloom/config.yaml
				homeDir, err := os.UserHomeDir()
				if err == nil {
					configPath = filepath.Join(homeDir, ".config", "dloom", "config.yaml")
					if _, err := os.Stat(configPath); os.IsNotExist(err) {
						// No config file found, use defaults
						return config, nil
					}
				}
			}
		}
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return defaults
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// GetSourcePath returns the full path for a source package
func (c *Config) GetSourcePath(packageName string) string {
	// Check for package-specific source directory
	if pkg, exists := c.Packages[packageName]; exists && pkg.SourceDir != "" {
		return pkg.SourceDir
	}

	// Fall back to global source directory + package name
	return filepath.Join(c.SourceDir, packageName)
}

// GetTargetPath returns the full path for a target in the target directory
func (c *Config) GetTargetPath(packageName, relativePath string) string {
	// Get the effective configuration for this file
	effectiveConfig := c.GetEffectiveConfig(packageName, relativePath)

	// Construct the target path using the effective target directory
	return filepath.Join(effectiveConfig.TargetDir, effectiveConfig.TargetName)
}

// GetBackupPath returns the path where a file should be backed up
func (c *Config) GetBackupPath(packageName, relativePath string) string {
	// Get the effective configuration for this file
	effectiveConfig := c.GetEffectiveConfig(packageName, relativePath)

	// If backup directory is empty, backups are disabled
	if effectiveConfig.BackupDir == "" {
		return ""
	}

	// Otherwise, construct the backup path
	return filepath.Join(effectiveConfig.BackupDir, relativePath)
}

// ShouldBeVerbose returns whether verbose output should be enabled for a specific file
func (c *Config) ShouldBeVerbose(packageName, relativePath string) bool {
	// Get the effective configuration for this file
	effectiveConfig := c.GetEffectiveConfig(packageName, relativePath)

	// The Verbose field should never be nil in the effective config
	// since it's initialized with the global value
	return *effectiveConfig.Verbose
}

// IsDryRun returns whether dry run mode is enabled for a specific file
func (c *Config) IsDryRun(packageName, relativePath string) bool {
	// Get the effective configuration for this file
	effectiveConfig := c.GetEffectiveConfig(packageName, relativePath)

	// Return the effective dry run setting
	return effectiveConfig.DryRun
}

// ShouldForce returns whether force mode is enabled for a specific file
func (c *Config) ShouldForce(packageName, relativePath string) bool {
	// Get the effective configuration for this file
	effectiveConfig := c.GetEffectiveConfig(packageName, relativePath)

	// The Force field should never be nil in the effective config
	// since it's initialized with the global value
	return *effectiveConfig.Force
}

// GetEffectiveConfig returns the effective configuration for a specific file
// by merging global, package-level, and file-specific settings
func (c *Config) GetEffectiveConfig(packageName, relativePath string) *FileConfig {
	// Start with default settings derived from global config
	effectiveConfig := &FileConfig{
		TargetDir:  c.TargetDir,
		TargetName: relativePath, // Default to the file name
		BackupDir:  c.BackupDir,
		Force:      &c.Force,
		Verbose:    &c.Verbose,
		DryRun:     c.DryRun,
		Conditions: nil, // No conditions at global level
	}

	// Apply package-level settings if they exist
	pkg, pkgExists := c.Packages[packageName]
	if pkgExists {
		// Override with package settings where specified
		if pkg.TargetDir != "" {
			effectiveConfig.TargetDir = pkg.TargetDir
		}

		if pkg.BackupDir != "" {
			effectiveConfig.BackupDir = pkg.BackupDir
		}

		if pkg.Force != nil {
			effectiveConfig.Force = pkg.Force
		}

		if pkg.Verbose != nil {
			effectiveConfig.Verbose = pkg.Verbose
		}

		if pkg.DryRun {
			effectiveConfig.DryRun = pkg.DryRun
		}

		// Apply package conditions
		effectiveConfig.Conditions = pkg.Conditions

		// Check for file-specific settings
		var fileConfig *FileConfig

		// First, try exact file match
		if fc, exists := pkg.Files[relativePath]; exists {
			fileConfig = fc
		} else {
			// Try regex matches
			for pattern, fc := range pkg.Files {
				// Only process entries that start with "regex:"
				if strings.HasPrefix(pattern, "regex:") {
					// Extract the actual regex pattern
					regexPattern := strings.TrimPrefix(pattern, "regex:")

					matched, err := regexp.MatchString(regexPattern, relativePath)
					if err == nil && matched {
						fileConfig = fc
						break
					}
				}
			}
		}

		// Apply file-specific settings if found
		if fileConfig != nil {
			if fileConfig.TargetDir != "" {
				effectiveConfig.TargetDir = fileConfig.TargetDir
			}

			if fileConfig.TargetName != "" {
				// First extract the directory from the relative path
				// and then join it with the target name
				relativePathDir := filepath.Dir(relativePath)
				effectiveConfig.TargetName = filepath.Join(relativePathDir, fileConfig.TargetName)
			}

			if fileConfig.BackupDir != "" {
				effectiveConfig.BackupDir = fileConfig.BackupDir
			}

			if fileConfig.Force != nil {
				effectiveConfig.Force = fileConfig.Force
			}

			if fileConfig.Verbose != nil {
				effectiveConfig.Verbose = fileConfig.Verbose
			}

			if fileConfig.DryRun {
				effectiveConfig.DryRun = fileConfig.DryRun
			}

			// For conditions, file conditions completely override package conditions
			// rather than merging them
			if fileConfig.Conditions != nil {
				effectiveConfig.Conditions = fileConfig.Conditions
			}
		}
	}

	return effectiveConfig
}

// MatchesConditions checks if the current environment matches the given conditions
func (c *Config) MatchesConditions(conditions *ConditionSet) bool {
	if conditions == nil {
		return true // No conditions means always match
	}

	// Check OS conditions
	if len(conditions.OS) > 0 && !c.matchesOSCondition(conditions.OS) {
		return false
	}

	// Check Linux distro condition
	if len(conditions.Distro) > 0 && !c.matchesDistroCondition(conditions.Distro) {
		return false
	}

	// Check executable conditions
	if len(conditions.Executable) > 0 && !c.matchesExecutableCondition(conditions.Executable) {
		return false
	}

	// Check executable version conditions
	if len(conditions.ExecutableVersion) > 0 && !c.matchesExecutableVersionCondition(conditions.ExecutableVersion) {
		return false
	}

	return true
}

// matchesExecutableVersionCondition checks if executables meet version requirements
func (c *Config) matchesExecutableVersionCondition(versionConditions map[string]string) bool {
	if len(versionConditions) == 0 {
		return true // No version conditions means always match
	}

	for execName, versionConstraint := range versionConditions {
		// First check if the executable exists
		execPath, err := exec.LookPath(execName)
		if err != nil {
			// Executable not found
			return false
		}

		// Get the version of the executable
		version, err := getExecutableVersion(execPath, execName)
		if err != nil {
			// Couldn't determine version
			if c.Verbose {
				fmt.Printf("Warning: Could not determine version of %s: %v\n", execName, err)
			}
			return false
		}

		// Compare the version against the constraint
		if !versionMeetsConstraint(version, versionConstraint) {
			// Version doesn't meet constraint
			return false
		}
	}

	// All version conditions passed
	return true
}

// getExecutableVersion attempts to determine the version of an executable
func getExecutableVersion(execPath, execName string) (string, error) {
	// Common version flags for different programs
	versionFlags := [][]string{
		{"--version"},
		{"-v"},
		{"-V"},
		{"version"},
		{"--ver"},
	}

	// Try each version flag
	for _, flags := range versionFlags {
		cmd := exec.Command(execPath, flags...)
		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			// Successfully got output, try to extract version
			version := extractVersionFromOutput(string(output))
			if version != "" {
				return version, nil
			}
		}
	}

	// Special handling for specific executables
	switch execName {
	case "node":
		// Node.js often works with -v
		cmd := exec.Command(execPath, "-v")
		output, err := cmd.CombinedOutput()
		if err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	case "python", "python3":
		// Python has a specific version flag
		cmd := exec.Command(execPath, "--version")
		output, err := cmd.CombinedOutput()
		if err == nil {
			return extractVersionFromOutput(string(output)), nil
		}
	}

	return "", fmt.Errorf("could not determine version")
}

// extractVersionFromOutput tries to find a version string in command output
func extractVersionFromOutput(output string) string {
	// Common version patterns
	patterns := []*regexp.Regexp{
		// Match "version X.Y.Z"
		regexp.MustCompile(`(?i)version\s+(\d+\.\d+\.\d+)`),
		// Match "vX.Y.Z"
		regexp.MustCompile(`v(\d+\.\d+\.\d+)`),
		// Match "X.Y.Z"
		regexp.MustCompile(`(\d+\.\d+\.\d+)`),
		// Match "X.Y"
		regexp.MustCompile(`(\d+\.\d+)\b`),
	}

	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(output)
		if len(matches) >= 2 {
			return matches[1]
		}
	}

	return ""
}

// versionMeetsConstraint checks if a version meets a version constraint
func versionMeetsConstraint(version, constraint string) bool {
	// Extract operator and required version
	operator := ""
	requiredVersion := constraint

	// Check for comparison operators
	if strings.HasPrefix(constraint, ">=") {
		operator = ">="
		requiredVersion = strings.TrimPrefix(constraint, ">=")
	} else if strings.HasPrefix(constraint, ">") {
		operator = ">"
		requiredVersion = strings.TrimPrefix(constraint, ">")
	} else if strings.HasPrefix(constraint, "<=") {
		operator = "<="
		requiredVersion = strings.TrimPrefix(constraint, "<=")
	} else if strings.HasPrefix(constraint, "<") {
		operator = "<"
		requiredVersion = strings.TrimPrefix(constraint, "<")
	} else if strings.HasPrefix(constraint, "=") {
		operator = "="
		requiredVersion = strings.TrimPrefix(constraint, "=")
	}

	// Trim whitespace
	requiredVersion = strings.TrimSpace(requiredVersion)
	version = strings.TrimSpace(version)

	// Compare versions
	comparison := compareVersions(version, requiredVersion)

	switch operator {
	case ">=":
		return comparison >= 0
	case ">":
		return comparison > 0
	case "<=":
		return comparison <= 0
	case "<":
		return comparison < 0
	case "=", "":
		return comparison == 0
	default:
		return false
	}
}

// compareVersions compares two version strings
// Returns:
//
//	-1 if v1 < v2
//	 0 if v1 == v2
//	 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	// Split versions into components
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	// Compare each component
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		// If we've run out of components in v1, v1 is less than v2
		if i >= len(parts1) {
			return -1
		}

		// If we've run out of components in v2, v1 is greater than v2
		if i >= len(parts2) {
			return 1
		}

		// Try to convert components to integers
		num1, err1 := strconv.Atoi(parts1[i])
		num2, err2 := strconv.Atoi(parts2[i])

		// If both convert to integers, compare numerically
		if err1 == nil && err2 == nil {
			if num1 < num2 {
				return -1
			} else if num1 > num2 {
				return 1
			}
			// If equal, continue to next component
			continue
		}

		// If not both numbers, compare as strings
		if comp := strings.Compare(parts1[i], parts2[i]); comp != 0 {
			return comp
		}
	}

	// All components are equal
	return 0
}

// matchesExecutableCondition checks if all required executables are available in PATH
func (c *Config) matchesExecutableCondition(executableConditions []string) bool {
	if len(executableConditions) == 0 {
		return true // No executable conditions means always match
	}

	for _, execName := range executableConditions {
		// Look for the executable in PATH
		_, err := exec.LookPath(execName)
		if err != nil {
			// Executable not found
			return false
		}
	}

	// All executables were found
	return true
}

// matchesOSCondition checks if the current OS matches any of the provided OS conditions
func (c *Config) matchesOSCondition(osConditions []string) bool {
	if len(osConditions) == 0 {
		return true // No OS conditions means always match
	}

	currentOS := runtime.GOOS

	for _, os := range osConditions {
		if os == currentOS {
			return true
		}
	}

	return false
}

// matchesDistroCondition checks if the current Linux distribution matches any of the provided distro conditions
func (c *Config) matchesDistroCondition(distroConditions []string) bool {
	if len(distroConditions) == 0 {
		return true // No distro conditions means always match
	}

	// If not on Linux, distro conditions don't apply
	if runtime.GOOS != "linux" {
		return true
	}

	// Try to detect the Linux distribution
	currentDistro := c.detectLinuxDistribution()
	if currentDistro == "" {
		// Couldn't detect distribution
		return false
	}

	// Check if the current distribution matches any of the conditions
	for _, distro := range distroConditions {
		if strings.EqualFold(distro, currentDistro) {
			return true
		}
	}

	return false
}

// detectLinuxDistribution attempts to determine the current Linux distribution
func (c *Config) detectLinuxDistribution() string {
	// Try reading /etc/os-release first (most modern distros)
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		return parseOSRelease(string(data))
	}

	// Try other common distribution files
	for _, file := range []string{"/etc/lsb-release", "/etc/debian_version", "/etc/fedora-release", "/etc/redhat-release"} {
		if _, err := os.Stat(file); err == nil {
			// Extract distribution name from filename
			base := filepath.Base(file)
			switch {
			case strings.Contains(base, "debian"):
				return "debian"
			case strings.Contains(base, "ubuntu") || strings.Contains(base, "lsb"):
				return "ubuntu"
			case strings.Contains(base, "fedora"):
				return "fedora"
			case strings.Contains(base, "redhat"):
				return "rhel"
			}
		}
	}

	// Check for Arch Linux
	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return "arch"
	}

	// Couldn't determine the distribution
	return ""
}

// parseOSRelease extracts the distribution ID from /etc/os-release content
func parseOSRelease(content string) string {
	// Look for ID= line
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "ID=") {
			id := strings.TrimPrefix(line, "ID=")
			// Remove quotes if present
			id = strings.Trim(id, "\"'")
			return id
		}
	}
	return ""
}
