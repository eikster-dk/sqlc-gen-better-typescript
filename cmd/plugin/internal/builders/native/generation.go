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

// modelsData holds the data passed to the models.ts template.
type modelsData struct {
	SqlcVersion   string
	PluginVersion string
}

func loadModelsTemplate() (*template.Template, error) {
	content, err := templateFiles.ReadFile("templates/models.ts.gotmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read models template: %w", err)
	}

	tmpl, err := template.New("models").Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse models template: %w", err)
	}

	return tmpl, nil
}

func (n *Native) generateModelsFile(catalog *models.Catalog, sqlcVersion string) (File, error) {
	_ = catalog // catalog will be used in later phases for enums etc.

	tmpl, err := loadModelsTemplate()
	if err != nil {
		return File{}, fmt.Errorf("failed to load models template: %w", err)
	}

	data := modelsData{
		SqlcVersion:   sqlcVersion,
		PluginVersion: version.Version,
	}

	content, err := executeTemplate(tmpl, data)
	if err != nil {
		return File{}, fmt.Errorf("failed to render models template: %w", err)
	}

	return File{Name: "models.ts", Content: []byte(content)}, nil
}

func executeTemplate(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return cleanWhitespace(buf.String()), nil
}

func cleanWhitespace(content string) string {
	re := regexp.MustCompile(`\n{3,}`)
	content = re.ReplaceAllString(content, "\n\n")

	re = regexp.MustCompile(`[ \t]+\n`)
	content = re.ReplaceAllString(content, "\n")

	return strings.TrimSpace(content) + "\n"
}
