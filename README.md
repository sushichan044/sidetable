# sidetable

**sidetable** is a Go CLI tool that helps manage a project-local private area (e.g. `.$USER/`) and integrates external commands.

You define subcommands in a config file and can execute arbitrary commands as `sidetable <subcmd> ...`. This lets you standardize project-specific directory layouts and tool settings, and seamlessly integrate external tools (e.g. [x-motemen/ghq](https://github.com/x-motemen/ghq)).

<!-- TOC -->

- [sidetable](#sidetable)
  - [Installation](#installation)
    - [Using Homebrew (macOS/Linux)](#using-homebrew-macoslinux)
    - [Using mise](#using-mise)
    - [Using `go install`](#using-go-install)
    - [Download binary](#download-binary)
  - [Usage](#usage)
    - [Basic usage](#basic-usage)
    - [Example: integrate with ghq](#example-integrate-with-ghq)
  - [Configuration](#configuration)
    - [Location](#location)
    - [Basic example](#basic-example)
    - [Template variables](#template-variables)
    - [Argument injection rules](#argument-injection-rules)
  - [Development](#development)
    - [Requirements](#requirements)
    - [Quick commands](#quick-commands)
    - [Standard Go commands](#standard-go-commands)
    - [Project structure](#project-structure)
  - [License](#license)
  - [Contributing](#contributing)

<!-- /TOC -->

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

## Usage

### Basic usage

```bash
# List commands
$ sidetable list
example        An example command

# Run a command.
$ sidetable example arg1 arg2

# Show help
$ sidetable --help
$ sidetable help

# Show version
$ sidetable --version
$ sidetable version
```

### Example: integrate with [ghq](https://github.com/x-motemen/ghq)

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
      GHQ_ROOT: "{{.CommandDir}}" # See Configuration section for details
    description: "Manage repositories in project-local directory"
    alias: "q"
```

Example:

```bash
$ cd ~/myproject
$ sidetable ghq https://github.com/example/repo
# Or you can use alias: `sidetable q https://github.com/example/repo`
#
# => cloned into ~/myproject/.local/ghq/github.com/example/repo
```

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
    # If omitted, sidetable just runs `<command>` with user-provided args.
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

### Template variables

Each string in `command`, `args`, `env`, and `description` is evaluated as a Go template. Available variables:

| Variable      | Description                                                |
| ------------- | ---------------------------------------------------------- |
| `.ProjectDir` | current directory when running `sidetable` (absolute path) |
| `.PrivateDir` | `.ProjectDir/<directory>` (absolute path)                  |
| `.CommandDir` | `.PrivateDir/<commandName>` (absolute path)                |
| `.ConfigDir`  | directory containing `config.yml` (absolute path)          |

### Argument injection rules

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
