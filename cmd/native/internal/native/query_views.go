package native

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
	"github.com/jinzhu/inflection"
)

// toSQLComment prefixes each line of sql with "// ".
func toSQLComment(sql string) string {
	lines := strings.Split(sql, "\n")
	for i, line := range lines {
		lines[i] = "// " + line
	}
	return strings.Join(lines, "\n")
}

func (n *Native) groupQueriesByFile(plans []models.QueryPlan, log *logger.Logger) map[string][]models.QueryPlan {
	groups := make(map[string][]models.QueryPlan)
	for _, plan := range plans {
		filename := plan.Filename
		if filename == "" {
			filename = "queries.sql"
			log.Warn("Query has no filename, using default", logger.F("query", plan.Name))
		}
		groups[filename] = append(groups[filename], plan)
	}
	return groups
}

func sortedGroupKeys(groups map[string][]models.QueryPlan) []string {
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

func (n *Native) buildQueryViews(plans []models.QueryPlan, log *logger.Logger) []QueryView {
	views := make([]QueryView, len(plans))
	for i, plan := range plans {
		views[i] = n.buildQueryView(plan, log)
	}
	return views
}

func (n *Native) buildQueryView(plan models.QueryPlan, log *logger.Logger) QueryView {
	_ = log
	namePascal := toPascalCase(plan.Name)
	nameCamel := toCamelCase(plan.Name)
	hasParams := len(plan.Request.Fields) > 0
	hasResults := len(plan.Response.Row.Fields) > 0

	paramFields := n.buildParamFields(plan.Request.Fields)
	resultFields := n.buildResultFields(plan.Response.Row.Fields)
	resultMappings := buildResultMappings(plan.Response.Result.Shape.Fields)
	sql := plan.Source.ExecSQL

	return QueryView{
		Name:              plan.Name,
		NamePascal:        namePascal,
		NameCamel:         nameCamel,
		Command:           plan.Command,
		HasParams:         hasParams,
		HasResults:        hasResults,
		HasResultMappings: hasResults,
		ParamFields:       paramFields,
		ResultFields:      resultFields,
		ResultMappings:    resultMappings,
		SQL:               fmt.Sprintf("%q", sql),
		SQLComment:        toSQLComment(sql),
		ParamList:         buildParamList(plan.Source.Parameters),
		QueryParamList:    buildQueryParamList(plan.Source.Parameters),
		HasSlices:         plan.Features.UsesSlices,
		EnumImports:       queryEnumImportsFromFields(paramFields, resultFields, resultMappings),
	}
}

func queryEnumImportsFromFields(paramFields, resultFields []ZodField, resultMappings []ResultMapping) []string {
	return appendUniqueSorted(enumImports(paramFields), enumImports(resultFields), resultMappingSchemas(resultMappings))
}

func enumImports(fields []ZodField) []string {
	imports := make([]string, 0, len(fields))
	for _, field := range fields {
		if field.EnumImport != "" {
			imports = append(imports, field.EnumImport)
		}
	}
	return imports
}

func resultMappingSchemas(mappings []ResultMapping) []string {
	imports := make([]string, 0, len(mappings))
	for _, mapping := range mappings {
		if mapping.Schema != "" {
			imports = append(imports, mapping.Schema)
		}
		imports = append(imports, resultMappingSchemas(mapping.Object)...)
	}
	return imports
}

func buildResultMappings(fields []models.ResultShapeField) []ResultMapping {
	return buildResultMappingsWithIndent(fields, "  ")
}

func buildResultMappingsWithIndent(fields []models.ResultShapeField, indent string) []ResultMapping {
	mappings := make([]ResultMapping, 0, len(fields))
	for i, field := range fields {
		mapping := buildResultMapping(field, indent)
		mapping.IsLast = i == len(fields)-1
		mappings = append(mappings, mapping)
	}
	return mappings
}

func buildResultMapping(field models.ResultShapeField, indent string) ResultMapping {
	if field.Kind == models.ResultShapeFieldObject && field.Object != nil {
		return ResultMapping{Name: toCamelCase(singular(field.Name)), Schema: toPascalCase(field.Name) + "Row", Object: buildResultMappingsWithIndent(field.Object.Fields, indent+"  "), IsObject: true, Indent: indent}
	}
	rowField := field.Name
	if field.Source != nil {
		rowField = field.Source.Name
	}
	return ResultMapping{Name: field.Name, RowField: rowField, Indent: indent}
}

func singular(name string) string {
	lower := strings.ToLower(name)
	switch lower {
	case "campus", "meta", "metadata":
		return name
	case "calories":
		return "calorie"
	case "waves":
		return "wave"
	}
	return inflection.Singular(name)
}

func (n *Native) buildParamFields(params []models.RequestField) []ZodField {
	fields := make([]ZodField, len(params))
	for i, p := range params {
		fields[i] = ZodField{
			Name:       toCamelCase(p.Name),
			Schema:     n.zodTypeForParam(p.Type),
			EnumImport: n.enumImport(p.Type),
		}
	}
	return fields
}

func (n *Native) buildResultFields(results []models.RowField) []ZodField {
	fields := make([]ZodField, len(results))
	for i, r := range results {
		fields[i] = ZodField{
			Name:       r.Name,
			Schema:     n.zodTypeForResult(r.Type),
			EnumImport: n.enumImport(r.Type),
		}
	}
	return fields
}

func (n *Native) enumImport(t models.SqlType) string {
	if !t.IsEnum {
		return ""
	}
	if _, ok := n.enumValues[t.Name]; !ok {
		return ""
	}
	return toPascalCase(t.Name)
}

func buildParamList(params []models.SQLParameter) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = fmt.Sprintf("inputParsed.data.%s", toCamelCase(p.FieldName))
	}
	return strings.Join(parts, ", ")
}

func buildQueryParamList(params []models.SQLParameter) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, len(params))
	for i, p := range params {
		parts[i] = fmt.Sprintf("{ value: inputParsed.data.%s, slice: %t }", toCamelCase(p.FieldName), p.Slice)
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
