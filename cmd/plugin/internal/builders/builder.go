package builders

import (
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/models"
)

type File struct {
	Name    string
	Content []byte
}

// Builder is the interface for code generators
type Builder interface {
	// Build generates files from the internal representation
	// The catalog contains all schema information (tables, enums)
	// The queries contain all the SQL queries with their params and results
	// The logger is used for structured logging during generation
	Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger) ([]File, error)
}
