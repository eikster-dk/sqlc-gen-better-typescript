package config

import (
	"encoding/json"
	"fmt"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Config struct {
	Builder                 string  `json:"builder"`
	Debug                   bool    `json:"debug"`                     // Enable debug mode
	DebugDir                string  `json:"debug_dir"`                 // Optional, defaults to "debug"
	DisableTemplateLiterals bool    `json:"disable_template_literals"` // Opt-out: use sql.unsafe instead of template literals
	ImportExtension         *string `json:"import_extension"`          // Optional relative import extension: "", ".js", ".ts" (nil means use default ".js")
	Driver                  string  `json:"driver"`                    // DB driver for native builder, default "pg"
	Validator               string  `json:"validator"`                 // Validation library for native builder, default "zod"
}

var ValidBuilders = []string{
	"effect-v4-unstable",
	"native",
}

const SupportedEngine = "postgresql"

func GetConfig(req *plugin.GenerateRequest) (Config, error) {
	var conf Config
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &conf); err != nil {
			return conf, err
		}
	}

	if conf.Builder == "" {
		conf.Builder = "native"
	}
	if conf.ImportExtension == nil {
		defaultExt := ".js"
		conf.ImportExtension = &defaultExt
	}
	if conf.Driver == "" {
		conf.Driver = "pg"
	}
	if conf.Validator == "" {
		conf.Validator = "zod"
	}

	return conf, nil
}

func Validate(cfg Config, req *plugin.GenerateRequest) error {
	engine := req.GetSettings().GetEngine()
	if engine != SupportedEngine {
		return fmt.Errorf("Option: engine value is %q but this plugin currently only supports %q", engine, SupportedEngine)
	}

	for _, builder := range ValidBuilders {
		if cfg.Builder == builder {
			ext := ""
			if cfg.ImportExtension != nil {
				ext = *cfg.ImportExtension
			}
			if ext == "" || ext == ".js" || ext == ".ts" {
				return nil
			}
			return fmt.Errorf("Option: import_extension value is %s but can only be one of [\"\", \".js\", \".ts\"]", ext)
		}
	}
	return fmt.Errorf("Option: builder value is %s but can only be one of %v", cfg.Builder, ValidBuilders)
}
