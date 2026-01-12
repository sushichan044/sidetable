# AGENTS.md

This file provides guidance to Coding Agents when working with code in this repository.

---

## Agent Guidelines

- **Language**: English code/docs/comments
- **Execution**: Test→Lint→Cleanup workflow
- **Files**: Cleanup temp files after use
- **Efficiency**: Parallel operations for independent tasks
- **Path Handling**: You MUST treat all paths as relative to the cwd unless starts with `/` (root) or a drive letter (e.g., `C:\`).

## Quick Commands

```bash
mise run test               # Run tests with coverage
mise run lint               # Run golangci-lint for code quality checks
mise run lint-fix           # Run golangci-lint and Auto-fix linting issues
mise run fmt                # Format code
mise run build-snapshot     # Build cross-platform binaries with goreleaser
mise run clean              # Remove generated files

# Standard Go commands
go run ./cmd/sidetable           # Run CLI in development mode
go test ./...               # Run all tests
go mod tidy                 # Clean up dependencies
```

## Project Context

sidetable is a Go CLI tool for managing personal directories per project.
See README.md for more details.

## Sources of Truth

Keep this file light. For implementation details, refer to:

- Product and usage overview: `README.md`
- CLI entry point: `cmd/sidetable/main.go`
- Package layout and behavior: `internal/`
- Dependencies and versions: `go.mod`, `go.sum`
- Task runner and scripts: `mise.toml`
- Lint/format rules: `.golangci.yml`
- Release/build configuration: `.goreleaser.yaml`

## Coding Standard

- use `pflag` for CLI flag parsing.
- use `cobra` for CLI framework. (built-in commands only; delegation is custom)
- use `testify` for testing and assertions.
- follow Go idioms and best practices.
- Check code quality with `mise run fmt` and `mise run lint-fix` and `mise run test` before completing work.
