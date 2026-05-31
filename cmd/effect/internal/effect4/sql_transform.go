package effect4

import (
	"fmt"
	"regexp"
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
		if param.Slice {
			paramName := toCamelCase(param.Name)
			var replacements int
			transformed, replacements = replaceSliceInClause(transformed, placeholder, paramName)
			if replacements > 0 {
				totalReplacements += replacements
				continue
			}
			paramRef = fmt.Sprintf("${sql.in(params.%s)}", paramName)
		}

		if param.Slice {
			wrappedPlaceholder := fmt.Sprintf("(%s)", placeholder)
			count := strings.Count(transformed, wrappedPlaceholder)
			if count > 0 {
				transformed = strings.ReplaceAll(transformed, wrappedPlaceholder, paramRef)
				totalReplacements += count
				continue
			}
		}

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

func replaceSliceInClause(sql, placeholder, paramName string) (string, int) {
	re := regexp.MustCompile(`(?i)([A-Za-z_][A-Za-z0-9_]*)\s+IN\s+\(` + regexp.QuoteMeta(placeholder) + `\)`)
	replacements := 0
	result := re.ReplaceAllStringFunc(sql, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) != 2 {
			return match
		}
		replacements++
		return fmt.Sprintf(`${sql.in(%q, params.%s)}`, submatches[1], paramName)
	})
	return result, replacements
}
