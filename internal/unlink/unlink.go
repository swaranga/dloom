// Package unlink provides functionality for removing symlinks created by dloom
package unlink

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/swaranga/dloom/internal/config"
)

// Options holds the options for unlink operations
type Options struct {
	// Config is the application configuration
	Config *config.Config

	// Packages is the list of package names to unlink
	Packages []string
}

// UnlinkPackages removes symlinks for all specified packages
func UnlinkPackages(opts Options) error {
	if len(opts.Packages) == 0 {
		return fmt.Errorf("no packages specified")
	}

	for _, pkg := range opts.Packages {
		if err := UnlinkPackage(pkg, opts.Config); err != nil {
			return fmt.Errorf("failed to unlink package %s: %w", pkg, err)
		}

		if opts.Config.Verbose {
			fmt.Printf("Successfully unlinked package: %s\n", pkg)
		}
	}

	return nil
}

// UnlinkPackage removes symlinks for a single package
func UnlinkPackage(pkgName string, cfg *config.Config) error {
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

	// Track empty directories to potentially remove them later
	emptyDirs := make(map[string]bool)

	// Walk through the package directory
	err = filepath.Walk(pkgDirAbs, func(sourcePath string, info os.FileInfo, err error) error {
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

		// Construct target path
		targetPath := cfg.GetTargetPath(relPath)

		// If it's a directory, add to the map of directories to check later
		if info.IsDir() {
			emptyDirs[targetPath] = true
			return nil
		}

		// It's a file, handle unlinking
		return unlinkFile(sourcePath, targetPath, relPath, cfg)
	})

	if err != nil {
		return err
	}

	// Clean up empty directories, starting from the deepest ones
	if !cfg.DryRun {
		var dirs []string
		for dir := range emptyDirs {
			dirs = append(dirs, dir)
		}

		// Sort directories by depth (deepest first)
		sortByDepth(dirs)

		for _, dir := range dirs {
			// Check if directory is empty
			entries, err := os.ReadDir(dir)
			if err != nil {
				// Directory might not exist or other error, just skip
				continue
			}

			if len(entries) == 0 {
				if cfg.DryRun {
					fmt.Printf("Would remove empty directory: %s\n", dir)
				} else {
					if err := os.Remove(dir); err != nil {
						// Non-critical error, just log and continue
						if cfg.Verbose {
							fmt.Printf("Failed to remove directory %s: %v\n", dir, err)
						}
					} else if cfg.Verbose {
						fmt.Printf("Removed empty directory: %s\n", dir)
					}
				}
			}
		}
	}

	return nil
}

// unlinkFile removes a symlink if it points to the expected source
func unlinkFile(sourcePath, targetPath, relPath string, cfg *config.Config) error {
	// Check if target exists
	fi, err := os.Lstat(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Target doesn't exist, nothing to unlink
			if cfg.Verbose {
				fmt.Printf("No symlink found at %s\n", targetPath)
			}
			return nil
		}
		return fmt.Errorf("failed to check target %s: %w", targetPath, err)
	}

	// Check if it's a symlink
	if fi.Mode()&os.ModeSymlink == 0 {
		// Not a symlink, leave it alone
		if cfg.Verbose {
			fmt.Printf("Not a symlink: %s\n", targetPath)
		}
		return nil
	}

	// Check if the symlink points to our source file
	linkDest, err := os.Readlink(targetPath)
	if err != nil {
		return fmt.Errorf("failed to read symlink %s: %w", targetPath, err)
	}

	// Only remove if it points to our source file
	if linkDest == sourcePath {
		if cfg.DryRun {
			fmt.Printf("Would remove symlink: %s -> %s\n", targetPath, sourcePath)
		} else {
			if err := os.Remove(targetPath); err != nil {
				return fmt.Errorf("failed to remove symlink %s: %w", targetPath, err)
			}

			if cfg.Verbose {
				fmt.Printf("Removed symlink: %s\n", targetPath)
			}
		}
	} else if cfg.Verbose {
		fmt.Printf("Symlink points elsewhere, not removing: %s -> %s\n", targetPath, linkDest)
	}

	// Restore from backup if available
	if cfg.BackupDir != "" {
		backupPath := cfg.GetBackupPath(relPath)

		// Check if backup exists
		if _, err := os.Stat(backupPath); err == nil {
			if cfg.DryRun {
				fmt.Printf("Would restore from backup: %s -> %s\n", backupPath, targetPath)
			} else {
				// Copy backup back to target
				if err := copyFile(backupPath, targetPath); err != nil {
					return fmt.Errorf("failed to restore from backup %s: %w", backupPath, err)
				}

				if cfg.Verbose {
					fmt.Printf("Restored from backup: %s -> %s\n", backupPath, targetPath)
				}

				// Remove the backup
				if err := os.Remove(backupPath); err != nil {
					fmt.Printf("Warning: Failed to remove backup %s: %v\n", backupPath, err)
				}
			}
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

// sortByDepth sorts paths by depth (deepest first)
func sortByDepth(paths []string) {
	sort.Slice(paths, func(i, j int) bool {
		// Count separators as a proxy for depth
		depthI := strings.Count(paths[i], string(os.PathSeparator))
		depthJ := strings.Count(paths[j], string(os.PathSeparator))
		return depthI > depthJ
	})
}
