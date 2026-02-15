# sidetable

**sidetable** is a Go CLI and library for managing a project-local tool area (for example `.$USER/`) and delegating external tools.

Define tools in config, then execute them as `sidetable <tool-or-alias> ...`.
Each tool gets its own directory under your configured local area, so you can keep per-project state in a predictable layout.

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
    - [Shell Completion](#shell-completion)
  - [Configuration](#configuration)
    - [Location](#location)
    - [Basic example](#basic-example)
    - [Template variables](#template-variables)
    - [Argument injection rules](#argument-injection-rules)
  - [Development](#development)
    - [Requirements](#requirements)
    - [Quick commands](#quick-commands)
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

Download the latest release binaries from [Releases](https://github.com/sushichan044/sidetable/releases).

## Usage

### Basic usage

```bash
# List tools and aliases
$ sidetable list
NAME           KIND      TARGET   DESCRIPTION
example        tool      -        An example tool
ex             alias     example  Shortcut for example

# Run a tool or alias
$ sidetable example arg1 arg2
$ sidetable ex arg1 arg2

# Show help
$ sidetable --help
$ sidetable help

# Show version
$ sidetable --version
```

### Example: integrate with [ghq](https://github.com/x-motemen/ghq)

```yaml
directory: ".private"

tools:
  ghq:
    run: "ghq"
    env:
      GHQ_ROOT: "{{.ToolDir}}"
    description: "Manage repositories in project-local directory"

aliases:
  q:
    tool: "ghq"
    args:
      prepend:
        - "get"
        - "-u"
    description: "Clone repository into project-local directory"
```

Example:

```bash
$ cd ~/myproject
$ sidetable q https://github.com/example/repo
# Or call the tool directly:
# sidetable ghq get -u https://github.com/example/repo
# => cloned into ~/myproject/.private/ghq/github.com/example/repo
```

### Shell Completion

You can enable shell completion for sidetable commands, including tools and aliases defined in config.

```bash
# bash
source <(sidetable completion bash)

# zsh
source <(sidetable completion zsh)

# fish
sidetable completion fish | source

# PowerShell
sidetable completion powershell | Out-String | Invoke-Expression
```

## Configuration

### Location

`config.yml` is searched in the following order:

1. If `SIDETABLE_CONFIG_DIR` is set: `$SIDETABLE_CONFIG_DIR/config.yml`
2. Otherwise: `$XDG_CONFIG_HOME/sidetable/config.yml` (or `~/.config/sidetable/config.yml` if `XDG_CONFIG_HOME` is not set)

### Basic example

```yaml
# Required. Project-local tool area name (relative path).
directory: ".sidetable"

tools:
  ghq:
    # Required. Program name to execute.
    # Templating: allowed.
    run: "ghq"
    # Optional. Arguments to inject.
    # Order: tool.prepend + userArgs + tool.append
    # Templating: allowed.
    args:
      # prepend:
      #   - "--some-flag"
      # append:
      # - "--some-flag"
    # Optional. Override environment variables for the tool.
    # Templating: allowed.
    env:
      GHQ_ROOT: "{{.ToolDir}}"
    # Optional. Description shown in `sidetable list`.
    description: "ghq wrapper with project-local root"

  note:
    run: "{{.ConfigDir}}/vim-note.sh"
    args:
      append:
        - "{{.ToolDir}}/note.md"
    description: "Open project note file"

# Optional. Aliases for tools.
aliases:
  gg:
    # Required. Target tool name defined in `tools`.
    tool: "ghq"
    # Optional. Arguments to inject.
    # Order: alias.prepend + tool.prepend + userArgs + tool.append + alias.append
    # Templating: allowed.
    args:
      prepend:
        - "get"
        - "-u"
      # append:
      # - "--some-flag"
    # Optional. Description shown in `sidetable list`.
    description: "ghq get shortcut"
```

### Template variables

These fields are treated as Go text/template and rendered with the following variables.

- `tools.<toolName>.run`
- `tools.<toolName>.args.prepend`
- `tools.<toolName>.args.append`
- `tools.<toolName>.env.<envVar>`
- `aliases.<aliasName>.args.prepend`
- `aliases.<aliasName>.args.append`

| Variable         | Description                              |
| ---------------- | ---------------------------------------- |
| `.WorkspaceRoot` | current directory when running sidetable |
| `.ToolDir`       | `.WorkspaceRoot/<directory>/<toolName>`  |
| `.ConfigDir`     | directory containing the config file     |

All directory variables are absolute paths.

### Argument injection rules

Arguments are concatenated in the following order.

```text
alias.prepend + tool.prepend + userArgs + tool.append + alias.append
```

Example:

```yaml
tools:
  example:
    run: "mycommand"
    args:
      prepend:
        - "--flag"
      append:
        - "--output=result.txt"
aliases:
  ex:
    tool: "example"
    args:
      prepend:
        - "--alias-start"
      append:
        - "--alias-end"
```

```bash
$ sidetable ex arg1 arg2
# Executed command:
# mycommand --alias-start --flag arg1 arg2 --output=result.txt --alias-end
```

## Development

### Requirements

- Go 1.25+
- [mise](https://mise.jdx.dev/) (optional)

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

## License

[MIT](LICENSE)

## Contributing

Issues and PRs are welcome.
