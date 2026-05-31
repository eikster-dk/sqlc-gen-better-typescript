package internal

import (
	"context"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/native/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/native/internal/native"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	return toolbelt.Generate(ctx, req, toolbelt.Options[config.Config]{
		ParseConfig:    config.Parse,
		ValidateConfig: config.Validate,
		NewBuilder: func(cfg config.Config) (toolbelt.Builder, error) {
			return native.New(cfg), nil
		},
		Debug: func(cfg config.Config) toolbelt.DebugOptions {
			return toolbelt.DebugOptions{Enabled: cfg.Debug, Dir: cfg.DebugDir, IncludeRequest: true}
		},
	})
}
