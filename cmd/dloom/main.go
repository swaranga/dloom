// Package main provides the entry point for the dloom command
package main

import (
	"flag"
	"fmt"
	"github.com/swaranga/dloom/internal/config"
	"github.com/swaranga/dloom/internal/link"
	"github.com/swaranga/dloom/internal/setup"
	"github.com/swaranga/dloom/internal/unlink"
	"os"
	"path/filepath"
)

// Command-line flags
var (
	configPath string
	force      bool
	verbose    bool
	dryRun     bool
	sourceDir  string
	targetDir  string
)

func init() {
	// Define command-line flags
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.BoolVar(&force, "force", false, "Force overwriting existing files")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	flag.StringVar(&sourceDir, "source", "", "Source directory (defaults to current directory)")
	flag.StringVar(&targetDir, "target", "", "Target directory (defaults to home directory)")

	// Add short-form flags
	flag.BoolVar(&force, "f", false, "Force overwriting existing files (shorthand)")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")
	flag.BoolVar(&dryRun, "n", false, "Show what would be done without making changes (shorthand)")
}

func main() {
	// Custom usage message
	flag.Usage = func() {
		_, err := fmt.Fprintf(os.Stderr, "dloom - Dotfile manager and system bootstrapper\n\n")
		if err != nil {
			return
		}
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  dloom [options] command [arguments]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  link <package>...    Create symlinks for specified packages\n")
		fmt.Fprintf(os.Stderr, "  unlink <package>...  Remove symlinks for specified packages\n")
		fmt.Fprintf(os.Stderr, "  setup <script>...    Run specified setup scripts\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}

	// Parse flags
	flag.Parse()

	// Need at least one argument (the command)
	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Load config
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override config with command-line flags
	if force {
		cfg.Force = true
	}
	if verbose {
		cfg.Verbose = true
	}
	if dryRun {
		cfg.DryRun = true
	}
	if sourceDir != "" {
		cfg.SourceDir = sourceDir
	}
	if targetDir != "" {
		cfg.TargetDir = targetDir
	}

	// Get absolute path for source directory
	if !filepath.IsAbs(cfg.SourceDir) {
		absSourceDir, err := filepath.Abs(cfg.SourceDir)
		if err == nil {
			cfg.SourceDir = absSourceDir
		}
	}

	// Handle command
	command := args[0]
	cmdArgs := args[1:]

	var cmdErr error
	switch command {
	case "link":
		cmdErr = handleLink(cmdArgs, cfg)
	case "unlink":
		cmdErr = handleUnlink(cmdArgs, cfg)
	case "setup":
		cmdErr = handleSetup(cmdArgs, cfg)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}

	if cmdErr != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr)
		os.Exit(1)
	}
}

// handleLink handles the "link" command
func handleLink(args []string, cfg *config.Config) error {
	if len(args) == 0 {
		return fmt.Errorf("no packages specified for link command")
	}

	opts := link.Options{
		Config:   cfg,
		Packages: args,
	}

	return link.LinkPackages(opts)
}

// handleUnlink handles the "unlink" command
func handleUnlink(args []string, cfg *config.Config) error {
	if len(args) == 0 {
		return fmt.Errorf("no packages specified for unlink command")
	}

	opts := unlink.Options{
		Config:   cfg,
		Packages: args,
	}

	return unlink.UnlinkPackages(opts)
}

// handleSetup handles the "setup" command
func handleSetup(args []string, cfg *config.Config) error {
	if len(args) == 0 {
		return fmt.Errorf("no scripts specified for setup command")
	}

	opts := setup.Options{
		Config:  cfg,
		Scripts: args,
	}

	return setup.RunScripts(opts)
}
