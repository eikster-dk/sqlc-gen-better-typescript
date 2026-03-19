package internal

import (
	"context"

	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/builders"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/debug"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/mapper"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	// 1. Parse config
	cfg, err := config.GetConfig(req)
	if err != nil {
		return nil, err
	}

	// 2. Initialize logger (enabled if debug mode is on)
	log := logger.New(cfg.Debug)

	log.Info("Starting code generation",
		logger.F("builder", cfg.Builder),
		logger.F("debug", cfg.Debug))

	// 3. Validate config
	if err := config.Validate(cfg); err != nil {
		log.Error("Config validation failed", err)
		return nil, err
	}
	log.Info("Config validated successfully")

	// 4. Map sqlc types to internal models
	log.Info("Mapping sqlc types to internal models")
	m := mapper.New(req, log)
	catalog := m.Catalog()
	queries := m.MapQueries(req)
	log.Info("Mapped queries",
		logger.F("count", len(queries)),
		logger.F("tables", len(catalog.Tables)),
		logger.F("enums", len(catalog.Enums)))

	// 5. Create builder and generate files
	builder, err := builders.NewBuilder(cfg)
	if err != nil {
		log.Error("Failed to create builder", err)
		return nil, err
	}
	log.Info("Builder created", logger.F("type", cfg.Builder))

	files, err := builder.Build(catalog, queries, log)
	if err != nil {
		log.Error("Build failed", err)
		return nil, err
	}
	log.Info("Files generated", logger.F("count", len(files)))

	// 6. Generate debug artifacts if enabled
	if cfg.Debug {
		log.Info("Generating debug artifacts")
		debugFiles := debug.GenerateArtifacts(catalog, queries, req, log, cfg.DebugDir)
		files = append(files, debugFiles...)
		log.Info("Debug artifacts added", logger.F("count", len(debugFiles)))
	}

	log.Info("Code generation complete")

	// 7. Convert builder files to plugin files
	pluginFiles := make([]*plugin.File, len(files))
	for i, f := range files {
		pluginFiles[i] = &plugin.File{
			Name:     f.Name,
			Contents: f.Content,
		}
	}

	return &plugin.GenerateResponse{
		Files: pluginFiles,
	}, nil
}
