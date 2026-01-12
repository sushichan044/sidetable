# Implementation Notes (Delegate + Config)

## 1) Argument parsing and delegation

- Parse **only global flags** first (e.g., `--version`, `--help`) and leave the rest untouched.
  - Use `pflag` (or equivalent) with **unknown flags allowed** to avoid breaking delegated commands.
- After global parsing:
  - If the first token is a **built-in subcommand** (`list`, `help`, `version`), route to the CLI framework.
  - Otherwise, **delegate** without trying to parse any more flags/args.
- Keep `stdin/stdout/stderr` intact and return the delegated process exit code as-is.

## 2) Config path resolution

- Default config path: `SIDETABLE_CONFIG_DIR/config.{yml,yaml}`.
- If `SIDETABLE_CONFIG_DIR` is unset, use `XDG_CONFIG_HOME/sidetable`.
- If `XDG_CONFIG_HOME` is unset, use `~/.config`.
- This repo already has `internal/xdg` for the fallback.

## 3) Command execution model

- **Do not** execute through a shell; use `exec.Command`.
- `command` is a **required** executable name (no spaces).
- `argv` is optional; build the final argv list by:
  1) `command` (evaluated from template)
  2) `argv` (each element template-evaluated)
  3) `userArgs` (only when allowed)
- If `argv` **does not contain** `{{.Args}}` and `userArgs` is non-empty, return an error.
- If `argv` **contains** `{{.Args}}`, replace it with `userArgs` and **flatten** into the argv list.
- `{{.Args}}` can appear **at most once** in `argv`.

## 4) Template evaluation

- Evaluate templates per **string element** (`command`, each `argv` item, each `env` value, `description`).
- Use `missingkey=error`.
- Template variables (see SPEC):
  - `.ProjectDir`, `.PrivateDir`, `.CommandDir`, `.ConfigDir`, `.Args`

## 5) Validation tips

- `directory` is required and must be a **relative** path.
- `command` is required and must be **non-empty and no spaces**.
- `alias` must be unique and must not collide with command names.

## 6) Completion (out of scope)

- Completion is explicitly out of scope for this phase.
- When adding later, avoid parsing delegated flags; prefer a `__complete`-style endpoint.
