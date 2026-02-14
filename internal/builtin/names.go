package builtin

// CommandNames defines built-in command names reserved by sidetable/cobra.
var CommandNames = map[string]struct{}{
	"list":       {},
	"completion": {},
	"doctor":     {},
	"init":       {},
	"help":       {},
}

// IsReservedCommand returns true when name is reserved as a built-in command.
func IsReservedCommand(name string) bool {
	_, ok := CommandNames[name]
	return ok
}
