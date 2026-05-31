package native

import (
	"fmt"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/native/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
)

// Native is the native TypeScript builder using plain async functions,
// Zod validation, and the pg driver.
type Native struct {
	cfg        config.Config
	enumValues map[string][]string // populated during Build from catalog enums
}

// New creates a new Native builder with the given config.
func New(cfg config.Config) *Native {
	return &Native{cfg: cfg}
}

// buildEnumValues constructs a lookup map from enum name to its ordered value strings.
func buildEnumValues(catalog *models.Catalog) map[string][]string {
	m := make(map[string][]string, len(catalog.Enums))
	for _, e := range catalog.Enums {
		values := make([]string, len(e.Values))
		for i, v := range e.Values {
			values[i] = v.Value
		}
		m[e.Name] = values
	}
	return m
}

// Build generates files from the internal representation.
func (n *Native) Build(ctx toolbelt.BuildContext) ([]toolbelt.File, error) {
	catalog := ctx.Catalog
	queries := ctx.Queries
	log := ctx.Logger
	sqlcVersion := ctx.SqlcVersion
	if log == nil {
		log = logger.New(false)
	}

	log.Info("Starting native code generation", logger.F("builder", "native"))

	if catalog == nil {
		catalog = &models.Catalog{}
	}

	n.enumValues = buildEnumValues(catalog)

	log.Debug("Catalog info", logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)))

	tmpls, err := loadAllTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	plans := models.BuildQueryPlans(queries)
	if err := validateSupportedCommands(plans); err != nil {
		return nil, err
	}
	queryGroups := n.groupQueriesByFile(plans, log)
	filenames := sortedGroupKeys(queryGroups)
	usedEmbedTables := collectUsedEmbedTables(plans)

	modelsFile, err := n.generateModelsFileFromTemplates(tmpls, catalog, usedEmbedTables, sqlcVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate models file: %w", err)
	}

	files := []toolbelt.File{modelsFile}

	for _, filename := range filenames {
		filePlans := queryGroups[filename]
		stem := filenameToStem(filename)
		viewName := toCamelCase(stem) // "customers" -> "customers"
		queryViews := n.buildQueryViews(filePlans, log)

		log.Info("Generating query files", logger.F("file", filename), logger.F("queries", len(filePlans)))

		requestsFile, responsesFile, queriesFile, err := n.generateQueryFiles(viewName, queryViews, tmpls, sqlcVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to generate query files for %s: %w", filename, err)
		}

		files = append(files, requestsFile, responsesFile, queriesFile)
		log.Info("Generated query files",
			logger.F("requests", requestsFile.Name),
			logger.F("responses", responsesFile.Name),
			logger.F("queries", queriesFile.Name))
	}

	log.Info("Native code generation complete", logger.F("files", len(files)))
	return files, nil
}

func collectUsedEmbedTables(plans []models.QueryPlan) map[string]struct{} {
	tables := make(map[string]struct{})
	for _, plan := range plans {
		if !plan.Features.UsesEmbeds {
			continue
		}
		for _, field := range plan.Response.Result.Shape.Fields {
			if field.Kind == models.ResultShapeFieldObject {
				tables[field.Name] = struct{}{}
			}
		}
	}
	return tables
}

func validateSupportedCommands(plans []models.QueryPlan) error {
	for _, plan := range plans {
		switch plan.Command {
		case ":one", ":many", ":exec", ":execrows", ":execresult":
			continue
		case ":copyfrom", ":batchexec", ":batchone", ":batchmany":
			return fmt.Errorf("unsupported sqlc command %s for query %s", plan.Command, plan.Name)
		default:
			return fmt.Errorf("unsupported sqlc command %s for query %s", plan.Command, plan.Name)
		}
	}
	return nil
}
