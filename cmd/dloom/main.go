package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("dloom - Dotfile manager and system bootstrapper")

	if len(os.Args) < 2 {
		fmt.Println("Usage: dloom [command]")
		fmt.Println("Commands:")
		fmt.Println("  link     - Create symlinks for dotfiles")
		fmt.Println("  unlink   - Remove symlinks for dotfiles")
		fmt.Println("  setup    - Run system setup scripts")
		os.Exit(1)
	}

	// Basic command handling
	command := os.Args[1]
	switch command {
	case "link":
		fmt.Println("Link command not implemented yet")
	case "unlink":
		fmt.Println("Unlink command not implemented yet")
	case "setup":
		fmt.Println("Setup command not implemented yet")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
