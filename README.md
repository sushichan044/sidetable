# sidetable

**sidetable** is a Go CLI tool that helps manage a project-local private area (e.g. `.$USER/`) and integrates external commands.

You define subcommands in a config file and can execute arbitrary commands as `sidetable <subcmd> ...`. This lets you standardize project-specific directory layouts and tool settings, and seamlessly integrate external tools (e.g. [x-motemen/ghq](https://github.com/x-motemen/ghq)).

## Features

- **Command integration**: Run subcommands defined in the config as external commands
- **Template expansion**: Dynamic generation of commands, args, and env vars with Go templates
- **Aliases**: Optional short aliases for subcommands
- **Project context**: Auto-resolve project and private directory paths
- **Transparent execution**: Fully preserves stdout/stderr and exit codes

## Installation

### Using Homebrew (macOS/Linux)

```bash
brew install --cask sushichan044/tap/sidetable
```

### Using [mise](https://mise.jdx.dev/)

  ```bash
mise install github:sushichan044/sidetable
```

### Using `go install`

```bash
go install github.com/sushichan044/sidetable/cmd/sidetable@latest
```

### Download binary

You can download the latest release binaries from [Releases](https://github.com/sushichan044/sidetable/releases).

## Configuration

### Location

The config file is searched in the following order:

1. If `SIDETABLE_CONFIG_DIR` is set: `$SIDETABLE_CONFIG_DIR/config.yml`
2. Otherwise: `$XDG_CONFIG_HOME/sidetable/config.yml` (or `~/.config/sidetable/config.yml` if `XDG_CONFIG_HOME` is not set)

### Basic example

```yaml
# Required. Sets the project-local private area name.
# If sets to ".sidetable", the private area path is "./.sidetable"
directory: ".sidetable"

commands:
  ghq:
    # Required. Command name to execute.
    command: "ghq"
    # Optional. Arguments to pass to the command.
    args:
      # Arguments added before user-provided args.
      prepend:
        - "get"
        - "-u"
    env:
      # You can use `{{.CommandDir}}` to get the command-local directory path.
      # In this example, GHQ_ROOT is set to "./.sidetable/ghq"
      GHQ_ROOT: "{{.CommandDir}}"
    # Optional. Description shown in `sidetable list`.
    description: "ghq wrapper with project-local root"
    # Optional. Short alias for the command.
    alias: "gg"

  note:
    # You can use `{{.ConfigDir}}` to get the configuration directory path.
    # You can set your own shell scripts as command.
    command: "{{.ConfigDir}}/vim-note.sh"
    args:
      # Optional. Arguments added after user-provided args.
      append:
        - "{{.CommandDir}}/note.md"
    description: "Open project note file"
    alias: "n"

```

### Format

#### Top-level

- `directory` (**required**): directory name of the project-local private area (relative path)
  - Example: `.sushichan044`
  - Resolved as a path relative to the current directory

#### `commands.<name>`

Fields for each command:

- `command` (**required**): command name to execute (Go template allowed)
- `args` (optional): argument injection settings (Go template allowed)
  - `args.prepend`: list of arguments added before user-provided args
  - `args.append`: list of arguments added after user-provided args
  - If omitted, user-provided arguments are passed through as-is
- `env` (optional): additional or overriding environment variables (key-value map)
- `description` (optional): description shown in `sidetable list`
- `alias` (optional): a short alias for the command (up to one)

### Template variables

Each string in `command`, `args`, `env`, and `description` is evaluated as a Go template. Available variables:

| Variable      | Description                                                |
| ------------- | ---------------------------------------------------------- |
| `.ProjectDir` | current directory when running `sidetable` (absolute path) |
| `.PrivateDir` | `.ProjectDir/<directory>` (absolute path)                  |
| `.CommandDir` | `.PrivateDir/<commandName>` (absolute path)                |
| `.ConfigDir`  | directory containing `config.yml` (absolute path)          |

#### Argument injection rules

`args.prepend` and `args.append` are evaluated as templates, then combined as:

```
prepend + userArgs + append
```

Example:

```yaml
commands:
  example:
    command: "mycommand"
    args:
      prepend:
        - "--flag"
      append:
        - "--output=result.txt"
```

```bash
$ sidetable example arg1 arg2
# Executed command: mycommand --flag arg1 arg2 --output=result.txt
```

### Validation

The config file is validated with these rules:

- `directory` is required (absolute paths are not allowed; must be relative)
- `commands` is required (at least one command)
- Each command must have a `command` field
- `command` must not contain spaces, tabs, or newlines
- `alias` must be unique
- `alias` must not conflict with existing command names

## Usage

### Basic usage

```bash
# List commands
$ sidetable list
ghq (q) ghq wrapper with project-local root
note Open project note

# Run a command by alias
$ sidetable q get https://github.com/example/repo

# Run a command by full name
$ sidetable ghq list

# Show help
$ sidetable --help
$ sidetable help

# Show version
$ sidetable --version
$ sidetable version
```

### Example: integrate with ghq

Example configuration for project-local Git repository management:

```yaml
directory: ".local"

commands:
  ghq:
    command: "ghq"
    args:
      prepend:
        - "get"
    env:
      GHQ_ROOT: "{{.CommandDir}}"
    description: "Manage repositories in project-local directory"
    alias: "q"
```

Example:

```bash
$ cd ~/myproject
$ sidetable q get https://github.com/example/repo
# => cloned into ~/myproject/.local/ghq/github.com/example/repo

$ sidetable q list
github.com/example/repo
```

### Example: integrate with direnv

Example configuration to manage project-local environment variables:

```yaml
directory: ".private"

commands:
  env:
    command: "direnv"
    args:
      prepend:
        - "allow"
    env:
      DIRENV_CONFIG: "{{.CommandDir}}"
    description: "Manage project-local environment variables"
```

## Commands

### Built-in commands

| Command   | Description                                          |
| --------- | ---------------------------------------------------- |
| `list`    | Show the list of commands defined in the config file |
| `version` | Show version info                                    |
| `help`    | Show help message                                    |

### Custom commands

Commands defined in `config.yml` are added dynamically. You can run them by command name or alias.

```bash
sidetable <command-name> [args...]
sidetable <alias> [args...]
```

## Development

### Requirements

- Go 1.23+
- [mise](https://mise.jdx.dev/) (task runner, optional)

### Quick commands

```bash
# Run tests
mise run test

# Format code
mise run fmt

# Run lint
mise run lint

# Fix lint issues
mise run lint-fix

# Build cross-platform snapshot binaries
mise run build-snapshot

# Remove generated files
mise run clean
```

### Standard Go commands

```bash
# Run in development mode
go run ./cmd/sidetable

# Run all tests
go test ./...

# Tidy dependencies
go mod tidy

# Tests with coverage
go test -cover ./...
```

### Project structure

```
sidetable/
├── cmd/sidetable/          # CLI entry point
│   └── main.go
├── pkg/sidetable/          # Public API
│   └── project.go
├── internal/               # Internal packages
│   ├── config/            # Config loading
│   ├── action/            # Command execution
│   └── xdg/               # XDG Base Directory support
├── docs/ai/               # Specification
│   └── SPEC.md
├── go.mod
├── go.sum
├── mise.toml              # Task runner config
├── .golangci.yml          # Lint config
└── .goreleaser.yaml       # Release config
```

## License

MIT License

## Contributing

Please open an issue for bug reports or feature requests: [Issues](https://github.com/sushichan044/sidetable/issues)
