package toolbelt

import (
	"context"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/mapper"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// DebugOptions controls generic debug artifacts emitted by toolbelt.
type DebugOptions struct {
	Enabled        bool
	Dir            string
	IncludeRequest bool
}

// Options configures the generic sqlc generation pipeline for a plugin-specific config type.
type Options[T any] struct {
	ParseConfig    func(*plugin.GenerateRequest) (T, error)
	ValidateConfig func(T, *plugin.GenerateRequest) error
	NewBuilder     func(T) (Builder, error)
	Debug          func(T) DebugOptions
}

// Generate runs the shared sqlc plugin pipeline and delegates output-specific work to a builder.
func Generate[T any](ctx context.Context, req *plugin.GenerateRequest, opts Options[T]) (*plugin.GenerateResponse, error) {
	cfg, err := opts.ParseConfig(req)
	if err != nil {
		return nil, err
	}

	debugOpts := DebugOptions{}
	if opts.Debug != nil {
		debugOpts = opts.Debug(cfg)
	}

	log := logger.New(debugOpts.Enabled)
	log.Info("Starting code generation", logger.F("debug", debugOpts.Enabled))

	if opts.ValidateConfig != nil {
		if err := opts.ValidateConfig(cfg, req); err != nil {
			log.Error("Config validation failed", err)
			return nil, err
		}
		log.Info("Config validated successfully")
	}

	log.Info("Mapping sqlc types to intermediate representation")
	m := mapper.New(req, log)
	catalog := m.Catalog()
	queries := m.MapQueries(req)
	log.Info("Mapped queries", logger.F("count", len(queries)), logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)))

	builder, err := opts.NewBuilder(cfg)
	if err != nil {
		log.Error("Failed to create builder", err)
		return nil, err
	}

	files, err := builder.Build(BuildContext{
		Context:     ctx,
		Catalog:     catalog,
		Queries:     queries,
		Logger:      log,
		SqlcVersion: req.SqlcVersion,
	})
	if err != nil {
		log.Error("Build failed", err)
		return nil, err
	}
	log.Info("Files generated", logger.F("count", len(files)))

	if debugOpts.Enabled {
		debugFiles := generateDebugArtifacts(catalog, queries, req, log, debugOpts)
		files = append(files, debugFiles...)
		log.Info("Debug artifacts added", logger.F("count", len(debugFiles)))
	}

	pluginFiles := make([]*plugin.File, len(files))
	for i, f := range files {
		pluginFiles[i] = &plugin.File{Name: f.Name, Contents: f.Content}
	}

	return &plugin.GenerateResponse{Files: pluginFiles}, nil
}
