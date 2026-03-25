package builders

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/version"
	"github.com/jinzhu/inflection"
)

//go:embed templates/**/*.gotmpl
var templates embed.FS

type Effect4 struct {
	cfg config.Config
}

type Imports map[string][]string

type templateSet struct {
	models     *template.Template
	request    *template.Template
	response   *template.Template
	repository *template.Template
}

// NewEffect4 creates a new Effect4 builder
func NewEffect4(cfg config.Config) *Effect4 {
	return &Effect4{cfg: cfg}
}

// Build implements builders.Builder
func (e *Effect4) Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger, sqlcVersion string) ([]File, error) {
	log.Info("Starting Effect4 code generation", logger.F("builder", "effect-v4-unstable"))
	log.Debug("Catalog info", logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)))

	// Load and parse templates
	tmpls, err := e.loadTemplates(log)
	if err != nil {
		return nil, fmt.Errorf("failed to load templates: %w", err)
	}

	// Group queries by filename
	queryGroups := e.groupQueriesByFile(queries, log)
	filenames := sortedGroupKeys(queryGroups)

	queryViewsByFile := make(map[string][]QueryView, len(queryGroups))
	for _, filename := range filenames {
		queryViewsByFile[filename] = e.buildQueryViews(queryGroups[filename], log)
	}

	modelsFile, err := e.generateModelsFile(tmpls.models, catalog, queryViewsByFile, sqlcVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to generate models file: %w", err)
	}

	files := []File{modelsFile}

	for _, filename := range filenames {
		fileQueries := queryGroups[filename]
		queryViews := queryViewsByFile[filename]
		repoName := e.filenameToRepoName(filename)

		log.Info("Generating repository",
			logger.F("file", filename),
			logger.F("queries", len(fileQueries)))

		requestFile, err := e.generateRequestFile(tmpls.request, repoName, queryViews, sqlcVersion)
		if err != nil {
			log.Error("Failed to generate request file", err, logger.F("file", filename))
			return nil, fmt.Errorf("failed to generate request file for %s: %w", filename, err)
		}

		responseFile, err := e.generateResponseFile(tmpls.response, repoName, queryViews, sqlcVersion)
		if err != nil {
			log.Error("Failed to generate response file", err, logger.F("file", filename))
			return nil, fmt.Errorf("failed to generate response file for %s: %w", filename, err)
		}

		repositoryFile, err := e.generateRepositoryFile(tmpls.repository, filename, queryViews, sqlcVersion)
		if err != nil {
			log.Error("Failed to generate repository", err, logger.F("file", filename))
			return nil, fmt.Errorf("failed to generate repository for %s: %w", filename, err)
		}

		files = append(files, requestFile, responseFile, repositoryFile)

		log.Info("Generated repository files",
			logger.F("repository", repositoryFile.Name),
			logger.F("request", requestFile.Name),
			logger.F("response", responseFile.Name))
	}

	log.Info("Effect4 code generation complete", logger.F("files", len(files)))

	return files, nil
}

func sortedGroupKeys(groups map[string][]models.Query) []string {
	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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

func (e *Effect4) loadTemplates(log *logger.Logger) (*templateSet, error) {
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

	repositoryTmpl, err := parseTemplate("repository", "templates/effect4/repository.ts.gotmpl")
	if err != nil {
		return nil, err
	}

	modelsTmpl, err := parseTemplate("models", "templates/effect4/models.ts.gotmpl")
	if err != nil {
		return nil, err
	}

	requestTmpl, err := parseTemplate("request", "templates/effect4/request.ts.gotmpl")
	if err != nil {
		return nil, err
	}

	responseTmpl, err := parseTemplate("response", "templates/effect4/response.ts.gotmpl")
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

func formatImports(imports Imports) string {
	if len(imports) == 0 {
		return ""
	}

	modules := make([]string, 0, len(imports))
	for mod := range imports {
		modules = append(modules, mod)
	}
	sort.Strings(modules)

	lines := make([]string, 0, len(modules))
	for _, mod := range modules {
		symbols := uniqueSorted(imports[mod])
		if len(symbols) == 0 {
			continue
		}

		lines = append(lines, fmt.Sprintf(`import { %s } from "%s"`, strings.Join(symbols, ", "), mod))
	}

	return strings.Join(lines, "\n")
}

func uniqueSorted(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		set[value] = struct{}{}
	}

	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)

	return result
}

// SchemaField represents a single field in a Schema.Struct for templates
type SchemaField struct {
	Name   string // Field name (camelCase for params, original for results)
	Schema string // e.g., "Schema.Int", "Schema.optional(Schema.String)"
}

type EnumView struct {
	NamePascal string
	Values     []string
}

type TableRowView struct {
	NamePascal string
	Fields     []SchemaField
}

// EmbedGroupView represents an embed group for template rendering
type EmbedGroupView struct {
	TableName    string        // Original table name, e.g., "orders"
	FieldName    string        // Singularized field name, e.g., "order"
	RowSchema    string        // Shared row schema name, e.g., "OrdersRow"
	SchemaName   string        // Legacy/internal name (kept for compatibility)
	Fields       []SchemaField // Fields in this embed group
	FieldMapping []FieldMap    // Maps row fields to embed fields
}

// FieldMap represents a single field mapping in the transform
type FieldMap struct {
	RowFieldName   string // e.g., "customers_id"
	EmbedFieldName string // e.g., "id"
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
	HasEmbeds  bool // True if this query uses sqlc.embed

	// Pre-computed template strings
	ReturnType      string // "void" | "number" | "Option.Option<GetUserResult>" | "GetUserResult[]"
	SqlSchemaMethod string // "SqlSchema.void" | "SqlSchema.findOneOption" | "SqlSchema.findAll" | "execRows"
	RequestSchema   string // "Schema.Void" | "GetUserParams"

	// Pre-computed schema fields for cleaner template iteration
	ParamFields  []SchemaField // For params schema
	ResultFields []SchemaField // For result schema

	// Embed-related fields
	EmbedGroups []EmbedGroupView // Embed groups for this query
	RowFields   []SchemaField    // All fields for the row schema (when HasEmbeds)

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
	Imports             Imports
	QueryViews          []QueryView
	SqlcVersion         string
	PluginVersion       string
}

type RequestData struct {
	RepositoryName string
	Imports        Imports
	QueryViews     []QueryView
	SqlcVersion    string
	PluginVersion  string
}

type ResponseData struct {
	RepositoryName string
	Imports        Imports
	QueryViews     []QueryView
	SqlcVersion    string
	PluginVersion  string
}

type ModelsData struct {
	Imports       Imports
	Enums         []EnumView
	TableRows     []TableRowView
	NeedsBigInt   bool
	NeedsExecRows bool
	SqlcVersion   string
	PluginVersion string
}

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

	return File{
		Name:    fmt.Sprintf("%s.ts", repoName),
		Content: []byte(content),
	}, nil
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

	return File{
		Name:    fmt.Sprintf("%sRequest.ts", repoName),
		Content: []byte(content),
	}, nil
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

	return File{
		Name:    fmt.Sprintf("%sResponse.ts", repoName),
		Content: []byte(content),
	}, nil
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

	return File{
		Name:    "models.ts",
		Content: []byte(content),
	}, nil
}

func executeTemplate(tmpl *template.Template, data any) (string, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return cleanWhitespace(buf.String()), nil
}

// cleanWhitespace removes excessive blank lines (more than 2 consecutive)
// and ensures single blank lines between major sections
func cleanWhitespace(content string) string {
	// Replace 3 or more consecutive newlines with 2 newlines (preserve section breaks)
	re := regexp.MustCompile(`\n{3,}`)
	content = re.ReplaceAllString(content, "\n\n")

	// Replace trailing whitespace at end of lines
	re = regexp.MustCompile(`[ \t]+\n`)
	content = re.ReplaceAllString(content, "\n")

	return strings.TrimSpace(content) + "\n"
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
		views[i] = EnumView{
			NamePascal: toPascalCase(enum.Name),
			Values:     values,
		}
	}

	sort.Slice(views, func(i, j int) bool {
		return views[i].NamePascal < views[j].NamePascal
	})

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
			fields[i] = SchemaField{
				Name:   column.Name,
				Schema: e.sqlTypeToEffectSchemaForResults(column.Type),
			}
		}

		tableRows = append(tableRows, TableRowView{
			NamePascal: toPascalCase(table.Name),
			Fields:     fields,
		})
	}

	sort.Slice(tableRows, func(i, j int) bool {
		return tableRows[i].NamePascal < tableRows[j].NamePascal
	})

	return tableRows
}

func extractSchemaReferences(schema string) []string {
	var refs []string
	tokens := strings.FieldsFunc(schema, func(r rune) bool {
		return !(r == '_' || (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z'))
	})

	for _, token := range tokens {
		if strings.HasSuffix(token, "Schema") && token != "Schema" {
			if strings.HasPrefix(token, "Schema") {
				continue
			}
			refs = append(refs, token)
		}
	}

	return uniqueSorted(refs)
}

func (e *Effect4) buildRequestImports(queryViews []QueryView) Imports {
	imports := Imports{
		"effect": []string{"Schema"},
	}

	var modelSymbols []string
	for _, query := range queryViews {
		for _, field := range query.ParamFields {
			modelSymbols = append(modelSymbols, extractSchemaReferences(field.Schema)...)
		}
	}

	if symbols := uniqueSorted(modelSymbols); len(symbols) > 0 {
		imports[e.localImportPath("./models")] = symbols
	}

	return imports
}

func (e *Effect4) buildResponseImports(queryViews []QueryView) Imports {
	imports := Imports{
		"effect": []string{"Schema"},
	}

	needsSchemaTransformation := false
	modelSymbols := []string{}

	for _, query := range queryViews {
		for _, field := range query.ResultFields {
			if strings.Contains(field.Schema, "BigIntFromString") {
				modelSymbols = append(modelSymbols, "BigIntFromString")
			}
			modelSymbols = append(modelSymbols, extractSchemaReferences(field.Schema)...)
		}

		if query.HasEmbeds {
			needsSchemaTransformation = true
			for _, group := range query.EmbedGroups {
				modelSymbols = append(modelSymbols, group.RowSchema)
			}
			for _, field := range query.RowFields {
				modelSymbols = append(modelSymbols, extractSchemaReferences(field.Schema)...)
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
	ext := e.cfg.ImportExtension

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
	hasEmbeds := q.HasEmbeds

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

	// Build embed groups if applicable
	var embedGroups []EmbedGroupView
	var rowFields []SchemaField
	if hasEmbeds {
		embedGroups = e.buildEmbedGroups(q.EmbedGroups, q.Name)
		rowFields = e.buildEmbedRowFields(q.Results) // Use NullOr for embed row fields
	}

	view := QueryView{
		Name:                q.Name,
		NamePascal:          namePascal,
		NameCamel:           nameCamel,
		Command:             q.Command,
		HasParams:           hasParams,
		HasResults:          hasResults,
		HasEmbeds:           hasEmbeds,
		ReturnType:          returnType,
		SqlSchemaMethod:     sqlSchemaMethod,
		RequestSchema:       requestSchema,
		ParamFields:         e.buildParamFields(q.Params),
		ResultFields:        e.buildResultFields(q.Results),
		EmbedGroups:         embedGroups,
		RowFields:           rowFields,
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

// buildEmbedRowFields builds row fields for embed queries using NullOr instead of OptionFromNullOr
func (e *Effect4) buildEmbedRowFields(results []models.ResultField) []SchemaField {
	fields := make([]SchemaField, len(results))
	for i, r := range results {
		schema := e.sqlTypeToEffectSchemaBase(r.Type)
		if r.Type.IsNullable {
			schema = fmt.Sprintf("Schema.NullOr(%s)", schema)
		}
		fields[i] = SchemaField{
			Name:   r.Name,
			Schema: schema,
		}
	}
	return fields
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

// buildEmbedGroups converts model embed groups to view embed groups
func (e *Effect4) buildEmbedGroups(groups []models.EmbedGroup, queryName string) []EmbedGroupView {
	views := make([]EmbedGroupView, len(groups))
	for i, group := range groups {
		views[i] = e.buildEmbedGroup(group, queryName)
	}
	return views
}

// buildEmbedGroup creates an EmbedGroupView from a model EmbedGroup
func (e *Effect4) buildEmbedGroup(group models.EmbedGroup, queryName string) EmbedGroupView {
	// Singularize the table name for the field name (e.g., "orders" -> "order")
	fieldName := toCamelCase(singular(group.TableName))
	// Make schema name unique per query to avoid duplicate declarations
	schemaName := toPascalCase(queryName) + toPascalCase(group.TableName) + "Embed"

	// Build fields for this embed group
	fields := make([]SchemaField, len(group.Fields))
	fieldMappings := make([]FieldMap, len(group.Fields))
	for i, field := range group.Fields {
		// For the embed schema, strip the table prefix from the field name
		embedFieldName := field.Name
		if strings.HasPrefix(field.Name, group.TableName+"_") {
			embedFieldName = strings.TrimPrefix(field.Name, group.TableName+"_")
		}

		// For embed schemas (user-facing), use OptionFromNullOr for nullable fields
		// This provides consistent API with non-embed queries
		schema := e.sqlTypeToEffectSchemaForResults(field.Type)

		fields[i] = SchemaField{
			Name:   embedFieldName,
			Schema: schema,
		}
		fieldMappings[i] = FieldMap{
			RowFieldName:   field.Name,
			EmbedFieldName: embedFieldName,
		}
	}

	return EmbedGroupView{
		TableName:    group.TableName,
		FieldName:    fieldName,
		RowSchema:    toPascalCase(group.TableName) + "Row",
		SchemaName:   schemaName,
		Fields:       fields,
		FieldMapping: fieldMappings,
	}
}

// singular converts a plural word to singular form, handling known edge cases
// in the inflection library before falling back to it.
// See: https://github.com/jinzhu/inflection/issues
func singular(name string) string {
	lower := strings.ToLower(name)

	// Known inflection library bugs
	switch lower {
	// https://github.com/sqlc-dev/sqlc/issues/430
	// https://github.com/jinzhu/inflection/issues/13
	case "campus":
		return name

	// https://github.com/sqlc-dev/sqlc/issues/1217
	// https://github.com/jinzhu/inflection/issues/21
	case "meta", "metadata":
		return name

	// https://github.com/sqlc-dev/sqlc/issues/2017
	// https://github.com/jinzhu/inflection/issues/23
	case "calories":
		return "calorie"

	// Incorrect handling of "-ves" suffix
	case "waves":
		return "wave"
	}

	return inflection.Singular(name)
}

var _ Builder = (*Effect4)(nil)
