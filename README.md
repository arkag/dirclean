# DirClean

DirClean is a Go program designed to clean up old files from directories based on a YAML configuration file. It removes files older than a specified number of days from directories defined in the configuration.

---

## Attribution

This project was developed with the assistance of **DeepSeek-V3**, an AI language model created by DeepSeek. Contributions include:
- Writing and debugging the Go program (`dirclean.go`).
- Creating the GitHub Actions workflow (`build.yml`).
- Generating the `README.md` file.

## Features

- **Configurable**: Define directories and file retention policies in a YAML configuration file.
- **Dry Run Mode**: Preview which files would be deleted without actually removing them.
- **Logging**: Detailed logging for debugging and monitoring.
- **Cross-Platform**: Built with Go, it can be compiled and run on multiple platforms.

---

## Configuration

The program uses a YAML configuration file (`config.yaml`) to define the directories and retention policies. Here's an example configuration:

```yaml
- delete_older_than_days: 30
  paths:
    - /foo_dir/foo_sub_dir
    - /foo_dir/foo_sub_dir/foo_wildcard_dir*
```

### Configuration Options

- **`delete_older_than_days`**: Number of days after which files are considered old and eligible for deletion.
- **`paths`**: List of directories to clean. Supports wildcards (`*`) for matching multiple directories.

---

## Usage

### Build the Program

To build the program, run:

```bash
go build -o dirclean ./dirclean.go
```

This will generate a binary named `dirclean`.

### Run the Program

Run the program with the following command:

```bash
./dirclean
```

#### Environment Variables

You can customize the program's behavior using the following environment variables:

- **`DRY_RUN`**: Set to `true` to enable dry run mode (default: `true`).
- **`CONFIG_FILE`**: Path to the YAML configuration file (default: `config.yaml`).
- **`LOG_FILE`**: Path to the log file (default: `dirclean.log`).
- **`LOG_LEVEL`**: Logging level (`DEBUG`, `INFO`, `WARN`, `ERROR`, `FATAL`) (default: `INFO`).

Example:

```bash
DRY_RUN=false ./dirclean
```

---

## GitHub Actions Pipeline

This repository includes a GitHub Actions workflow that automatically builds the program whenever changes are pushed to the `main` branch or a pull request is opened.

### Workflow Details

- **Trigger**: Runs on `push` to `main` and `pull_request` targeting `main`.
- **Steps**:
  1. Checks out the repository.
  2. Sets up Go.
  3. Builds the program.
  4. Uploads the compiled binary as a build artifact.

### Accessing Build Artifacts

After the workflow runs, you can download the compiled binary (`dirclean`) from the **Artifacts** section of the workflow run.

---

## Logging

The program logs its activities to a file (`dirclean.log` by default). Logs include information about files processed, deleted, and any errors encountered.

---

## Contributing

Contributions are welcome! If you find a bug or have a feature request, please open an issue or submit a pull request.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.