package cmd

var (
	builtinCommands = map[string]struct{}{
		"list":       {},
		"completion": {}, // defined by cobra
		"doctor":     {},
		"init":       {},
	}
)

func isBuiltinCommand(name string) bool {
	_, ok := builtinCommands[name]
	return ok
}
