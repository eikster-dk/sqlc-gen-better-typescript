package effect4

import (
	"fmt"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
	"github.com/jinzhu/inflection"
)

func (e *Effect4) buildQueryViews(plans []models.QueryPlan, log *logger.Logger) []QueryView {
	views := make([]QueryView, len(plans))
	for i, plan := range plans {
		views[i] = e.buildQueryView(plan, log)
	}
	return views
}

func (e *Effect4) buildQueryView(plan models.QueryPlan, log *logger.Logger) QueryView {
	namePascal := toPascalCase(plan.Name)
	hasParams := len(plan.Request.Fields) > 0
	hasResults := len(plan.Response.Row.Fields) > 0

	var returnType string
	switch plan.Command {
	case ":exec":
		returnType = "void"
	case ":execrows":
		returnType = "number"
	case ":execresult":
		returnType = "SqlExecResult"
	case ":one":
		returnType = fmt.Sprintf("Option.Option<%sResult>", namePascal)
	default:
		returnType = fmt.Sprintf("%sResult[]", namePascal)
	}

	var sqlSchemaMethod string
	switch plan.Command {
	case ":exec":
		sqlSchemaMethod = "SqlSchema.void"
	case ":execrows":
		sqlSchemaMethod = "execRows"
	case ":execresult":
		sqlSchemaMethod = "execResult"
	case ":one":
		sqlSchemaMethod = "SqlSchema.findOneOption"
	default:
		sqlSchemaMethod = "SqlSchema.findAll"
	}

	requestSchema := "Schema.Void"
	if hasParams {
		requestSchema = namePascal + "Params"
	}

	var embedGroups []EmbedGroupView
	var rowFields []SchemaField
	if plan.Features.UsesEmbeds {
		embedGroups = e.buildEmbedGroupsFromPlan(plan)
		rowFields = e.buildEmbedRowFieldsFromPlan(plan)
	}

	useTemplateLiterals := !e.cfg.DisableTemplateLiterals
	view := QueryView{
		Name:                plan.Name,
		NamePascal:          namePascal,
		NameCamel:           toCamelCase(plan.Name),
		Command:             plan.Command,
		HasParams:           hasParams,
		HasResults:          hasResults,
		HasEmbeds:           plan.Features.UsesEmbeds,
		ReturnType:          returnType,
		SqlSchemaMethod:     sqlSchemaMethod,
		RequestSchema:       requestSchema,
		ParamFields:         e.buildParamFields(plan.Request.Fields),
		ResultFields:        e.buildResultFields(plan.Response.Row.Fields),
		EmbedGroups:         embedGroups,
		RowFields:           rowFields,
		OriginalSQL:         plan.Source.ExecSQL,
		ParamList:           e.generateParamList(plan.Request.Fields),
		UseTemplateLiterals: useTemplateLiterals,
	}

	if useTemplateLiterals {
		if hasParams {
			transformer := &SQLTransformer{}
			result, err := transformer.Transform(plan.Source.ExecSQL, paramsFromRequestFields(plan.Request.Fields), log)
			if err != nil {
				log.Warn("Failed to transform SQL to template literal, falling back to original", logger.F("query", plan.Name), logger.F("error", err.Error()))
				view.SQLTemplateLiteral = plan.Source.ExecSQL
			} else {
				view.SQLTemplateLiteral = result.TemplateLiteral
			}
		} else {
			view.SQLTemplateLiteral = plan.Source.ExecSQL
		}
	}

	return view
}

func (e *Effect4) buildParamFields(requestFields []models.RequestField) []SchemaField {
	fields := make([]SchemaField, len(requestFields))
	for i, field := range requestFields {
		expr := e.sqlTypeToEffectSchemaForParams(field.Type)
		fields[i] = SchemaField{Name: toCamelCase(field.Name), Schema: expr.Schema, ModelImports: expr.ModelImports}
	}
	return fields
}

func (e *Effect4) buildResultFields(rowFields []models.RowField) []SchemaField {
	fields := make([]SchemaField, len(rowFields))
	for i, field := range rowFields {
		expr := e.sqlTypeToEffectSchemaForResults(field.Type)
		fields[i] = SchemaField{Name: field.Name, Schema: expr.Schema, ModelImports: expr.ModelImports}
	}
	return fields
}

func (e *Effect4) generateParamList(requestFields []models.RequestField) string {
	if len(requestFields) == 0 {
		return ""
	}
	parts := make([]string, 0, len(requestFields))
	for _, field := range requestFields {
		parts = append(parts, fmt.Sprintf("params.%s", toCamelCase(field.Name)))
	}
	return strings.Join(parts, ", ")
}

func paramsFromRequestFields(requestFields []models.RequestField) []models.Param {
	params := make([]models.Param, len(requestFields))
	for i, field := range requestFields {
		params[i] = field.Param
	}
	return params
}

func (e *Effect4) buildEmbedRowFieldsFromPlan(plan models.QueryPlan) []SchemaField {
	fields := make([]SchemaField, len(plan.Response.Row.Fields))
	for i, rowField := range plan.Response.Row.Fields {
		expr := e.sqlTypeToEffectSchemaBase(rowField.Type)
		schema := expr.Schema
		if rowField.Type.IsNullable {
			schema = fmt.Sprintf("Schema.NullOr(%s)", schema)
		}
		fields[i] = SchemaField{Name: rowField.Name, Schema: schema, ModelImports: expr.ModelImports}
	}
	return fields
}

func (e *Effect4) buildEmbedGroupsFromPlan(plan models.QueryPlan) []EmbedGroupView {
	views := make([]EmbedGroupView, 0, len(plan.Response.Result.Shape.Fields))
	rowFields := make(map[string]models.RowField, len(plan.Response.Row.Fields))
	for _, field := range plan.Response.Row.Fields {
		rowFields[field.Name] = field
	}
	for _, field := range plan.Response.Result.Shape.Fields {
		if field.Kind != models.ResultShapeFieldObject || field.Object == nil {
			continue
		}
		views = append(views, e.buildEmbedGroupFromShape(field, plan.Name, rowFields))
	}
	return views
}

func (e *Effect4) buildEmbedGroupFromShape(field models.ResultShapeField, queryName string, rowFields map[string]models.RowField) EmbedGroupView {
	tableName := field.Name
	fieldName := toCamelCase(singular(tableName))
	schemaName := toPascalCase(queryName) + toPascalCase(tableName) + "Embed"
	fields := make([]SchemaField, 0, len(field.Object.Fields))
	fieldMappings := make([]FieldMap, 0, len(field.Object.Fields))
	for _, child := range field.Object.Fields {
		if child.Kind != models.ResultShapeFieldValue || child.Source == nil {
			continue
		}
		rowField := child.Source.Name
		expr := e.sqlTypeToEffectSchemaForResults(rowFields[rowField].Type)
		fields = append(fields, SchemaField{Name: child.Name, Schema: expr.Schema, ModelImports: expr.ModelImports})
		fieldMappings = append(fieldMappings, FieldMap{RowFieldName: rowField, EmbedFieldName: child.Name})
	}
	return EmbedGroupView{TableName: tableName, FieldName: fieldName, RowSchema: toPascalCase(tableName) + "Row", SchemaName: schemaName, Fields: fields, FieldMapping: fieldMappings}
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
