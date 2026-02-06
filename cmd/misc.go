package cmd

var (
	builtinCommands = map[string]struct{}{
		"list":       {},
		"completion": {}, // defined by cobra
		"doctor":     {},
	}
)

func isBuiltinCommand(name string) bool {
	_, ok := builtinCommands[name]
	return ok
}
