package config

import (
	_ "embed"
)

//go:embed default/config.yml
var DefaultConfigYAML []byte
