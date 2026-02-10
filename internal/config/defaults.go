package config

import (
	_ "embed"
)

//go:embed init.yml
var DefaultConfigYAML []byte
