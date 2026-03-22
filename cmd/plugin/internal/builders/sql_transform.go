package builders

import (
	"fmt"
	"sort"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

// TransformResult contains the transformed SQL and validation info
type TransformResult struct {
	OriginalSQL      string
	TemplateLiteral  string
	ReplacementsMade int
}

// SQLTransformer converts SQL with $N placeholders to template literal format
// using simple string replacement (no regex)
type SQLTransformer struct{}

// Transform converts a SQL query with $N placeholders to template literal format
// It processes placeholders from highest to lowest position to avoid any edge cases
func (st *SQLTransformer) Transform(sql string, params []models.Param, log *logger.Logger) (TransformResult, error) {
	log.Debug("Starting SQL transformation",
		logger.F("original_sql", sql),
		logger.F("param_count", len(params)))

	transformed := sql
	totalReplacements := 0

	// Sort params by position descending (highest first)
	// IMPORTANT: This is required, not optional. We must replace $10 before $1,
	// otherwise $1 would match as a substring inside $10, $11, $12, etc.
	// Example: "$10" would become "${params.a}0" if we replaced $1 first.
	sortedParams := make([]models.Param, len(params))
	copy(sortedParams, params)
	sort.Slice(sortedParams, func(i, j int) bool {
		return sortedParams[i].Position > sortedParams[j].Position
	})

	for _, param := range sortedParams {
		placeholder := fmt.Sprintf("$%d", param.Position)
		paramRef := fmt.Sprintf("${params.%s}", toCamelCase(param.Name))

		// Count occurrences before replacement for validation
		count := strings.Count(transformed, placeholder)
		if count == 0 {
			log.Warn("Placeholder not found in SQL",
				logger.F("placeholder", placeholder),
				logger.F("param_name", param.Name))
			continue
		}

		// Simple string replacement - no regex
		transformed = strings.ReplaceAll(transformed, placeholder, paramRef)
		totalReplacements += count

		log.Debug("Replaced placeholder",
			logger.F("placeholder", placeholder),
			logger.F("with", paramRef),
			logger.F("occurrences", count))
	}

	// Validation: ensure we made at least as many replacements as params
	// (some params might be used multiple times, hence "at least")
	if totalReplacements < len(params) {
		return TransformResult{}, fmt.Errorf(
			"SQL transformation validation failed: expected at least %d replacements but made %d",
			len(params), totalReplacements)
	}

	log.Debug("SQL transformation complete",
		logger.F("transformed_sql", transformed),
		logger.F("total_replacements", totalReplacements))

	return TransformResult{
		OriginalSQL:      sql,
		TemplateLiteral:  transformed,
		ReplacementsMade: totalReplacements,
	}, nil
}
