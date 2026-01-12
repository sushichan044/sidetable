package main

var (
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
