package native

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

// toSQLComment prefixes each line of sql with "// ".
func toSQLComment(sql string) string {
	lines := strings.Split(sql, "\n")
	for i, line := range lines {
		lines[i] = "// " + line
	}
	return strings.Join(lines, "\n")
}

func (n *Native) groupQueriesByFile(queries []models.Query, log *logger.Logger) map[string][]models.Query {
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

func sortedGroupKeys(groups map[string][]models.Query) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// filenameToStem converts "customers.sql" -> "customers"
func filenameToStem(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func (n *Native) buildQueryViews(queries []models.Query, log *logger.Logger) []QueryView {
	views := make([]QueryView, len(queries))
	for i, q := range queries {
		views[i] = n.buildQueryView(q, log)
	}
	return views
}

func (n *Native) buildQueryView(q models.Query, log *logger.Logger) QueryView {
	namePascal := toPascalCase(q.Name)
	nameCamel := toCamelCase(q.Name)
	hasParams := len(q.Params) > 0
	hasResults := len(q.Results) > 0

	paramFields := n.buildParamFields(q.Params)
	resultFields := n.buildResultFields(q.Results)

	sql := q.RewrittenSQL
	if sql == "" {
		sql = q.SQL
	}

	return QueryView{
		Name:         q.Name,
		NamePascal:   namePascal,
		NameCamel:    nameCamel,
		Command:      q.Command,
		HasParams:    hasParams,
		HasResults:   hasResults,
		ParamFields:  paramFields,
		ResultFields: resultFields,
		SQL:          sql,
		SQLComment:   toSQLComment(sql),
		ParamList:    buildParamList(q.Params),
	}
}

func (n *Native) buildParamFields(params []models.Param) []ZodField {
	fields := make([]ZodField, len(params))
	for i, p := range params {
		fields[i] = ZodField{
			Name:   toCamelCase(p.Name),
			Schema: n.zodTypeForParam(p.Type),
		}
	}
	return fields
}

func (n *Native) buildResultFields(results []models.ResultField) []ZodField {
	fields := make([]ZodField, len(results))
	for i, r := range results {
		fields[i] = ZodField{
			Name:   r.Name,
			Schema: n.zodTypeForResult(r.Type),
		}
	}
	return fields
}

func buildParamList(params []models.Param) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = fmt.Sprintf("inputParsed.data.%s", toCamelCase(p.Name))
	}
	return strings.Join(parts, ", ")
}

func toPascalCase(s string) string {
	words := strings.Split(s, "_")
	for i, word := range words {
		if word == "" {
			continue
		}
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, "")
}

func toCamelCase(s string) string {
	pascal := toPascalCase(s)
	if pascal == "" {
		return ""
	}
	runes := []rune(pascal)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}
