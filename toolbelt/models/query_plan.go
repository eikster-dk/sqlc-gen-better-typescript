package models

import "fmt"

// BuildQueryPlan derives a builder-facing plan from a mapped query.
func BuildQueryPlan(query Query) QueryPlan {
	row := buildRowPlan(query.Results)
	shape := buildResultShape(query)
	request := buildRequestPlan(query.Params)

	features := QueryFeatures{
		UsesEmbeds:  query.HasEmbeds || len(query.EmbedGroups) > 0,
		RewritesSQL: query.RewrittenSQL != "" && query.RewrittenSQL != query.SQL,
	}

	for _, field := range request.Fields {
		features.UsesNamedArgs = features.UsesNamedArgs || field.Named
		features.UsesNullableArgs = features.UsesNullableArgs || field.Nullable
		features.UsesSlices = features.UsesSlices || field.Slice
	}

	canonicalSQL := query.RewrittenSQL
	if canonicalSQL == "" {
		canonicalSQL = query.SQL
	}

	return QueryPlan{
		Name:     query.Name,
		Command:  query.Command,
		Filename: query.Filename,
		Source: QuerySourcePlan{
			OriginalSQL:  query.SQL,
			CanonicalSQL: canonicalSQL,
			ExecSQL:      canonicalSQL,
			Parameters:   buildSQLParameters(query.Params),
		},
		Request:  request,
		Response: ResponsePlan{Row: row, Result: ResultPlan{Shape: shape}},
		Features: features,
	}
}

// BuildQueryPlans derives plans for all mapped queries.
func BuildQueryPlans(queries []Query) []QueryPlan {
	plans := make([]QueryPlan, len(queries))
	for i, query := range queries {
		plans[i] = BuildQueryPlan(query)
	}
	return plans
}

func buildRequestPlan(params []Param) RequestPlan {
	fields := make([]RequestField, 0, len(params))
	seen := make(map[string]struct{})
	for _, param := range params {
		if _, ok := seen[param.Name]; ok {
			continue
		}
		seen[param.Name] = struct{}{}
		fields = append(fields, RequestField{
			Name:     param.Name,
			Param:    param,
			Type:     param.Type,
			Optional: param.Type.IsNullable,
			Nullable: param.Type.IsNullable,
			Slice:    param.Slice,
			Named:    param.Named,
		})
	}
	return RequestPlan{Fields: fields}
}

func buildSQLParameters(params []Param) []SQLParameter {
	sqlParams := make([]SQLParameter, len(params))
	for i, param := range params {
		sqlParams[i] = SQLParameter{
			Position:    param.Position,
			FieldName:   param.Name,
			Placeholder: fmt.Sprintf("$%d", param.Position),
			Slice:       param.Slice,
		}
	}
	return sqlParams
}

func buildRowPlan(results []ResultField) RowPlan {
	fields := make([]RowField, len(results))
	for i, result := range results {
		fields[i] = RowField{
			Name:         result.Name,
			OriginalName: result.OriginalName,
			Type:         result.Type,
			Table:        result.Table,
			EmbedTable:   result.EmbedTable,
			Aliased:      result.IsAliased,
		}
	}
	return RowPlan{Fields: fields}
}

func buildResultShape(query Query) ResultShape {
	if !query.HasEmbeds && len(query.EmbedGroups) == 0 {
		fields := make([]ResultShapeField, len(query.Results))
		for i, result := range query.Results {
			fields[i] = valueShapeField(result.Name, result.Name)
		}
		return ResultShape{Fields: fields}
	}

	fields := make([]ResultShapeField, 0, len(query.EmbedGroups))
	for _, group := range query.EmbedGroups {
		objectFields := make([]ResultShapeField, len(group.Fields))
		for i, field := range group.Fields {
			objectFields[i] = valueShapeField(embedFieldName(group.TableName, field.Name), field.Name)
		}
		object := ResultShape{Fields: objectFields}
		fields = append(fields, ResultShapeField{Kind: ResultShapeFieldObject, Name: group.TableName, Object: &object})
	}

	for _, result := range query.Results {
		if result.EmbedTable != "" {
			continue
		}
		fields = append(fields, valueShapeField(result.Name, result.Name))
	}

	return ResultShape{Fields: fields}
}

func valueShapeField(name, rowFieldName string) ResultShapeField {
	return ResultShapeField{Kind: ResultShapeFieldValue, Name: name, Source: &RowFieldRef{Name: rowFieldName}}
}

func embedFieldName(tableName, rowFieldName string) string {
	prefix := tableName + "_"
	if len(rowFieldName) > len(prefix) && rowFieldName[:len(prefix)] == prefix {
		return rowFieldName[len(prefix):]
	}
	return rowFieldName
}
