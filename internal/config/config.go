// Package config provides configuration handling for dloom
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	// SourceDir is the base directory for source files, defaults to current directory
	SourceDir string `yaml:"sourceDir"`

	// TargetDir is the base directory where symlinks will be created, defaults to home directory
	TargetDir string `yaml:"targetDir"`

	// BackupDir is where existing files will be backed up before linking
	// If empty, no backups will be created
	BackupDir string `yaml:"backupDir"`

	// Force determines whether to replace existing files without prompting
	Force bool `yaml:"force"`

	// Verbose enables detailed output
	Verbose bool `yaml:"verbose"`

	// DryRun shows what would happen without making any changes
	DryRun bool `yaml:"dryRun"`

	// Packages holds package-specific configurations
	Packages map[string]*PackageConfig `yaml:"packages"`
}

// PackageConfig holds configuration for a specific package
type PackageConfig struct {
	// SourceDir overrides the global source directory for this package
	SourceDir string `yaml:"sourceDir"`

	// TargetDir overrides the global target directory for this package
	TargetDir string `yaml:"targetDir"`

	// BackupDir overrides the global backup directory for this package
	BackupDir string `yaml:"backupDir"`

	// Force overrides the global force setting for this package
	Force *bool `yaml:"force"`

	// Conditions for conditional linking of this package
	Conditions *ConditionSet `yaml:"conditions"`

	// Files holds file-specific configurations within this package
	Files map[string]*FileConfig `yaml:"files"`
}

// FileConfig holds configuration for a specific file within a package
type FileConfig struct {
	// TargetPath specifies an exact target path for this file
	TargetPath string `yaml:"targetPath"`

	// Conditions for conditional linking of this file
	Conditions *ConditionSet `yaml:"conditions"`
}

// ConditionSet holds conditions for conditional linking
type ConditionSet struct {
	// OS conditions (e.g., "linux", "darwin", "windows")
	OS []string `yaml:"os"`

	// Distro conditions for Linux distributions (e.g., "ubuntu", "arch")
	Distro []string `yaml:"distro"`

	// Executable conditions check if executables exist in PATH
	Executable []string `yaml:"executable"`

	// ExecutableVersion checks versions of executables
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
	if configPath == "" {
		// First, try current directory
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

// SaveConfig saves the configuration to the specified file
func SaveConfig(config *Config, configPath string) error {
	// If no path specified, use default
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		// Ensure directory exists
		configDir := filepath.Join(homeDir, ".config", "dloom")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		configPath = filepath.Join(configDir, "config.yaml")
	}

	// Convert to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetSourcePath returns the full path for a source package
func (c *Config) GetSourcePath(packageName string) string {
	if pkg, exists := c.Packages[packageName]; exists && pkg.SourceDir != "" {
		return pkg.SourceDir
	}
	return filepath.Join(c.SourceDir, packageName)
}

// GetTargetPath returns the full path for a target in the target directory
func (c *Config) GetTargetPath(packageName, relativePath string) string {
	// Check if we have a specific file config with a targetPath
	if pkg, exists := c.Packages[packageName]; exists {
		// Try exact file match first
		if fileConfig, exists := pkg.Files[relativePath]; exists && fileConfig.TargetPath != "" {
			return fileConfig.TargetPath
		}

		// Try regex matches
		for pattern, fileConfig := range pkg.Files {
			if fileConfig.TargetPath != "" {
				// Skip if first character is not ^ (indicating it's not meant as a regex)
				if len(pattern) == 0 || pattern[0] != '^' {
					continue
				}

				matched, err := regexp.MatchString(pattern, relativePath)
				if err == nil && matched {
					return fileConfig.TargetPath
				}
			}
		}

		// Use package-specific target directory if specified
		if pkg.TargetDir != "" {
			return filepath.Join(pkg.TargetDir, relativePath)
		}
	}

	// Fall back to global target directory
	return filepath.Join(c.TargetDir, relativePath)
}

// GetBackupPath returns the path where a file should be backed up
func (c *Config) GetBackupPath(packageName, relativePath string) string {
	// Check for package-specific backup directory
	if pkg, exists := c.Packages[packageName]; exists {
		// Check if backups are disabled for this package
		if pkg.BackupDir == "" {
			return ""
		}

		// Use package-specific backup directory if specified
		if pkg.BackupDir != "" {
			return filepath.Join(pkg.BackupDir, relativePath)
		}
	}

	// Fall back to global backup directory
	if c.BackupDir == "" {
		return ""
	}
	return filepath.Join(c.BackupDir, relativePath)
}

// ShouldForce returns whether to force overwriting for a specific package
func (c *Config) ShouldForce(packageName string) bool {
	if pkg, exists := c.Packages[packageName]; exists && pkg.Force != nil {
		return *pkg.Force
	}
	return c.Force
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

	// Other condition types would be checked here
	// For now, we're only implementing OS conditions
	// Return true for all other conditions

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
