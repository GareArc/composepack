package config

import (
	"fmt"

	ms "github.com/go-viper/mapstructure/v2"
)

// Config contains process-wide settings derived from flags/env.
type Config struct {
	ReleasesBaseDir string `mapstructure:"releases_base_dir"`
}

// Default returns baseline configuration derived from the PRD runtime layout.
func Default() Config {
	return Config{
		ReleasesBaseDir: ".cpack-releases",
	}
}

// NewWithSubstitute creates a new config with the given substitutions mappings.
// The substitutions map is a mapping of config keys to values.
func NewWithSubstitutions(substitutions map[string]string) (Config, error) {
	// map over the config and substitute the values by mapstructure
	config := Default()
	err := ms.Decode(substitutions, &config)
	if err != nil {
		return config, fmt.Errorf("failed to decode substitutions: %w", err)
	}
	return config, nil
}
