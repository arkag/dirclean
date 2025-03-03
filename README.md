# DirClean

DirClean is a Go program designed to clean up old files from directories based on a YAML configuration file. It removes files older than a specified number of days from directories defined in the configuration.

---

## Attribution -- AKA Boiling the Ocean

This project was developed with the assistance of Augment, an AI language model based on Claude. Contributions include:
- Writing and debugging the Go program
- Implementing the self-update mechanism
- Creating the GitHub Actions workflow for multi-platform builds and packaging
- Generating documentation

---

## Features

- **Configurable**: Define directories and file retention policies in a YAML configuration file
- **Multiple Operation Modes**: Support for analyze, dry-run, interactive, and scheduled modes
- **Logging**: Detailed logging for debugging and monitoring
- **Cross-Platform**: Built with Go, runs on Linux, macOS, and Windows
- **Auto-Update**: Built-in mechanism to update to the latest version

---

## Installation

### Quick Install (Linux and macOS)

Install the latest version with a single command:

```bash
curl -sSL https://raw.githubusercontent.com/arkag/dirclean/main/install.sh | bash
```

Or if you prefer to inspect the script first:

```bash
curl -O https://raw.githubusercontent.com/arkag/dirclean/main/install.sh
chmod +x install.sh
./install.sh
```

The script will:
- Detect your OS and architecture
- Download the appropriate binary
- Verify the checksum
- Install to `/usr/local/bin` (if run as root) or `~/.local/bin` (if run as user)

### From Binary Release

1. Download the latest release for your platform from the [Releases page](https://github.com/arkag/dirclean/releases)
2. Extract the archive:
   ```bash
   tar xzf dirclean-<os>-<arch>.tar.gz
   ```
3. Move the binary to your PATH:
   ```bash
   sudo mv dirclean /usr/local/bin/
   ```

### From Package Manager

#### Debian/Ubuntu
```bash
curl -LO https://github.com/arkag/dirclean/releases/latest/download/dirclean.deb
sudo dpkg -i dirclean.deb
```

#### RHEL/Fedora
```bash
curl -LO https://github.com/arkag/dirclean/releases/latest/download/dirclean.rpm
sudo rpm -i dirclean.rpm
```

#### Arch Linux
```bash
curl -LO https://github.com/arkag/dirclean/releases/latest/download/dirclean.pkg.tar.zst
sudo pacman -U dirclean.pkg.tar.zst
```

#### Homebrew (macOS and Linux)
```bash
# Add the tap
brew tap arkag/dirclean

# Install dirclean
brew install dirclean
```

To upgrade to the latest version:
```bash
brew upgrade dirclean
```

### From Source
```bash
go install github.com/arkag/dirclean@latest
```

---

## Configuration

The program uses a YAML configuration file to define the directories and retention policies. An example configuration file is installed at:

- Linux: `/usr/share/dirclean/example.config.yaml`
- macOS: `/usr/local/share/dirclean/example.config.yaml`
- Windows: `C:\ProgramData\dirclean\example.config.yaml`

### Example Configuration
```yaml
defaults:
  older_than_days: 30
  mode: dry-run # All operations default to dry-run for safety
  log_level: INFO
  log_file: dirclean.log
  clean_broken_symlinks: false # Default to false for safety

rules:
  # Example 1: Minimal configuration with only required paths and mode
  - paths:
      - /var/lib/**/*
    mode: analyze

  # Example 2: Override some defaults
  - paths:
      - ~/Downloads
      - ~/Documents/temp
    older_than_days: 7
    clean_broken_symlinks: true

  # Example 3: Another minimal configuration
  - paths:
      - /tmp/*
      - /var/tmp/*
    older_than_days: 1
    mode: interactive
```

### Configuration Options

- **`older_than_days`**: Number of days after which files are considered old and eligible for deletion
- **`paths`**: List of directories to clean. Supports wildcards (`*`) for matching multiple directories
- **`mode`**: Operation mode
  - `analyze`: Only report files that would be deleted
  - `dry-run`: List files that would be deleted without actually removing them
  - `interactive`: Prompt for confirmation before deleting each file
  - `scheduled`: Delete files automatically without confirmation
- **`clean_broken_symlinks`**: Boolean flag to enable cleaning of broken symbolic links (default: `false`)
- **`log_level`**: Logging level (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`)
- **`log_file`**: Path to log file

---

## Usage

### Basic Usage
```bash
dirclean [flags]
```

### Flags
- `--config`: Path to config file (default: `config.yaml`)
- `--mode`: Only process paths configured with this mode (`analyze`, `dry-run`, `interactive`, `scheduled`)
- `--log`: Path to log file
- `--log-level`: Logging level (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`)
- `--update`: Update to the latest version
- `--version`: Show version information
- `--tag`: Version tag for update (default: `latest`)

Example:
```bash
# Process only paths configured with interactive mode
dirclean --mode interactive

# Use a specific config file
dirclean --config /etc/dirclean/config.yaml

# Update to the latest version
dirclean --update
```

Note: By default, all operations run in `dry-run` mode for safety. Use the `--mode` flag to change this behavior.

---

## Auto-Update

To update to the latest version:
```bash
dirclean --update
```

---

## GitHub Actions Pipeline

This repository includes a GitHub Actions workflow that:
- Builds binaries for multiple platforms (Linux, macOS, Windows)
- Creates distribution packages (DEB, RPM, Arch)
- Runs tests and quality checks
- Creates GitHub releases with all artifacts
- Generates SHA256 checksums for verification

---

## Logging

The program logs its activities to a file (`dirclean.log` by default). Logs include:
- Files processed and their ages
- Deletion operations and results
- Error conditions and warnings
- Update operations

---

## Contributing

Contributions are welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Security and Code Quality

This project employs multiple security and code quality scanning tools:

- **CodeQL**: GitHub's semantic code analysis engine
- **gosec**: Go security checker
- **staticcheck**: Advanced Go linter
- **govulncheck**: Go vulnerability checker
- **nancy**: Dependency vulnerability scanner
- **go vet**: Go source code static analysis

Security scans run:
- On every push to main
- On every pull request
- Weekly scheduled scans

You can view scan results in the Security tab of the GitHub repository.

For local security scanning:

```bash
# Install security tools
go install github.com/securego/gosec/v2/cmd/gosec@latest
go install honnef.co/go/staticcheck/cmd/staticcheck@latest
go install golang.org/x/vuln/cmd/govulncheck@latest
go install github.com/sonatype-nexus-community/nancy@latest

# Run security checks
gosec ./...
staticcheck ./...
govulncheck ./...
go list -json -deps ./... | nancy sleuth
```
