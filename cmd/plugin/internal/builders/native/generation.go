package native

import (
	"bytes"
	"embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/version"
)

//go:embed templates/*.gotmpl
var templateFiles embed.FS

var (
	reExcessiveNewlines  = regexp.MustCompile(`\n{3,}`)
	reTrailingWhitespace = regexp.MustCompile(`[ \t]+\n`)
)

// modelsData holds the data passed to the models.ts template.
type modelsData struct {
	SqlcVersion   string
	PluginVersion string
}

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

func (n *Native) generateModelsFileFromTemplates(tmpls *templateSet, catalog *models.Catalog, sqlcVersion string) (File, error) {
	_ = catalog // catalog will be used in later phases for enums etc.

	data := modelsData{
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
	}

	content, err := executeTemplate(tmpls.models, data)
	if err != nil {
		return File{}, fmt.Errorf("failed to render models template: %w", err)
	}

	return File{Name: "models.ts", Content: []byte(content)}, nil
}

func (n *Native) generateQueryFiles(fileStem string, queryViews []QueryView, tmpls *templateSet, sqlcVersion string) (File, File, File, error) {
	importExt := ""
	if n.cfg.ImportExtension != nil {
		importExt = *n.cfg.ImportExtension
	}

	requestsData := RequestsData{
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
		QueryViews:    queryViews,
	}
	requestsContent, err := executeTemplate(tmpls.requests, requestsData)
	if err != nil {
		return File{}, File{}, File{}, fmt.Errorf("failed to render requests template: %w", err)
	}
	requestsFile := File{Name: fileStem + "Requests.ts", Content: []byte(requestsContent)}

	responsesData := ResponsesData{
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
		QueryViews:    queryViews,
	}
	responsesContent, err := executeTemplate(tmpls.responses, responsesData)
	if err != nil {
		return File{}, File{}, File{}, fmt.Errorf("failed to render responses template: %w", err)
	}
	responsesFile := File{Name: fileStem + "Responses.ts", Content: []byte(responsesContent)}

	queriesData := QueriesData{
		FileStem:      fileStem,
		ImportExt:     importExt,
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
		QueryViews:    queryViews,
	}
	queriesContent, err := executeTemplate(tmpls.queries, queriesData)
	if err != nil {
		return File{}, File{}, File{}, fmt.Errorf("failed to render queries template: %w", err)
	}
	queriesFile := File{Name: fileStem + "Queries.ts", Content: []byte(queriesContent)}

	return requestsFile, responsesFile, queriesFile, nil
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
