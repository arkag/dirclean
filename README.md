# Boiling the ocean to save on storage

# dirclean

`dirclean` is a command-line utility written in Go that helps you clean up old files from your system. It's designed to be simple, efficient, and customizable.

## Features

- Removes files older than a specified number of days.
- Supports multiple directories and wildcard patterns.
- Dry run mode to preview files that would be deleted.
- Logging to track the cleanup process.
- Configurable via a YAML file.
- No external dependencies.

## Installation

1. Make sure you have Go installed on your system.
2. Copy the code and save it as `dirclean.go`.
3. Open a terminal or command prompt in the directory where you saved the file.
4. Run the command `go build dirclean.go` to compile the code.
5. This will create an executable file named `dirclean` in the same directory.

## Usage

```
dirclean [--help|-h]
```

### Options

- `--help`, `-h`: Show the help message.

### Environment variables

- `DRY_RUN`: When set to `true` or `1`, the program will only show which files would be deleted without actually deleting them. Defaults to `true`.
- `CONFIG_FILE`: Path to the YAML configuration file. Defaults to `/etc/dirclean/dirclean.yaml`.
- `LOG_FILE`: Path to the log file. Defaults to `/var/log/dirclean.log`.
- `LOG_LEVEL`: Logging level (DEBUG, INFO, WARN, ERROR, FATAL). Defaults to `INFO`.

## Configuration

The program is configured using a YAML file. The default configuration file is `/etc/dirclean/dirclean.yaml`.

### Configuration format

YAML

```
- delete_older_than_days: 30
  paths:
    - /path/to/directory1
    - /path/to/directory2/*
    - /path/to/directory3/**/*.log
```

- `delete_older_than_days`: The number of days after which files are considered old and will be deleted.
- `paths`: A list of paths to be processed. You can use wildcard patterns (`*`) to match multiple directories or files.

## Examples

- To run the program with the default configuration:

- To run the program in dry run mode:

Bash

```
DRY_RUN=true./dirclean
```

- To run the program with a custom configuration file:

Bash

```
CONFIG_FILE=/path/to/config.yaml./dirclean
```

## Contributing

Contributions are welcome! Please feel free to submit bug reports, feature requests, or pull requests.

## License

This program is licensed under the GNU General Public License version 3 (GPLv3). See LICENSE for the full license text.