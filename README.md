# dloom

A lightweight, flexible dotfile manager and system bootstrapper for macOS and Linux.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Overview

`dloom` is a modern CLI tool that weaves your configuration files into a cohesive environment across systems. It manages symlinks between your dotfiles repository and your home directory, while also providing system bootstrapping capabilities.

## Features

- **Smart Symlink Management**: Create and manage symlinks for your dotfiles with ease
- **Conditional Linking**: Link files only when specific conditions are met (OS, distro, installed tools)
- **Hierarchical Configuration**: Override settings at global, package, or file level
- **Backup System**: Automatically back up existing files before replacing them
- **Dry Run Mode**: Preview changes without modifying your system
- **Cross-Platform**: Works consistently across macOS and Linux

## Installation

```bash
# Install from source
go install github.com/swaranga/dloom/cmd/dloom@latest

# Or clone and build
git clone https://github.com/swaranga/dloom.git
cd dloom
go build -o build/dloom ./cmd/dloom
```

## Quick Start

```bash
# Link all dotfiles from your vim package
dloom link vim

# Link multiple packages
dloom link vim tmux bash

# Link with verbose output
dloom -v link vim

# Preview changes without making them
dloom -d link vim

# Unlink a package
dloom unlink vim
```

## Configuration

`dloom` can be configured via a YAML file. By default, it looks for:
1. `./dloom/config.yaml` (in current directory)
2. `~/.config/dloom/config.yaml` (in user config directory)

Or specify a custom location with `-c path/to/config.yaml`.

### Basic Configuration

```yaml
# Global settings
sourceDir: "~/dotfiles"     # Where your dotfiles are stored
targetDir: "~"              # Where to create symlinks
backupDir: "~/.dloom/backups"  # Where to back up existing files
verbose: true               # Enable detailed output
force: false                # Don't overwrite without asking
dryRun: false               # Actually make changes

# Package-specific settings
packages:
  vim:
    targetDir: "~/.config/nvim"  # Override target for vim package
    conditions:
      os:
        - "linux"
        - "darwin"  # Only link on Linux or macOS
```

### Advanced Configuration

```yaml
packages:
  tmux:
    conditions:
      executable:
        - "tmux"  # Only link if tmux is installed
    
    # File-specific configurations
    files:
      # Regular file
      "tmux.conf": {}
      
      # File with regex pattern matching
      "regex:^tmux.*\.local$":
        conditions:
          os:
            - "darwin"  # Only link on macOS
      
      # Version-specific configurations
      "tmux.new.conf":
        conditions:
          executable_version:
            "tmux": ">=3.0"  # Only link for tmux 3.0+
```

## Usage

### Linking Dotfiles

```bash
# Basic linking
dloom link <package>...

# Link with options
dloom -v -f link <package>...  # Verbose and force overwrite

# Specify packages with -p flag
dloom -p vim,tmux link
```

### Unlinking Dotfiles

```bash
# Remove symlinks
dloom unlink <package>...

# Unlink with options
dloom -d unlink <package>...  # Dry run (preview only)
```

### Command-line Options

| Option | Description |
|--------|-------------|
| `-c, --config` | Path to config file |
| `-f, --force` | Force overwrite existing files |
| `-v, --verbose` | Enable verbose output |
| `-d, -n, --dry-run` | Show what would happen without making changes |
| `-s, --source, --src` | Source directory |
| `-t, --target, --dest` | Target directory |
| `-p, --package` | Package name(s) to process |

## Conditional Linking

`dloom` supports conditional linking based on:

- **Operating System**: Link files only on specific OS
- **Linux Distribution**: Link files only on specific distros
- **Executable Presence**: Link files only if certain executables exist
- **Executable Version**: Link files only if executables meet version requirements

## Project Structure

```
dloom/
├── cmd/dloom/          # Command-line interface
├── internal/           # Internal implementation
│   ├── config/         # Configuration handling
│   ├── link/           # Link implementation
│   ├── unlink/         # Unlink implementation
│   └── setup/          # System setup implementation
└── pkg/                # Public packages
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Inspired by GNU Stow and other dotfile managers
- Built with Go

---

*dloom - Weave your digital environment together*