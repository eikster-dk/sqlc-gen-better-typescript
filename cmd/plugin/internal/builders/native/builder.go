package native

import (
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

// File represents a generated file.
type File struct {
	Name    string
	Content []byte
}

// Native is the native TypeScript builder using plain async functions,
// Zod validation, and the pg driver.
type Native struct {
	cfg config.Config
}

// New creates a new Native builder with the given config.
func New(cfg config.Config) *Native {
	return &Native{cfg: cfg}
}

// Build generates files from the internal representation.
// Phase 1: stub -- returns an empty file list without error.
func (n *Native) Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger, sqlcVersion string) ([]File, error) {
	log.Info("Starting native code generation", logger.F("builder", "native"))
	log.Debug("Catalog info", logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)))
	log.Info("Native code generation complete (stub)", logger.F("files", 0))
	return []File{}, nil
}
