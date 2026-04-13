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
func (n *Native) Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger, sqlcVersion string) ([]File, error) {
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

	modelsFile, err := n.generateModelsFileFromTemplates(tmpls, catalog, sqlcVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate models file: %w", err)
	}

	files := []File{modelsFile}

	if len(queries) > 0 {
		queryGroups := n.groupQueriesByFile(queries, log)
		filenames := sortedGroupKeys(queryGroups)

		for _, filename := range filenames {
			fileQueries := queryGroups[filename]
			stem := filenameToStem(filename)
			viewName := toCamelCase(stem) // "customers" -> "customers"
			queryViews := n.buildQueryViews(fileQueries, log)

			log.Info("Generating query files", logger.F("file", filename), logger.F("queries", len(fileQueries)))

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
	}

	log.Info("Native code generation complete", logger.F("files", len(files)))
	return files, nil
}
