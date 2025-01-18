package start

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

// ParsePGConfig
func ParsePGConfig(filePath string) (*PGConfig, error) {
	var config struct {
		PG PGConfig `toml:"PG"`
	}

	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, fmt.Errorf("failed to decode pg.toml: %w", err)
	}

	return &config.PG, nil
}
