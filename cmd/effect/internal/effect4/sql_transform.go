package effect4

import (
	"fmt"
	"sort"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
)

type TransformResult struct {
	OriginalSQL      string
	TemplateLiteral  string
	ReplacementsMade int
}

type SQLTransformer struct{}

func (st *SQLTransformer) Transform(sql string, params []models.Param, log *logger.Logger) (TransformResult, error) {
	log.Debug("Starting SQL transformation", logger.F("original_sql", sql), logger.F("param_count", len(params)))
	transformed := sql
	totalReplacements := 0
	sortedParams := make([]models.Param, len(params))
	copy(sortedParams, params)
	sort.Slice(sortedParams, func(i, j int) bool { return sortedParams[i].Position > sortedParams[j].Position })

	for _, param := range sortedParams {
		placeholder := fmt.Sprintf("$%d", param.Position)
		paramRef := fmt.Sprintf("${params.%s}", toCamelCase(param.Name))
		count := strings.Count(transformed, placeholder)
		if count == 0 {
			log.Warn("Placeholder not found in SQL", logger.F("placeholder", placeholder), logger.F("param_name", param.Name))
			continue
		}
		transformed = strings.ReplaceAll(transformed, placeholder, paramRef)
		totalReplacements += count
	}

	if totalReplacements < len(params) {
		return TransformResult{}, fmt.Errorf("SQL transformation validation failed: expected at least %d replacements but made %d", len(params), totalReplacements)
	}
	return TransformResult{OriginalSQL: sql, TemplateLiteral: transformed, ReplacementsMade: totalReplacements}, nil
}
