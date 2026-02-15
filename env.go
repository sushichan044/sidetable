package sidetable

import (
	"fmt"
	"maps"
	"strings"
)

// envMapFromSlice converts environment variables from slice format.
//
//	envMap := envMapFromSlice([]string{"KEY=value", "FOO=bar"})
//	fmt.Println(envMap["KEY"]) // Output: value
func envMapFromSlice(env []string) map[string]string {
	result := make(map[string]string, len(env))
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if ok {
			result[key] = value
		}
	}
	return result
}

// envSliceFromMap converts environment variables from map format.
//
//	envSlice := envSliceFromMap(map[string]string{"KEY": "value", "FOO": "bar"})
//	fmt.Println(envSlice) // Output: []string{"KEY=value", "FOO=bar"}
func envSliceFromMap(envMap map[string]string) []string {
	result := make([]string, 0, len(envMap))
	for key, value := range envMap {
		result = append(result, fmt.Sprintf("%s=%s", key, value))
	}
	return result
}

func mergeMap(base map[string]string, overrides map[string]string) map[string]string {
	result := make(map[string]string, len(base)+len(overrides))
	maps.Copy(result, base)
	maps.Copy(result, overrides)
	return result
}
