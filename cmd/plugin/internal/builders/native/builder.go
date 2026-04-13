package native

import (
	"fmt"

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
func (n *Native) Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger, sqlcVersion string) ([]File, error) {
	log.Info("Starting native code generation", logger.F("builder", "native"))
	log.Debug("Catalog info", logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)))

	modelsFile, err := n.generateModelsFile(catalog, sqlcVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate models file: %w", err)
	}

	files := []File{modelsFile}
	log.Info("Native code generation complete", logger.F("files", len(files)))
	return files, nil
}
