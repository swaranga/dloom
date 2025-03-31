// Package link provides functionality for creating symlinks between source and target directories
package link

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/swaranga/dloom/internal/config"
)

// Options holds the options for link operations
type Options struct {
	// Config is the application configuration
	Config *config.Config

	// Packages is the list of package names to link
	Packages []string
}

// LinkPackages creates symlinks for all specified packages
func LinkPackages(opts Options) error {
	if len(opts.Packages) == 0 {
		return fmt.Errorf("no packages specified")
	}

	for _, pkg := range opts.Packages {
		if err := LinkPackage(pkg, opts.Config); err != nil {
			return fmt.Errorf("failed to link package %s: %w", pkg, err)
		}

		if opts.Config.Verbose {
			fmt.Printf("Successfully linked package: %s\n", pkg)
		}
	}

	return nil
}

// LinkPackage creates symlinks for a single package
func LinkPackage(pkgName string, cfg *config.Config) error {
	// Check if package has conditions and if they match
	pkgConfig := cfg.GetEffectiveConfig(pkgName, "")
	if pkgConfig.Conditions != nil && !cfg.MatchesConditions(pkgConfig.Conditions) {
		if cfg.ShouldBeVerbose(pkgName, "") {
			fmt.Printf("Skipping package %s: conditions not met\n", pkgName)
		}
		return nil
	}

	// Get the absolute path to the source package
	pkgDir := cfg.GetSourcePath(pkgName)

	// Ensure package directory exists
	pkgDirAbs, err := filepath.Abs(pkgDir)
	if err != nil {
		return err
	}

	if _, err := os.Stat(pkgDirAbs); os.IsNotExist(err) {
		return fmt.Errorf("package directory %s does not exist", pkgDirAbs)
	}

	// Walk through the package directory
	return filepath.Walk(pkgDirAbs, func(sourcePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if sourcePath == pkgDirAbs {
			return nil
		}

		// Calculate relative path from package directory
		relPath, err := filepath.Rel(pkgDirAbs, sourcePath)
		if err != nil {
			return err
		}

		// Check file-specific conditions
		fileConfig := cfg.GetEffectiveConfig(pkgName, relPath)
		if fileConfig.Conditions != nil && !cfg.MatchesConditions(fileConfig.Conditions) {
			if cfg.ShouldBeVerbose(pkgName, relPath) {
				fmt.Printf("Skipping file %s: conditions not met\n", relPath)
			}
			return nil
		}

		// Construct target path
		targetPath := cfg.GetTargetPath(pkgName, relPath)

		// If it's a directory, create it in the target directory
		if info.IsDir() {
			if cfg.IsDryRun(pkgName, relPath) {
				fmt.Printf("Would create directory: %s\n", targetPath)
				return nil
			}

			if err := os.MkdirAll(targetPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}

			if cfg.ShouldBeVerbose(pkgName, relPath) {
				fmt.Printf("Created directory: %s\n", targetPath)
			}
			return nil
		}

		// It's a file, handle symlinking
		return linkFile(sourcePath, targetPath, relPath, pkgName, cfg)
	})
}

// linkFile creates a symlink from sourcePath to targetPath
func linkFile(sourcePath, targetPath, relPath, pkgName string, cfg *config.Config) error {
	// Create parent directories if needed
	targetDir := filepath.Dir(targetPath)

	if cfg.IsDryRun(pkgName, relPath) {
		fmt.Printf("Would ensure directory exists: %s\n", targetDir)
	} else {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory %s: %w", targetDir, err)
		}
	}

	// Check if target already exists
	_, err := os.Lstat(targetPath)
	if err == nil {
		// Target exists, check if it's already the correct symlink
		linkDest, err := os.Readlink(targetPath)
		if err == nil && linkDest == sourcePath {
			// Already correctly linked
			if cfg.ShouldBeVerbose(pkgName, relPath) {
				fmt.Printf("Already linked: %s\n", relPath)
			}
			return nil
		}

		// Target exists but is not the correct symlink
		if !cfg.ShouldForce(pkgName, relPath) {
			// Ask user for confirmation before removing
			fmt.Printf("Target already exists: %s, rel-path: %s, pkg: %s. Replace? [y/N] ", targetPath, relPath, pkgName)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" {
				fmt.Println("Skipping file")
				return nil
			}
		}

		// Backup if backup directory is set
		backupPath := cfg.GetBackupPath(pkgName, relPath)
		if backupPath != "" {
			backupDir := filepath.Dir(backupPath)

			if cfg.IsDryRun(pkgName, relPath) {
				fmt.Printf("Would backup %s to %s\n", targetPath, backupPath)
			} else {
				if err := os.MkdirAll(backupDir, 0755); err != nil {
					return fmt.Errorf("failed to create backup directory %s: %w", backupDir, err)
				}

				// Copy the file to backup location
				if err := copyFile(targetPath, backupPath); err != nil {
					return fmt.Errorf("failed to backup file %s: %w", targetPath, err)
				}

				if cfg.ShouldBeVerbose(pkgName, relPath) {
					fmt.Printf("Backed up %s to %s\n", targetPath, backupPath)
				}
			}
		}

		// Remove existing target
		if cfg.IsDryRun(pkgName, relPath) {
			fmt.Printf("Would remove existing target: %s\n", targetPath)
		} else {
			if err := os.Remove(targetPath); err != nil {
				return fmt.Errorf("failed to remove existing target %s: %w", targetPath, err)
			}
		}
	} else if !os.IsNotExist(err) {
		// Some other error occurred
		return fmt.Errorf("failed to check target %s: %w", targetPath, err)
	}

	// Create symlink
	if cfg.IsDryRun(pkgName, relPath) {
		fmt.Printf("Would link: %s -> %s\n", targetPath, sourcePath)
	} else {
		if err := os.Symlink(sourcePath, targetPath); err != nil {
			return fmt.Errorf("failed to create symlink from %s to %s: %w", sourcePath, targetPath, err)
		}

		if cfg.ShouldBeVerbose(pkgName, relPath) {
			fmt.Printf("Linked: %s -> %s\n", targetPath, sourcePath)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceData, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, sourceData, 0644)
}
