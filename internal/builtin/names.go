package builtin

// IsReservedCommand returns true when name is reserved as a built-in command.
func IsReservedCommand(name string) bool {
	switch name {
	case "list", "completion", "doctor", "init", "help":
		return true
	default:
		return false
	}
}
