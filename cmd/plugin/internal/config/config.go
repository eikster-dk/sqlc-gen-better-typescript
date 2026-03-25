package config

import (
	"encoding/json"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Config struct {
	Builder                 string `json:"builder"`
	Debug                   bool   `json:"debug"`                     // Enable debug mode
	DebugDir                string `json:"debug_dir"`                 // Optional, defaults to "debug"
	DisableTemplateLiterals bool   `json:"disable_template_literals"` // Opt-out: use sql.unsafe instead of template literals
	ImportExtension         string `json:"import_extension"`          // Optional relative import extension: "", ".js", ".ts"
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
			if cfg.ImportExtension == "" || cfg.ImportExtension == ".js" || cfg.ImportExtension == ".ts" {
				return nil
			}
			return fmt.Errorf("Option: import_extension value is %s but can only be one of [\"\", \".js\", \".ts\"]", cfg.ImportExtension)
		}
	}
	return fmt.Errorf("Option: builder value is %s but can only be one of %v", cfg.Builder, ValidBuilders)
}
