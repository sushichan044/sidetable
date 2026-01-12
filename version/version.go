// Package version provides version information for the memo-cli application.
// Version information is automatically embedded at build time using Go's build info.
package version

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const (
	// gitShortHashLength is the standard length for git short hashes (7 characters).
	gitShortHashLength = 7
)

// Get returns the version information of the memo-cli application.
// It reads version from runtime/debug.ReadBuildInfo() which is automatically
// populated when built with Go modules and version tags.
//
// The version format is:
//   - "vX.Y.Z" when built from a tagged release
//   - "dev" when built locally without version info
//   - "vX.Y.Z (rev: abc1234)" when built with VCS revision info
//   - "vX.Y.Z (rev: abc1234, modified)" when built with uncommitted changes
func Get() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	version := info.Main.Version
	if version == "" || version == "(devel)" {
		version = "dev"
	}

	// Extract VCS information from build settings
	var revision string
	var modified bool
	var dirty string

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			revision = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}

	// Format revision info
	if revision != "" {
		// Shorten revision to first gitShortHashLength characters (git short hash style)
		if len(revision) > gitShortHashLength {
			revision = revision[:gitShortHashLength]
		}

		if modified {
			dirty = ", modified"
		}

		return fmt.Sprintf("%s (rev: %s%s)", version, revision, dirty)
	}

	return version
}

// GetDetailed returns detailed version information including Go version and dependencies.
// This can be useful for debugging and support purposes.
func GetDetailed() map[string]string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return map[string]string{
			"version": "unknown",
		}
	}

	result := map[string]string{
		"version":    Get(),
		"go_version": info.GoVersion,
		"path":       info.Path,
	}

	// Add all build settings
	for _, setting := range info.Settings {
		// Use dot notation for nested keys (e.g., vcs.revision)
		key := strings.ReplaceAll(setting.Key, ".", "_")
		result[key] = setting.Value
	}

	return result
}
