package effect4

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/effect/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
)

type Effect4 struct {
	cfg config.Config
}

func New(cfg config.Config) *Effect4 {
	return &Effect4{cfg: cfg}
}

func (e *Effect4) Build(ctx toolbelt.BuildContext) ([]toolbelt.File, error) {
	log := ctx.Logger
	catalog := ctx.Catalog
	if catalog == nil {
		catalog = &models.Catalog{}
	}

	log.Info("Starting Effect4 code generation")
	log.Debug("Catalog info", logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)))

	tmpls, err := loadTemplates(log)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	queryGroups := e.groupQueriesByFile(ctx.Queries, log)
	filenames := sortedGroupKeys(queryGroups)

	queryViewsByFile := make(map[string][]QueryView, len(queryGroups))
	for _, filename := range filenames {
		queryViewsByFile[filename] = e.buildQueryViews(queryGroups[filename], log)
	}

	modelsFile, err := e.generateModelsFile(tmpls.models, catalog, queryViewsByFile, ctx.SqlcVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate models file: %w", err)
	}

	files := []toolbelt.File{modelsFile}
	for _, filename := range filenames {
		queryViews := queryViewsByFile[filename]
		repoName := e.filenameToRepoName(filename)
		log.Info("Generating repository", logger.F("file", filename), logger.F("queries", len(queryGroups[filename])))

		requestFile, err := e.generateRequestFile(tmpls.request, repoName, queryViews, ctx.SqlcVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to generate request file for %s: %w", filename, err)
		}
		responseFile, err := e.generateResponseFile(tmpls.response, repoName, queryViews, ctx.SqlcVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to generate response file for %s: %w", filename, err)
		}
		repositoryFile, err := e.generateRepositoryFile(tmpls.repository, filename, queryViews, ctx.SqlcVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to generate repository for %s: %w", filename, err)
		}

		files = append(files, requestFile, responseFile, repositoryFile)
		log.Info("Generated repository files", logger.F("repository", repositoryFile.Name), logger.F("request", requestFile.Name), logger.F("response", responseFile.Name))
	}

	log.Info("Effect4 code generation complete", logger.F("files", len(files)))
	return files, nil
}

func sortedGroupKeys(groups map[string][]models.Query) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (e *Effect4) groupQueriesByFile(queries []models.Query, log *logger.Logger) map[string][]models.Query {
	groups := make(map[string][]models.Query)
	for _, q := range queries {
		filename := q.Filename
		if filename == "" {
			filename = "queries.sql"
			log.Warn("Query has no filename, using default", logger.F("query", q.Name))
		}
		groups[filename] = append(groups[filename], q)
	}
	return groups
}

func (e *Effect4) filenameToRepoName(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	return toPascalCase(name) + "Repository"
}
