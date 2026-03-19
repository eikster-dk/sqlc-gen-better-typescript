package models

import (
	"fmt"
	"strings"
)

// RewriteSQLWithAliases replaces aliased column references with explicit aliases
// e.g., sms_episodes.id -> sms_episodes.id AS sms_episodes_id
func RewriteSQLWithAliases(sql string, results []ResultField) string {
	for _, result := range results {
		if result.IsAliased {
			original := fmt.Sprintf("%s.%s", result.Table, result.OriginalName)
			replacement := fmt.Sprintf("%s.%s AS %s", result.Table, result.OriginalName, result.Name)
			sql = strings.Replace(sql, original, replacement, 1)
		}
	}
	return sql
}
