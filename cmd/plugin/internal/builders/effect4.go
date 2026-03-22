package builders

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"unicode"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/version"
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
	// Most logic is now pre-computed in QueryView, keeping only what's needed
	funcMap := template.FuncMap{
		// String transformations (still needed for enums in Catalog)
		"pascalCase": toPascalCase,

		// Utility functions for template rendering
		"splitLines": splitLines,
	}

	tmpl, err := template.New("repository").Funcs(funcMap).Parse(string(tmplContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl, nil
}

// SchemaField represents a single field in a Schema.Struct for templates
type SchemaField struct {
	Name   string // Field name (camelCase for params, original for results)
	Schema string // e.g., "Schema.Int", "Schema.optional(Schema.String)"
}

// QueryView is a pre-computed view of a Query for templates
// This moves computation logic from templates to Go for better readability
type QueryView struct {
	// Names in different cases
	Name       string // Original: "GetUser"
	NamePascal string // "GetUser"
	NameCamel  string // "getUser"

	// Command type (for helper flag computation)
	Command string // ":one", ":many", ":exec", ":execrows"

	// Flags
	HasParams  bool
	HasResults bool

	// Pre-computed template strings
	ReturnType      string // "void" | "number" | "Option.Option<GetUserResult>" | "GetUserResult[]"
	SqlSchemaMethod string // "SqlSchema.void" | "SqlSchema.findOneOption" | "SqlSchema.findAll" | "execRows"
	RequestSchema   string // "Schema.Void" | "GetUserParams"

	// Pre-computed schema fields for cleaner template iteration
	ParamFields  []SchemaField // For params schema
	ResultFields []SchemaField // For result schema

	// SQL-related fields
	OriginalSQL         string // The SQL (with explicit column aliases if needed)
	SQLTemplateLiteral  string // Transformed SQL with ${params.name} (default, unless disabled)
	ParamList           string // e.g., "params.id, params.name" for sql.unsafe
	UseTemplateLiterals bool   // True by default, false if disable_template_literals is set
}

type RepositoryData struct {
	RepositoryName      string
	RepositoryNameCamel string // e.g., "ordersRepository"
	Filename            string
	QueryViews          []QueryView
	Catalog             *models.Catalog
	Config              config.Config
	SqlcVersion         string
	PluginVersion       string

	// Helper flags - only include helpers that are actually used
	NeedsBigInt   bool // True if any field uses BigIntFromString
	NeedsExecRows bool // True if any query uses :execrows command
}

func (e *Effect4) generateRepository(tmpl *template.Template, filename string, queries []models.Query, catalog *models.Catalog, sqlcVersion string, log *logger.Logger) (string, error) {
	repoName := e.filenameToRepoName(filename)
	queryViews := e.buildQueryViews(queries, log)

	// Compute which helpers are needed
	needsBigInt, needsExecRows := e.computeHelperFlags(queryViews)

	data := RepositoryData{
		RepositoryName:      repoName,
		RepositoryNameCamel: toCamelCase(repoName),
		Filename:            filename,
		QueryViews:          queryViews,
		Catalog:             catalog,
		Config:              e.cfg,
		SqlcVersion:         sqlcVersion,
		PluginVersion:       version.Version,
		NeedsBigInt:         needsBigInt,
		NeedsExecRows:       needsExecRows,
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

// computeHelperFlags analyzes query views to determine which helpers are needed
func (e *Effect4) computeHelperFlags(views []QueryView) (needsBigInt, needsExecRows bool) {
	for _, v := range views {
		// Check if any query uses :execrows
		if v.Command == ":execrows" {
			needsExecRows = true
		}

		// Check if any schema field uses BigIntFromString
		for _, f := range v.ParamFields {
			if f.Schema == "BigIntFromString" || strings.Contains(f.Schema, "BigIntFromString") {
				needsBigInt = true
			}
		}
		for _, f := range v.ResultFields {
			if f.Schema == "BigIntFromString" || strings.Contains(f.Schema, "BigIntFromString") {
				needsBigInt = true
			}
		}

		// Early exit if both flags are set
		if needsBigInt && needsExecRows {
			return
		}
	}
	return
}

// buildQueryViews transforms a slice of queries into QueryViews with pre-computed values
func (e *Effect4) buildQueryViews(queries []models.Query, log *logger.Logger) []QueryView {
	views := make([]QueryView, len(queries))
	for i, q := range queries {
		views[i] = e.buildQueryView(q, log)
	}
	return views
}

// buildQueryView creates a QueryView with all pre-computed template values
func (e *Effect4) buildQueryView(q models.Query, log *logger.Logger) QueryView {
	namePascal := toPascalCase(q.Name)
	nameCamel := toCamelCase(q.Name)
	hasParams := len(q.Params) > 0
	hasResults := len(q.Results) > 0

	// Compute return type based on command
	var returnType string
	switch q.Command {
	case ":exec":
		returnType = "void"
	case ":execrows":
		returnType = "number"
	case ":one":
		returnType = fmt.Sprintf("Option.Option<%sResult>", namePascal)
	default: // :many
		returnType = fmt.Sprintf("%sResult[]", namePascal)
	}

	// Compute SqlSchema method
	var sqlSchemaMethod string
	switch q.Command {
	case ":exec":
		sqlSchemaMethod = "SqlSchema.void"
	case ":execrows":
		sqlSchemaMethod = "execRows"
	case ":one":
		sqlSchemaMethod = "SqlSchema.findOneOption"
	default:
		sqlSchemaMethod = "SqlSchema.findAll"
	}

	// Compute request schema
	requestSchema := "Schema.Void"
	if hasParams {
		requestSchema = namePascal + "Params"
	}

	// Template literals are enabled by default, disabled via config
	useTemplateLiterals := !e.cfg.DisableTemplateLiterals

	view := QueryView{
		Name:                q.Name,
		NamePascal:          namePascal,
		NameCamel:           nameCamel,
		Command:             q.Command,
		HasParams:           hasParams,
		HasResults:          hasResults,
		ReturnType:          returnType,
		SqlSchemaMethod:     sqlSchemaMethod,
		RequestSchema:       requestSchema,
		ParamFields:         e.buildParamFields(q.Params),
		ResultFields:        e.buildResultFields(q.Results),
		OriginalSQL:         q.RewrittenSQL,
		ParamList:           e.generateParamList(q.Params),
		UseTemplateLiterals: useTemplateLiterals,
	}

	// Generate template literal version if enabled (default)
	if useTemplateLiterals {
		if hasParams {
			transformer := &SQLTransformer{}
			result, err := transformer.Transform(q.RewrittenSQL, q.Params, log)
			if err != nil {
				log.Warn("Failed to transform SQL to template literal, falling back to original",
					logger.F("query", q.Name),
					logger.F("error", err.Error()))
				// On error, use original SQL as template literal
				view.SQLTemplateLiteral = q.RewrittenSQL
			} else {
				view.SQLTemplateLiteral = result.TemplateLiteral
				log.Debug("Successfully transformed SQL to template literal",
					logger.F("query", q.Name),
					logger.F("replacements", result.ReplacementsMade))
			}
		} else {
			// No params, just use the SQL directly
			view.SQLTemplateLiteral = q.RewrittenSQL
		}
	}

	return view
}

// buildParamFields converts query parameters to SchemaFields for template rendering
func (e *Effect4) buildParamFields(params []models.Param) []SchemaField {
	fields := make([]SchemaField, len(params))
	for i, p := range params {
		fields[i] = SchemaField{
			Name:   toCamelCase(p.Name),
			Schema: e.sqlTypeToEffectSchemaForParams(p.Type),
		}
	}
	return fields
}

// buildResultFields converts query results to SchemaFields for template rendering
func (e *Effect4) buildResultFields(results []models.ResultField) []SchemaField {
	fields := make([]SchemaField, len(results))
	for i, r := range results {
		fields[i] = SchemaField{
			Name:   r.Name, // Keep original column name (preserves snake_case)
			Schema: e.sqlTypeToEffectSchemaForResults(r.Type),
		}
	}
	return fields
}

// generateParamList generates the parameter array string for sql.unsafe (e.g., "params.id, params.name")
func (e *Effect4) generateParamList(params []models.Param) string {
	if len(params) == 0 {
		return ""
	}

	var parts []string
	for _, param := range params {
		paramName := toCamelCase(param.Name)
		parts = append(parts, fmt.Sprintf("params.%s", paramName))
	}

	return strings.Join(parts, ", ")
}

// sqlTypeToEffectSchemaBase converts internal SqlType to Effect Schema expression
// This returns the base schema without nullability handling
func (e *Effect4) sqlTypeToEffectSchemaBase(t models.SqlType) string {
	var baseSchema string

	switch strings.ToLower(t.Name) {
	// serials
	case "serial", "serial4":
		baseSchema = "Schema.Int"
	case "bigserial", "serial8":
		baseSchema = "BigIntFromString" // PostgreSQL returns bigint as string
	case "smallserial", "serial2":
		baseSchema = "Schema.Int"

	// ints
	case "integer", "int", "int4":
		baseSchema = "Schema.Int"
	case "bigint", "int8":
		baseSchema = "BigIntFromString" // PostgreSQL returns bigint as string
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

	return baseSchema
}

// sqlTypeToEffectSchemaForParams converts internal SqlType to Effect Schema expression for parameters
// Uses Schema.optional() for nullable params so callers can omit optional fields
func (e *Effect4) sqlTypeToEffectSchemaForParams(t models.SqlType) string {
	baseSchema := e.sqlTypeToEffectSchemaBase(t)

	// For parameters, use optional() so callers can omit nullable fields
	if t.IsNullable {
		baseSchema = fmt.Sprintf("Schema.optional(%s)", baseSchema)
	}

	return baseSchema
}

// sqlTypeToEffectSchemaForResults converts internal SqlType to Effect Schema expression for results
// Uses Schema.OptionFromNullOr() for nullable results - transforms null to Option.None
func (e *Effect4) sqlTypeToEffectSchemaForResults(t models.SqlType) string {
	baseSchema := e.sqlTypeToEffectSchemaBase(t)

	// For results, use OptionFromNullOr() to transform null to Option.None
	if t.IsNullable {
		baseSchema = fmt.Sprintf("Schema.OptionFromNullOr(%s)", baseSchema)
	}

	return baseSchema
}

// toPascalCase converts snake_case or camelCase to PascalCase
func toPascalCase(s string) string {
	// Handle special SQL names - split by underscore
	words := strings.Split(s, "_")
	for i, word := range words {
		if word == "" {
			continue
		}
		// Capitalize first letter of each word
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, "")
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

// splitLines splits a string into lines for template rendering
func splitLines(s string) []string {
	return strings.Split(s, "\n")
}

var _ Builder = (*Effect4)(nil)
