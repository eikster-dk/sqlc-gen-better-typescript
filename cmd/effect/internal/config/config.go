package config

import (
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type Config struct {
	Debug                   bool   `json:"debug"`
	DebugDir                string `json:"debug_dir"`
	DisableTemplateLiterals bool   `json:"disable_template_literals"`
	ImportExtension         string `json:"import_extension"`
}

func Parse(req *plugin.GenerateRequest) (Config, error) {
	return toolbelt.ParseJSONConfig[Config](req)
}

func Validate(cfg Config, req *plugin.GenerateRequest) error {
	if err := toolbelt.RequireEngine(req, "postgresql"); err != nil {
		return err
	}
	return toolbelt.RequireImportExtension(cfg.ImportExtension)
}
