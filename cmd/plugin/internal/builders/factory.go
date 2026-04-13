package builders

import (
	"fmt"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/builders/effect4"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/builders/native"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

type effect4Adapter struct {
	impl *effect4.Effect4
}

func (a *effect4Adapter) Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger, sqlcVersion string) ([]File, error) {
	files, err := a.impl.Build(catalog, queries, log, sqlcVersion)
	if err != nil {
		return nil, err
	}

	converted := make([]File, len(files))
	for i, f := range files {
		converted[i] = File{Name: f.Name, Content: f.Content}
	}

	return converted, nil
}

type nativeAdapter struct {
	impl *native.Native
}

func (a *nativeAdapter) Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger, sqlcVersion string) ([]File, error) {
	files, err := a.impl.Build(catalog, queries, log, sqlcVersion)
	if err != nil {
		return nil, err
	}

	converted := make([]File, len(files))
	for i, f := range files {
		converted[i] = File{Name: f.Name, Content: f.Content}
	}

	return converted, nil
}

// NewBuilder creates a builder based on the config
func NewBuilder(cfg config.Config) (Builder, error) {
	switch cfg.Builder {
	case "effect-v4-unstable":
		return &effect4Adapter{impl: effect4.New(cfg)}, nil
	case "native":
		return &nativeAdapter{impl: native.New(cfg)}, nil
	default:
		return nil, fmt.Errorf("unknown builder: %s", cfg.Builder)
	}
}
