package config_test

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/config"
)

func TestDefaultConfigYAMLValid(t *testing.T) {
	require.NotEmpty(t, config.DefaultConfigYAML)

	var cfg config.Config
	require.NoError(t, yaml.Unmarshal(config.DefaultConfigYAML, &cfg))
	require.NoError(t, cfg.Validate())
}
