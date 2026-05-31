package native

import (
	"bytes"
	"embed"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/native/internal/version"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
)

//go:embed templates/*.gotmpl
var templateFiles embed.FS

var (
	reExcessiveNewlines  = regexp.MustCompile(`\n{3,}`)
	reTrailingWhitespace = regexp.MustCompile(`[ \t]+\n`)
)

// templateSet holds all loaded templates.
type templateSet struct {
	models    *template.Template
	requests  *template.Template
	responses *template.Template
	queries   *template.Template
}

func loadAllTemplates() (*templateSet, error) {
	load := func(name, path string) (*template.Template, error) {
		content, err := templateFiles.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s template: %w", name, err)
		}
		tmpl, err := template.New(name).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s template: %w", name, err)
		}
		return tmpl, nil
	}

	modelsTmpl, err := load("models", "templates/models.ts.gotmpl")
	if err != nil {
		return nil, err
	}
	requestsTmpl, err := load("requests", "templates/requests.ts.gotmpl")
	if err != nil {
		return nil, err
	}
	responsesTmpl, err := load("responses", "templates/responses.ts.gotmpl")
	if err != nil {
		return nil, err
	}
	queriesTmpl, err := load("queries", "templates/queries.ts.gotmpl")
	if err != nil {
		return nil, err
	}

	return &templateSet{
		models:    modelsTmpl,
		requests:  requestsTmpl,
		responses: responsesTmpl,
		queries:   queriesTmpl,
	}, nil
}

func (n *Native) generateModelsFileFromTemplates(tmpls *templateSet, catalog *models.Catalog, usedEmbedTables map[string]struct{}, sqlcVersion string) (toolbelt.File, error) {
	data := ModelsData{
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
		Enums:         buildEnumViews(catalog.Enums),
		TableRows:     n.buildTableRows(catalog, usedEmbedTables),
	}

	content, err := executeTemplate(tmpls.models, data)
	if err != nil {
		return toolbelt.File{}, fmt.Errorf("failed to render models template: %w", err)
	}

	return toolbelt.File{Name: "models.ts", Content: []byte(content)}, nil
}

func (n *Native) buildTableRows(catalog *models.Catalog, usedTables map[string]struct{}) []TableRowView {
	tableRows := make([]TableRowView, 0, len(usedTables))
	for _, table := range catalog.Tables {
		if _, ok := usedTables[table.Name]; !ok {
			continue
		}
		fields := make([]ZodField, len(table.Columns))
		for i, column := range table.Columns {
			typeInfo := column.Type
			typeInfo.IsNullable = column.Nullable
			fields[i] = ZodField{Name: column.Name, Schema: n.zodTypeForResult(typeInfo), EnumImport: n.enumImport(typeInfo)}
		}
		tableRows = append(tableRows, TableRowView{NamePascal: toPascalCase(table.Name), Fields: fields})
	}
	sort.Slice(tableRows, func(i, j int) bool { return tableRows[i].NamePascal < tableRows[j].NamePascal })
	return tableRows
}

func (n *Native) generateQueryFiles(fileStem string, queryViews []QueryView, tmpls *templateSet, sqlcVersion string) (toolbelt.File, toolbelt.File, toolbelt.File, error) {
	importExt := n.cfg.ImportExtension

	requestsData := RequestsData{
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
		QueryViews:    queryViews,
		EnumImports:   queryEnumImports(queryViews),
		ImportExt:     importExt,
	}
	requestsContent, err := executeTemplate(tmpls.requests, requestsData)
	if err != nil {
		return toolbelt.File{}, toolbelt.File{}, toolbelt.File{}, fmt.Errorf("failed to render requests template: %w", err)
	}
	requestsFile := toolbelt.File{Name: fileStem + "Requests.ts", Content: []byte(requestsContent)}

	responsesData := ResponsesData{
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
		QueryViews:    queryViews,
		EnumImports:   queryEnumImports(queryViews),
		ImportExt:     importExt,
	}
	responsesContent, err := executeTemplate(tmpls.responses, responsesData)
	if err != nil {
		return toolbelt.File{}, toolbelt.File{}, toolbelt.File{}, fmt.Errorf("failed to render responses template: %w", err)
	}
	responsesFile := toolbelt.File{Name: fileStem + "Responses.ts", Content: []byte(responsesContent)}

	queriesData := QueriesData{
		FileStem:        fileStem,
		ImportExt:       importExt,
		NeedsExecResult: needsExecResult(queryViews),
		NeedsSlices:     needsSlices(queryViews),
		SqlcVersion:     sqlcVersion,
		PluginVersion:   version.Version,
		QueryViews:      queryViews,
	}
	queriesContent, err := executeTemplate(tmpls.queries, queriesData)
	if err != nil {
		return toolbelt.File{}, toolbelt.File{}, toolbelt.File{}, fmt.Errorf("failed to render queries template: %w", err)
	}
	queriesFile := toolbelt.File{Name: fileStem + "Queries.ts", Content: []byte(queriesContent)}

	return requestsFile, responsesFile, queriesFile, nil
}

func buildEnumViews(enums []models.Enum) []EnumView {
	views := make([]EnumView, len(enums))
	for i, enum := range enums {
		values := make([]string, len(enum.Values))
		for j, value := range enum.Values {
			values[j] = value.Value
		}
		views[i] = EnumView{NamePascal: toPascalCase(enum.Name), Schema: zodEnumUnion(values)}
	}
	sort.Slice(views, func(i, j int) bool { return views[i].NamePascal < views[j].NamePascal })
	return views
}

func queryEnumImports(queryViews []QueryView) []string {
	imports := make([]string, 0)
	for _, query := range queryViews {
		imports = append(imports, query.EnumImports...)
	}
	return uniqueSorted(imports)
}

func needsExecResult(queryViews []QueryView) bool {
	for _, query := range queryViews {
		if query.Command == ":execresult" {
			return true
		}
	}
	return false
}

func needsSlices(queryViews []QueryView) bool {
	for _, query := range queryViews {
		if query.HasSlices {
			return true
		}
	}
	return false
}

func executeTemplate(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return cleanWhitespace(buf.String()), nil
}

func cleanWhitespace(content string) string {
	content = reExcessiveNewlines.ReplaceAllString(content, "\n\n")
	content = reTrailingWhitespace.ReplaceAllString(content, "\n")

	return strings.TrimSpace(content) + "\n"
}
