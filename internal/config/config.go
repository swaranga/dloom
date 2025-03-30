// Package config provides configuration handling for dloom
package config

import (
	"fmt"
	"os"
	"path/filepath"

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
	}
}

// LoadConfig loads configuration from the specified file
// If the file doesn't exist, returns default config
func LoadConfig(configPath string) (*Config, error) {
	config := DefaultConfig()

	// If no config path specified, look in default locations
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			// Try ~/.config/dloom/config.yaml
			configPath = filepath.Join(homeDir, ".config", "dloom", "config.yaml")
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Try ~/.dloom.yaml
				configPath = filepath.Join(homeDir, ".dloom.yaml")
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					// No config file found, use defaults
					return config, nil
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
	return filepath.Join(c.SourceDir, packageName)
}

// GetTargetPath returns the full path for a target in the target directory
func (c *Config) GetTargetPath(relativePath string) string {
	return filepath.Join(c.TargetDir, relativePath)
}

// GetBackupPath returns the path where a file should be backed up
func (c *Config) GetBackupPath(relativePath string) string {
	if c.BackupDir == "" {
		return ""
	}
	return filepath.Join(c.BackupDir, relativePath)
}
