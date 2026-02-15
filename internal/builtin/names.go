package builtin

// IsReservedName returns true when name is reserved as a built-in CLI command.
func IsReservedName(name string) bool {
	switch name {
	case "list", "completion", "init", "help":
		return true
	default:
		return false
	}
}
