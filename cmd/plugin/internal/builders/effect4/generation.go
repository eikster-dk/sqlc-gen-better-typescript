package effect4

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/version"
)

func (e *Effect4) generateRepositoryFile(tmpl *template.Template, filename string, queryViews []QueryView, sqlcVersion string) (File, error) {
	repoName := e.filenameToRepoName(filename)

	data := RepositoryData{
		RepositoryName:      repoName,
		RepositoryNameCamel: toCamelCase(repoName),
		Filename:            filename,
		QueryViews:          queryViews,
		Imports:             e.buildRepositoryImports(repoName, queryViews),
		SqlcVersion:         sqlcVersion,
		PluginVersion:       version.Version,
	}

	content, err := executeTemplate(tmpl, data)
	if err != nil {
		return File{}, fmt.Errorf("failed to render repository template: %w", err)
	}

	return File{Name: fmt.Sprintf("%s.ts", repoName), Content: []byte(content)}, nil
}

func (e *Effect4) generateRequestFile(tmpl *template.Template, repoName string, queryViews []QueryView, sqlcVersion string) (File, error) {
	data := RequestData{
		RepositoryName: repoName,
		QueryViews:     queryViews,
		Imports:        e.buildRequestImports(queryViews),
		SqlcVersion:    sqlcVersion,
		PluginVersion:  version.Version,
	}

	content, err := executeTemplate(tmpl, data)
	if err != nil {
		return File{}, fmt.Errorf("failed to render request template: %w", err)
	}

	return File{Name: fmt.Sprintf("%sRequest.ts", repoName), Content: []byte(content)}, nil
}

func (e *Effect4) generateResponseFile(tmpl *template.Template, repoName string, queryViews []QueryView, sqlcVersion string) (File, error) {
	data := ResponseData{
		RepositoryName: repoName,
		QueryViews:     queryViews,
		Imports:        e.buildResponseImports(queryViews),
		SqlcVersion:    sqlcVersion,
		PluginVersion:  version.Version,
	}

	content, err := executeTemplate(tmpl, data)
	if err != nil {
		return File{}, fmt.Errorf("failed to render response template: %w", err)
	}

	return File{Name: fmt.Sprintf("%sResponse.ts", repoName), Content: []byte(content)}, nil
}

func (e *Effect4) generateModelsFile(tmpl *template.Template, catalog *models.Catalog, queryViewsByFile map[string][]QueryView, sqlcVersion string) (File, error) {
	needsBigInt, needsExecRows := e.computeGlobalHelpers(queryViewsByFile)
	usedEmbedTables := collectUsedEmbedTables(queryViewsByFile)

	data := ModelsData{
		Imports:       buildModelsImports(needsBigInt, needsExecRows),
		Enums:         buildEnumViews(catalog.Enums),
		TableRows:     e.buildTableRows(catalog, usedEmbedTables),
		NeedsBigInt:   needsBigInt,
		NeedsExecRows: needsExecRows,
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
	}

	content, err := executeTemplate(tmpl, data)
	if err != nil {
		return File{}, fmt.Errorf("failed to render models template: %w", err)
	}

	return File{Name: "models.ts", Content: []byte(content)}, nil
}

func (e *Effect4) computeGlobalHelpers(byFile map[string][]QueryView) (needsBigInt, needsExecRows bool) {
	for _, views := range byFile {
		for _, v := range views {
			if v.Command == ":execrows" {
				needsExecRows = true
			}
			for _, f := range v.ParamFields {
				if strings.Contains(f.Schema, "BigIntFromString") {
					needsBigInt = true
				}
			}
			for _, f := range v.ResultFields {
				if strings.Contains(f.Schema, "BigIntFromString") {
					needsBigInt = true
				}
			}
			for _, f := range v.RowFields {
				if strings.Contains(f.Schema, "BigIntFromString") {
					needsBigInt = true
				}
			}
			if needsBigInt && needsExecRows {
				return
			}
		}
	}
	return
}

func buildModelsImports(needsBigInt, needsExecRows bool) Imports {
	imports := Imports{}
	effectSymbols := []string{"Schema"}
	if needsExecRows {
		effectSymbols = append(effectSymbols, "Effect")
	}
	if needsBigInt {
		effectSymbols = append(effectSymbols, "SchemaGetter")
	}
	imports["effect"] = effectSymbols
	return imports
}

func buildEnumViews(enums []models.Enum) []EnumView {
	views := make([]EnumView, len(enums))
	for i, enum := range enums {
		values := make([]string, len(enum.Values))
		for j, value := range enum.Values {
			values[j] = value.Value
		}
		views[i] = EnumView{NamePascal: toPascalCase(enum.Name), Values: values}
	}
	sort.Slice(views, func(i, j int) bool { return views[i].NamePascal < views[j].NamePascal })
	return views
}

func collectUsedEmbedTables(byFile map[string][]QueryView) map[string]struct{} {
	tables := make(map[string]struct{})
	for _, views := range byFile {
		for _, query := range views {
			for _, group := range query.EmbedGroups {
				tables[group.TableName] = struct{}{}
			}
		}
	}
	return tables
}

func (e *Effect4) buildTableRows(catalog *models.Catalog, usedTables map[string]struct{}) []TableRowView {
	tableRows := make([]TableRowView, 0, len(usedTables))
	for _, table := range catalog.Tables {
		if _, ok := usedTables[table.Name]; !ok {
			continue
		}
		fields := make([]SchemaField, len(table.Columns))
		for i, column := range table.Columns {
			expr := e.sqlTypeToEffectSchemaForResults(column.Type)
			fields[i] = SchemaField{Name: column.Name, Schema: expr.Schema, ModelImports: expr.ModelImports}
		}
		tableRows = append(tableRows, TableRowView{NamePascal: toPascalCase(table.Name), Fields: fields})
	}
	sort.Slice(tableRows, func(i, j int) bool { return tableRows[i].NamePascal < tableRows[j].NamePascal })
	return tableRows
}

func (e *Effect4) buildRequestImports(queryViews []QueryView) Imports {
	imports := Imports{"effect": []string{"Schema"}}
	var modelSymbols []string
	for _, query := range queryViews {
		for _, field := range query.ParamFields {
			modelSymbols = append(modelSymbols, field.ModelImports...)
		}
	}
	if symbols := uniqueSorted(modelSymbols); len(symbols) > 0 {
		imports[e.localImportPath("./models")] = symbols
	}
	return imports
}

func (e *Effect4) buildResponseImports(queryViews []QueryView) Imports {
	imports := Imports{"effect": []string{"Schema"}}
	needsSchemaTransformation := false
	modelSymbols := []string{}

	for _, query := range queryViews {
		for _, field := range query.ResultFields {
			modelSymbols = append(modelSymbols, field.ModelImports...)
		}
		if query.HasEmbeds {
			needsSchemaTransformation = true
			for _, group := range query.EmbedGroups {
				modelSymbols = append(modelSymbols, group.RowSchema)
			}
			for _, field := range query.RowFields {
				modelSymbols = append(modelSymbols, field.ModelImports...)
			}
		}
	}

	if needsSchemaTransformation {
		imports["effect"] = append(imports["effect"], "SchemaTransformation")
	}
	if symbols := uniqueSorted(modelSymbols); len(symbols) > 0 {
		imports[e.localImportPath("./models")] = symbols
	}
	return imports
}

func (e *Effect4) buildRepositoryImports(repoName string, queryViews []QueryView) Imports {
	imports := Imports{
		"effect":              []string{"ServiceMap", "Effect", "Layer", "Schema", "Option"},
		"effect/unstable/sql": []string{"SqlClient", "SqlError", "SqlSchema"},
	}

	requestSymbols := make([]string, 0, len(queryViews))
	responseSymbols := make([]string, 0, len(queryViews))
	needsExecRows := false
	for _, query := range queryViews {
		if query.HasParams {
			requestSymbols = append(requestSymbols, query.NamePascal+"Params")
		}
		if query.HasResults {
			responseSymbols = append(responseSymbols, query.NamePascal+"Result")
		}
		if query.Command == ":execrows" {
			needsExecRows = true
		}
	}

	if symbols := uniqueSorted(requestSymbols); len(symbols) > 0 {
		imports[e.localImportPath("./"+repoName+"Request")] = symbols
	}
	if symbols := uniqueSorted(responseSymbols); len(symbols) > 0 {
		imports[e.localImportPath("./"+repoName+"Response")] = symbols
	}
	if needsExecRows {
		imports[e.localImportPath("./models")] = []string{"execRows"}
	}
	return imports
}

func (e *Effect4) localImportPath(path string) string {
	ext := ""
	if e.cfg.ImportExtension != nil {
		ext = *e.cfg.ImportExtension
	}
	if ext == "" {
		return path
	}
	if !strings.HasPrefix(path, "./") {
		return path
	}
	if strings.HasSuffix(path, ".js") || strings.HasSuffix(path, ".ts") {
		return path
	}
	return path + ext
}
