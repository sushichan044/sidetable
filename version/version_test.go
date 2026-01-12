package version_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/sushichan044/sidetable/version"
)

func TestGet(t *testing.T) {
	t.Parallel()

	v := version.Get()

	// Version should never be empty
	if v == "" {
		t.Error("version.Get() returned empty string")
	}

	// When built without version info, should return "dev"
	// When built with tags, should return a version string
	// We can't predict the exact value, but we can verify it's not empty
	t.Logf("Version: %s", v)
}

func TestGetDetailed(t *testing.T) {
	t.Parallel()

	detailed := version.GetDetailed()

	// Should have at least version, go_version, and path
	if detailed["version"] == "" {
		t.Error("detailed version info missing 'version' key")
	}

	if detailed["go_version"] == "" {
		t.Error("detailed version info missing 'go_version' key")
	}

	if detailed["path"] == "" {
		t.Error("detailed version info missing 'path' key")
	}

	// Verify go_version format (should be like "go1.24.0")
	goVersion := detailed["go_version"]
	if !strings.HasPrefix(goVersion, "go") {
		t.Errorf("go_version has unexpected format: %s", goVersion)
	}

	t.Logf("Detailed version info:")
	for k, v := range detailed {
		t.Logf("  %s: %s", k, v)
	}
}

func TestGetVersionFormat(t *testing.T) {
	t.Parallel()

	v := version.Get()

	// Should be one of these formats:
	// - "dev" (local build)
	// - "vX.Y.Z" (tagged release)
	// - "vX.Y.Z (rev: abc1234)" (with revision)
	// - "vX.Y.Z (rev: abc1234, modified)" (with uncommitted changes)
	// - "unknown" (edge case)

	validFormats := []string{"dev", "unknown"}
	isValid := slices.Contains(validFormats, v)

	// Or starts with "v" (version tag)
	if strings.HasPrefix(v, "v") {
		isValid = true
	}

	if !isValid {
		t.Errorf("version format unexpected: %s", v)
	}
}
