package main

var (
	//nolint:gochecknoglobals // This is necessary
	builtinCommands = map[string]struct{}{
		"version":    {},
		"list":       {},
		"completion": {},
	}
)

func isBuiltinCommand(name string) bool {
	_, ok := builtinCommands[name]
	return ok
}
