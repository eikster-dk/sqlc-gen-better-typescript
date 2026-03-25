package effect4

import (
	"fmt"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
	"github.com/jinzhu/inflection"
)

func (e *Effect4) buildQueryViews(queries []models.Query, log *logger.Logger) []QueryView {
	views := make([]QueryView, len(queries))
	for i, q := range queries {
		views[i] = e.buildQueryView(q, log)
	}
	return views
}

func (e *Effect4) buildQueryView(q models.Query, log *logger.Logger) QueryView {
	namePascal := toPascalCase(q.Name)
	nameCamel := toCamelCase(q.Name)
	hasParams := len(q.Params) > 0
	hasResults := len(q.Results) > 0
	hasEmbeds := q.HasEmbeds

	var returnType string
	switch q.Command {
	case ":exec":
		returnType = "void"
	case ":execrows":
		returnType = "number"
	case ":one":
		returnType = fmt.Sprintf("Option.Option<%sResult>", namePascal)
	default:
		returnType = fmt.Sprintf("%sResult[]", namePascal)
	}

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

	requestSchema := "Schema.Void"
	if hasParams {
		requestSchema = namePascal + "Params"
	}

	useTemplateLiterals := !e.cfg.DisableTemplateLiterals

	var embedGroups []EmbedGroupView
	var rowFields []SchemaField
	if hasEmbeds {
		embedGroups = e.buildEmbedGroups(q.EmbedGroups, q.Name)
		rowFields = e.buildEmbedRowFields(q.Results)
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

	if useTemplateLiterals {
		if hasParams {
			transformer := &SQLTransformer{}
			result, err := transformer.Transform(q.RewrittenSQL, q.Params, log)
			if err != nil {
				log.Warn("Failed to transform SQL to template literal, falling back to original",
					logger.F("query", q.Name),
					logger.F("error", err.Error()))
				view.SQLTemplateLiteral = q.RewrittenSQL
			} else {
				view.SQLTemplateLiteral = result.TemplateLiteral
				log.Debug("Successfully transformed SQL to template literal",
					logger.F("query", q.Name),
					logger.F("replacements", result.ReplacementsMade))
			}
		} else {
			view.SQLTemplateLiteral = q.RewrittenSQL
		}
	}

	return view
}

func (e *Effect4) buildParamFields(params []models.Param) []SchemaField {
	fields := make([]SchemaField, len(params))
	for i, p := range params {
		expr := e.sqlTypeToEffectSchemaForParams(p.Type)
		fields[i] = SchemaField{Name: toCamelCase(p.Name), Schema: expr.Schema, ModelImports: expr.ModelImports}
	}
	return fields
}

func (e *Effect4) buildResultFields(results []models.ResultField) []SchemaField {
	fields := make([]SchemaField, len(results))
	for i, r := range results {
		expr := e.sqlTypeToEffectSchemaForResults(r.Type)
		fields[i] = SchemaField{Name: r.Name, Schema: expr.Schema, ModelImports: expr.ModelImports}
	}
	return fields
}

func (e *Effect4) generateParamList(params []models.Param) string {
	if len(params) == 0 {
		return ""
	}
	parts := make([]string, 0, len(params))
	for _, param := range params {
		parts = append(parts, fmt.Sprintf("params.%s", toCamelCase(param.Name)))
	}
	return strings.Join(parts, ", ")
}

func (e *Effect4) buildEmbedRowFields(results []models.ResultField) []SchemaField {
	fields := make([]SchemaField, len(results))
	for i, r := range results {
		expr := e.sqlTypeToEffectSchemaBase(r.Type)
		schema := expr.Schema
		if r.Type.IsNullable {
			schema = fmt.Sprintf("Schema.NullOr(%s)", schema)
		}
		fields[i] = SchemaField{Name: r.Name, Schema: schema, ModelImports: expr.ModelImports}
	}
	return fields
}

func (e *Effect4) buildEmbedGroups(groups []models.EmbedGroup, queryName string) []EmbedGroupView {
	views := make([]EmbedGroupView, len(groups))
	for i, group := range groups {
		views[i] = e.buildEmbedGroup(group, queryName)
	}
	return views
}

func (e *Effect4) buildEmbedGroup(group models.EmbedGroup, queryName string) EmbedGroupView {
	fieldName := toCamelCase(singular(group.TableName))
	schemaName := toPascalCase(queryName) + toPascalCase(group.TableName) + "Embed"

	fields := make([]SchemaField, len(group.Fields))
	fieldMappings := make([]FieldMap, len(group.Fields))
	for i, field := range group.Fields {
		embedFieldName := field.Name
		if strings.HasPrefix(field.Name, group.TableName+"_") {
			embedFieldName = strings.TrimPrefix(field.Name, group.TableName+"_")
		}

		expr := e.sqlTypeToEffectSchemaForResults(field.Type)
		fields[i] = SchemaField{Name: embedFieldName, Schema: expr.Schema, ModelImports: expr.ModelImports}
		fieldMappings[i] = FieldMap{RowFieldName: field.Name, EmbedFieldName: embedFieldName}
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
