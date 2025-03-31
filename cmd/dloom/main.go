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
	"strings"
)

// Command-line flags
var (
	configPath  string
	force       bool
	verbose     bool
	dryRun      bool
	sourceDir   string
	targetDir   string
	packageArgs stringSliceFlag // Custom type to handle multiple package args
)

// stringSliceFlag is a custom flag type that allows for multiple values
type stringSliceFlag []string

func (s *stringSliceFlag) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSliceFlag) Set(value string) error {
	// Handle comma-separated values
	if strings.Contains(value, ",") {
		*s = append(*s, strings.Split(value, ",")...)
	} else {
		*s = append(*s, value)
	}
	return nil
}

func init() {
	// Define command-line flags
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&configPath, "c", "", "Path to config file (shorthand)")

	flag.BoolVar(&force, "force", false, "Force overwriting existing files")
	flag.BoolVar(&force, "f", false, "Force overwriting existing files (shorthand)")

	flag.BoolVar(&verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&verbose, "v", false, "Enable verbose output (shorthand)")

	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	flag.BoolVar(&dryRun, "d", false, "Show what would be done without making changes (shorthand)")
	flag.BoolVar(&dryRun, "n", false, "Show what would be done without making changes (shorthand)")

	flag.StringVar(&sourceDir, "source", "", "Source directory (defaults to current directory)")
	flag.StringVar(&sourceDir, "src", "", "Source directory (shorthand)")
	flag.StringVar(&sourceDir, "s", "", "Source directory (shorthand)")

	flag.StringVar(&targetDir, "target", "", "Target directory (defaults to home directory)")
	flag.StringVar(&targetDir, "dest", "", "Target directory (alias)")
	flag.StringVar(&targetDir, "t", "", "Target directory (shorthand)")

	// Add package flag that can be specified multiple times
	flag.Var(&packageArgs, "package", "Package name (can be used multiple times or as comma-separated list)")
	flag.Var(&packageArgs, "p", "Package name (shorthand)")
}

func main() {
	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "dloom - Dotfile manager and system bootstrapper\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  dloom [options] <command> [packages...]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  link      Create symlinks for specified packages\n")
		fmt.Fprintf(os.Stderr, "  unlink    Remove symlinks for specified packages\n")
		fmt.Fprintf(os.Stderr, "  setup     Run specified setup scripts\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  dloom link vim bash        # Link vim and bash packages\n")
		fmt.Fprintf(os.Stderr, "  dloom -p vim,bash link     # Same as above using -p flag\n")
		fmt.Fprintf(os.Stderr, "  dloom -v -d link vim       # Verbose dry-run for vim package\n")
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

	// If packages were specified via -p/--package flags, use those instead of command arguments
	// but only for link and unlink commands
	if len(packageArgs) > 0 && (command == "link" || command == "unlink") {
		cmdArgs = packageArgs
	}

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
		return fmt.Errorf("no packages specified for link command\n" +
			"Use: dloom link <package>... or dloom -p <package>[,<package>...] link")
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
		return fmt.Errorf("no packages specified for unlink command\n" +
			"Use: dloom unlink <package>... or dloom -p <package>[,<package>...] unlink")
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
