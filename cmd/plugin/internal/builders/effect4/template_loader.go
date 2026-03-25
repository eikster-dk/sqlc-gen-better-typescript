package effect4

import (
	"embed"
	"fmt"
	"text/template"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
)

//go:embed templates/*.gotmpl
var templates embed.FS

type templateSet struct {
	models     *template.Template
	request    *template.Template
	response   *template.Template
	repository *template.Template
}

func loadTemplates(log *logger.Logger) (*templateSet, error) {
	funcMap := template.FuncMap{
		"splitLines":    splitLines,
		"formatImports": formatImports,
	}

	parseTemplate := func(name, path string) (*template.Template, error) {
		content, err := templates.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s template file: %w", name, err)
		}

		tmpl, err := template.New(name).Funcs(funcMap).Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s template: %w", name, err)
		}

		return tmpl, nil
	}

	repositoryTmpl, err := parseTemplate("repository", "templates/repository.ts.gotmpl")
	if err != nil {
		return nil, err
	}

	modelsTmpl, err := parseTemplate("models", "templates/models.ts.gotmpl")
	if err != nil {
		return nil, err
	}

	requestTmpl, err := parseTemplate("request", "templates/request.ts.gotmpl")
	if err != nil {
		return nil, err
	}

	responseTmpl, err := parseTemplate("response", "templates/response.ts.gotmpl")
	if err != nil {
		return nil, err
	}

	log.Debug("Templates loaded")

	return &templateSet{
		models:     modelsTmpl,
		request:    requestTmpl,
		response:   responseTmpl,
		repository: repositoryTmpl,
	}, nil
}
