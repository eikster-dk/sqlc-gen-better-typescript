package config

import (
	"encoding/json"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Config struct {
	Builder  string `json:"builder"`
	Debug    bool   `json:"debug"`     // Enable debug mode
	DebugDir string `json:"debug_dir"` // Optional, defaults to "debug"
}

var ValidBuilders = []string{
	"effect-v4-unstable",
}

func GetConfig(req *plugin.GenerateRequest) (Config, error) {
	var conf Config
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &conf); err != nil {
			return conf, err
		}
	}

	return conf, nil
}

func Validate(cfg Config) error {
	for _, builder := range ValidBuilders {
		if cfg.Builder == builder {
			return nil
		}
	}
	return fmt.Errorf("Option: builder value is %s but can only be one of %v", cfg.Builder, ValidBuilders)
}
