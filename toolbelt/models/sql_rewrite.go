package models

import (
	"fmt"
	"strings"
)

// RewriteSQLWithAliases replaces duplicate column references with explicit aliases.
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
