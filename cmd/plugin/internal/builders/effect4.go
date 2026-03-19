package builders

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/models"
)

//go:embed templates/**/*.gotmpl
var templates embed.FS

type Effect4 struct {
	cfg config.Config
}

// NewEffect4 creates a new Effect4 builder
func NewEffect4(cfg config.Config) *Effect4 {
	return &Effect4{cfg: cfg}
}

// Build implements builders.Builder
func (e *Effect4) Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger, sqlcVersion string) ([]File, error) {
	log.Info("Starting Effect4 code generation", logger.F("builder", "effect-v4-unstable"))
	log.Debug("Catalog info", logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)))

	// Load and parse template
	tmpl, err := e.loadTemplate(log)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	// Group queries by filename
	queryGroups := e.groupQueriesByFile(queries, log)

	var files []File

	for filename, fileQueries := range queryGroups {
		log.Info("Generating repository",
			logger.F("file", filename),
			logger.F("queries", len(fileQueries)))

		// Generate repository code for this file
		content, err := e.generateRepository(tmpl, filename, fileQueries, catalog, sqlcVersion, log)
		if err != nil {
			log.Error("Failed to generate repository", err, logger.F("file", filename))
			return nil, fmt.Errorf("failed to generate repository for %s: %w", filename, err)
		}

		// Repository name based on filename (without extension, PascalCase)
		repoName := e.filenameToRepoName(filename)
		outputFilename := fmt.Sprintf("%s.ts", repoName)

		file := File{
			Name:    outputFilename,
			Content: []byte(content),
		}
		files = append(files, file)

		log.Info("Generated repository file",
			logger.F("name", file.Name),
			logger.F("size", len(file.Content)))
	}

	log.Info("Effect4 code generation complete", logger.F("files", len(files)))

	return files, nil
}

// groupQueriesByFile groups queries by their source SQL filename
func (e *Effect4) groupQueriesByFile(queries []models.Query, log *logger.Logger) map[string][]models.Query {
	groups := make(map[string][]models.Query)

	for _, q := range queries {
		filename := q.Filename
		if filename == "" {
			// Fallback: use a default name if filename is empty
			filename = "queries.sql"
			log.Warn("Query has no filename, using default", logger.F("query", q.Name))
		}

		groups[filename] = append(groups[filename], q)
	}

	return groups
}

// filenameToRepoName converts a SQL filename to a repository name
// e.g., "queries.sql" -> "QueriesRepository"
// e.g., "user_queries.sql" -> "UserQueriesRepository"
func (e *Effect4) filenameToRepoName(filename string) string {
	// Get base name without extension
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	// Convert to PascalCase and add Repository suffix
	pascal := toPascalCase(name)
	return pascal + "Repository"
}

func (e *Effect4) loadTemplate(log *logger.Logger) (*template.Template, error) {
	tmplContent, err := templates.ReadFile("templates/effect4/repository.ts.gotmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	// Create function map for template
	funcMap := template.FuncMap{
		"pascalCase":          toPascalCase,
		"camelCase":           toCamelCase,
		"hasParams":           func(q models.Query) bool { return len(q.Params) > 0 },
		"hasResults":          func(q models.Query) bool { return len(q.Results) > 0 },
		"isOne":               func(q models.Query) bool { return q.Command == ":one" },
		"isMany":              func(q models.Query) bool { return q.Command == ":many" },
		"isExec":              func(q models.Query) bool { return q.Command == ":exec" },
		"paramsSchema":        e.generateParamsSchema,
		"resultSchema":        e.generateResultSchema,
		"sqlWithPlaceholders": e.generateSQLWithPlaceholders,
		"paramList":           e.generateParamList,
	}

	tmpl, err := template.New("repository").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

type RepositoryData struct {
	RepositoryName string
	Filename       string
	Queries        []models.Query
	Catalog        *models.Catalog
	Config         config.Config
	SqlcVersion    string
	PluginVersion  string
}

// PluginVersion is the version of this sqlc-effect plugin
const PluginVersion = "v0.1.0"

func (e *Effect4) generateRepository(tmpl *template.Template, filename string, queries []models.Query, catalog *models.Catalog, sqlcVersion string, log *logger.Logger) (string, error) {
	repoName := e.filenameToRepoName(filename)

	data := RepositoryData{
		RepositoryName: repoName,
		Filename:       filename,
		Queries:        queries,
		Catalog:        catalog,
		Config:         e.cfg,
		SqlcVersion:    sqlcVersion,
		PluginVersion:  PluginVersion,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	// Post-process to clean up excessive whitespace
	return cleanWhitespace(buf.String()), nil
}

// cleanWhitespace removes excessive blank lines (more than 2 consecutive)
// and ensures single blank lines between major sections
func cleanWhitespace(content string) string {
	// Replace 4 or more consecutive newlines with 2 newlines (preserve section breaks)
	re := regexp.MustCompile(`\n{4,}`)
	content = re.ReplaceAllString(content, "\n\n")

	// Replace trailing whitespace at end of lines
	re = regexp.MustCompile(`[ \t]+\n`)
	content = re.ReplaceAllString(content, "\n")

	return strings.TrimSpace(content) + "\n"
}

// generateParamsSchema generates Schema.Struct for parameters
func (e *Effect4) generateParamsSchema(query models.Query) string {
	if len(query.Params) == 0 {
		return "Schema.Struct({})"
	}

	var fields []string
	for _, param := range query.Params {
		fieldName := toCamelCase(param.Name)
		schema := e.sqlTypeToEffectSchema(param.Type)
		fields = append(fields, fmt.Sprintf("%s: %s", fieldName, schema))
	}

	return fmt.Sprintf("Schema.Struct({\n  %s\n})", strings.Join(fields, ",\n  "))
}

// generateResultSchema generates Schema.Struct for results
// Preserves original column names from SQL (no camelCase conversion)
func (e *Effect4) generateResultSchema(query models.Query) string {
	if len(query.Results) == 0 {
		return "Schema.Struct({})"
	}

	var fields []string
	for _, result := range query.Results {
		// Use original column name from SQL (preserves snake_case)
		fieldName := result.Name
		schema := e.sqlTypeToEffectSchema(result.Type)
		fields = append(fields, fmt.Sprintf("%s: %s", fieldName, schema))
	}

	return fmt.Sprintf("Schema.Struct({\n  %s\n})", strings.Join(fields, ",\n  "))
}

// generateSQLWithPlaceholders returns the SQL to use
// Uses RewrittenSQL if available (with explicit column aliases), otherwise original SQL
func (e *Effect4) generateSQLWithPlaceholders(query models.Query) string {
	// Use rewritten SQL if available (has explicit column aliases for duplicates)
	// Otherwise fall back to original SQL
	if query.RewrittenSQL != "" && query.RewrittenSQL != query.SQL {
		return query.RewrittenSQL
	}
	return query.SQL
}

// generateParamList generates the parameter array for sql.unsafe
func (e *Effect4) generateParamList(query models.Query) string {
	if len(query.Params) == 0 {
		return ""
	}

	var params []string
	for _, param := range query.Params {
		paramName := toCamelCase(param.Name)
		params = append(params, fmt.Sprintf("params.%s", paramName))
	}

	return strings.Join(params, ", ")
}

// sqlTypeToEffectSchema converts internal SqlType to Effect Schema expression
func (e *Effect4) sqlTypeToEffectSchema(t models.SqlType) string {
	var baseSchema string

	switch strings.ToLower(t.Name) {
	// serials
	case "serial", "serial4":
		baseSchema = "Schema.Int"
	case "bigserial", "serial8":
		baseSchema = "Schema.BigInt"
	case "smallserial", "serial2":
		baseSchema = "Schema.Int"

	// ints
	case "integer", "int", "int4":
		baseSchema = "Schema.Int"
	case "bigint", "int8":
		baseSchema = "Schema.BigInt"
	case "smallint", "int2":
		baseSchema = "Schema.Int"

	// floats
	case "float", "double precision", "float8":
		baseSchema = "Schema.Number"
	case "real", "float4":
		baseSchema = "Schema.Number"

	// numeric / money
	case "numeric":
		baseSchema = "Schema.String"
	case "money":
		baseSchema = "Schema.String"

	// boolean
	case "boolean", "bool":
		baseSchema = "Schema.Boolean"

	// json
	case "json", "jsonb":
		baseSchema = "Schema.Unknown"

	// bytes
	case "bytea", "blob":
		baseSchema = "Schema.Uint8Array"

	// dates/times
	case "date":
		baseSchema = "Schema.Date"
	case "time", "timetz":
		baseSchema = "Schema.String"
	case "timestamp", "timestamptz":
		baseSchema = "Schema.Date"
	case "interval":
		baseSchema = "Schema.String"

	// strings
	case "text", "varchar", "bpchar", "string", "citext":
		baseSchema = "Schema.String"

	// uuid
	case "uuid":
		baseSchema = "Schema.String"

	// network-ish
	case "inet", "cidr", "macaddr", "macaddr8":
		baseSchema = "Schema.String"

	// ltree family
	case "ltree", "lquery", "ltxtquery":
		baseSchema = "Schema.String"

	// Enums
	default:
		if t.IsEnum {
			baseSchema = fmt.Sprintf("%sSchema", toPascalCase(t.Name))
		} else {
			baseSchema = "Schema.String" // Default to String for unknown types
		}
	}

	// Handle arrays
	if t.IsArray {
		baseSchema = fmt.Sprintf("Schema.Array(%s)", baseSchema)
	}

	// Handle nullability
	if t.IsNullable {
		baseSchema = fmt.Sprintf("Schema.optional(%s)", baseSchema)
	}

	return baseSchema
}

// toPascalCase converts snake_case or camelCase to PascalCase
func toPascalCase(s string) string {
	// Handle special SQL names
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Title(s)
	s = strings.ReplaceAll(s, " ", "")

	return s
}

// toCamelCase converts snake_case to camelCase
func toCamelCase(s string) string {
	// Convert to Pascal first, then lowercase first letter
	pascal := toPascalCase(s)
	if pascal == "" {
		return ""
	}

	return strings.ToLower(pascal[:1]) + pascal[1:]
}

var _ Builder = (*Effect4)(nil)
