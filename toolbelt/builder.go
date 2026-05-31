package toolbelt

import (
	"context"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
)

// BuildContext contains the stable sqlc intermediate representation passed to builders.
type BuildContext struct {
	Context     context.Context
	Catalog     *models.Catalog
	Queries     []models.Query
	Logger      *logger.Logger
	SqlcVersion string
}

// Builder generates files from the toolbelt intermediate representation.
type Builder interface {
	Build(ctx BuildContext) ([]File, error)
}
