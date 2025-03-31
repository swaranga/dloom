# dloom

A lightweight, flexible dotfile manager and system bootstrapper for macOS and Linux.

![License](https://img.shields.io/badge/license-MIT-blue.svg)

## Overview

`dloom` is a CLI tool that links and unlinks your configuration files to your development machine. It manages symlinks between your dotfiles repository and your home directory, while also providing system bootstrapping capabilities. The tool is inspired from GNU Stow and other dotfile managers, but differs in its approach by focusing on file-level symlinks rather than directory-level symlinks. This allows for the creation of symlinks for individual files, enabling other applications to add files to the same directories without them being tracked in your dotfiles repository.

## Features

- **Symlink Management**: Create and manage symlinks for your dotfiles with ease.
- **File-Level Symlinks**: Links individual files (not directories), allowing other applications to add files to the same directories without them being tracked in your dotfiles repo.
  - This is the main difference from GNU Stow. _It does mean that addition of a file to a directory in your dotfiles repository will not automatically create a symlink for it. You will need to run `dloom link` again to create the symlink for the new file._
- **Conditional Linking**: Link files only when specific conditions are met (OS, distro, installed tools, tool versions).
- **Hierarchical Configuration**: Override settings at global, package, or file level including support for regex patterns.
- **Backup System**: Automatically back up existing files before replacing them.
- **Dry Run Mode**: Preview changes without modifying your system.
- **Cross-Platform**: Works consistently across macOS and Linux.
  - Windows support is not planned, but contributions are welcome.

## Installation

### From Source

**Requirements:**
- Go 1.18 or later

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

### How Symlinks Work

Consider this example dotfiles repository:

```
~/dotfiles/
├── vim/
│   ├── vimrc
│   └── config/
│       └── plugins.vim
├── bash/
│   ├── bashrc
│   └── bash_profile
└── tmux/
    └── tmux.conf
```

When you run `dloom link vim`, it will create:

```
~/                                ~/dotfiles/
├── .vimrc ----------------→      vim/vimrc
└── .config/                      └── config/
    └── plugins.vim --------→         └── plugins.vim
```

Notice that:
- Only files get symlinked, not directories
- The directory structure is mirrored in your home directory
- Files in the same directories from other sources remain untouched

Different commands and their effects:

```bash
# Link vim package to home directory
dloom link vim

# Link vim package to a different target directory
dloom -t ~/.config/nvim link vim
# Creates: ~/.config/nvim/vimrc → ~/dotfiles/vim/vimrc
#          ~/.config/nvim/config/plugins.vim → ~/dotfiles/vim/config/plugins.vim

# Link from a different source directory
dloom -s /path/to/dotfiles link vim
# Uses: /path/to/dotfiles/vim/ as the source

# Dry run to preview changes
dloom -d link vim
# Output:
# Would create directory: /home/user/.config
# Would link: /home/user/.vimrc → /home/user/dotfiles/vim/vimrc
# Would link: /home/user/.config/plugins.vim → /home/user/dotfiles/vim/.config/plugins.vim
```

## Configuration (Optional)

`dloom` can be (optionally) configured via a YAML file. By default, it looks for:
1. `./dloom/config.yaml` (in current directory)
2. `~/.config/dloom/config.yaml` (in user config directory)

Or specify a custom location with `-c path/to/config.yaml`. For easiest configuration, create a `dloom/config.yaml` file in the root of your dotfiles repository.

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

### Full Configuration

For a complete example, check the `examples/` directory in the repository. It contains various configurations for different setups.

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
└── examples/           # Sample configurations
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

- Inspired by [GNU Stow](https://www.gnu.org/software/stow/) and other dotfile managers
- Built with Go

---

*dloom - Weave your digital environment together*