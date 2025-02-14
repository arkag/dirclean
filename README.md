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
- **Dry Run Mode**: Preview which files would be deleted without actually removing them
- **Logging**: Detailed logging for debugging and monitoring
- **Cross-Platform**: Built with Go, runs on Linux, macOS, and Windows
- **Auto-Update**: Built-in mechanism to update to the latest version
- **Multiple Package Formats**: Available as `.deb`, `.rpm`, and Arch Linux packages

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

### From Source
```bash
go install github.com/arkag/dirclean@latest
```

---

## Configuration

The program uses a YAML configuration file (`config.yaml`) to define the directories and retention policies. Here's an example configuration:

```yaml
- delete_older_than_days: 30
  paths:
    - /foo_dir/foo_sub_dir
    - /foo_dir/foo_sub_dir/foo_wildcard_dir*
  mode: dry-run
```

### Configuration Options

- **`delete_older_than_days`**: Number of days after which files are considered old and eligible for deletion
- **`paths`**: List of directories to clean. Supports wildcards (`*`) for matching multiple directories
- **`mode`**: Operation mode (default: `dry-run`)
  - `analyze`: Only report files that would be deleted
  - `dry-run`: List files that would be deleted without actually removing them
  - `interactive`: Prompt for confirmation before deleting each file
  - `scheduled`: Delete files automatically without confirmation

---

## Usage

### Basic Usage
```bash
dirclean [flags]
```

### Flags
- `--config`: Path to config file (default: `config.yaml`)
- `--dry-run`: Run in dry-run mode (default: true)
- `--update`: Update to the latest version
- `--version`: Show version information

### Environment Variables

- `CONFIG_FILE`: Path to the YAML configuration file (default: `config.yaml`)
- `LOG_FILE`: Path to the log file (default: `dirclean.log`)
- `LOG_LEVEL`: Logging level (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`) (default: `INFO`)
- `DRY_RUN`: Enable dry-run mode (default: `true`)

Example:
```bash
CONFIG_FILE=/etc/dirclean/config.yaml LOG_LEVEL=DEBUG dirclean
```

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
