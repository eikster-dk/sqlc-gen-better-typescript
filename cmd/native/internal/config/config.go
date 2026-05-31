package config

import (
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Config struct {
	Debug           bool   `json:"debug"`
	DebugDir        string `json:"debug_dir"`
	ImportExtension string `json:"import_extension"`
	Driver          string `json:"driver"`
	Validator       string `json:"validator"`
}

func Parse(req *plugin.GenerateRequest) (Config, error) {
	cfg, err := toolbelt.ParseJSONConfig[Config](req)
	if err != nil {
		return cfg, err
	}
	if cfg.ImportExtension == "" {
		cfg.ImportExtension = ".js"
	}
	if cfg.Driver == "" {
		cfg.Driver = "pg"
	}
	if cfg.Validator == "" {
		cfg.Validator = "zod"
	}
	return cfg, nil
}

func Validate(cfg Config, req *plugin.GenerateRequest) error {
	if err := toolbelt.RequireEngine(req, "postgresql"); err != nil {
		return err
	}
	if err := toolbelt.RequireImportExtension(cfg.ImportExtension); err != nil {
		return err
	}
	if err := toolbelt.RequireOneOf("driver", cfg.Driver, "pg"); err != nil {
		return err
	}
	return toolbelt.RequireOneOf("validator", cfg.Validator, "zod")
}
