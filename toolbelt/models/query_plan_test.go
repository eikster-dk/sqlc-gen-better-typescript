package models

import "testing"

func TestBuildQueryPlanFlatResultAndSQLVariants(t *testing.T) {
	query := Query{
		Name:         "GetPost",
		Command:      ":one",
		SQL:          "SELECT users.id, posts.id FROM users JOIN posts ON posts.user_id = users.id WHERE users.id = $1",
		RewrittenSQL: "SELECT users.id, posts.id AS posts_id FROM users JOIN posts ON posts.user_id = users.id WHERE users.id = $1",
		Params: []Param{
			{Name: "id", Position: 1, Type: SqlType{Name: "integer"}, Named: true},
		},
		Results: []ResultField{
			{Name: "id", OriginalName: "id", Type: SqlType{Name: "integer"}, Table: "users"},
			{Name: "posts_id", OriginalName: "id", Type: SqlType{Name: "integer"}, Table: "posts", IsAliased: true},
		},
	}

	plan := BuildQueryPlan(query)

	if plan.Source.OriginalSQL != query.SQL {
		t.Fatalf("expected original SQL to be preserved")
	}
	if plan.Source.CanonicalSQL != query.RewrittenSQL || plan.Source.ExecSQL != query.RewrittenSQL {
		t.Fatalf("expected canonical and exec SQL to use rewritten SQL")
	}
	if !plan.Features.RewritesSQL || !plan.Features.UsesNamedArgs {
		t.Fatalf("expected rewrite and named arg features: %#v", plan.Features)
	}
	if len(plan.Request.Fields) != 1 || plan.Request.Fields[0].Name != "id" {
		t.Fatalf("unexpected request fields: %#v", plan.Request.Fields)
	}
	if len(plan.Source.Parameters) != 1 || plan.Source.Parameters[0].Placeholder != "$1" {
		t.Fatalf("unexpected SQL parameters: %#v", plan.Source.Parameters)
	}
	if len(plan.Response.Row.Fields) != 2 || !plan.Response.Row.Fields[1].Aliased {
		t.Fatalf("unexpected row fields: %#v", plan.Response.Row.Fields)
	}
	assertValueField(t, plan.Response.Result.Shape.Fields[0], "id", "id")
	assertValueField(t, plan.Response.Result.Shape.Fields[1], "posts_id", "posts_id")
}

func TestBuildQueryPlanRepeatedParamOccurrences(t *testing.T) {
	query := Query{
		Name:    "FindNode",
		Command: ":many",
		SQL:     "SELECT id FROM nodes WHERE id = $1 OR parent_id = $1",
		Params: []Param{
			{Name: "id", Position: 1, Type: SqlType{Name: "integer"}, Named: true},
			{Name: "id", Position: 1, Type: SqlType{Name: "integer"}, Named: true},
		},
	}

	plan := BuildQueryPlan(query)

	if len(plan.Request.Fields) != 1 {
		t.Fatalf("expected one logical request field, got %#v", plan.Request.Fields)
	}
	if len(plan.Source.Parameters) != 2 {
		t.Fatalf("expected two SQL parameter occurrences, got %#v", plan.Source.Parameters)
	}
	for _, param := range plan.Source.Parameters {
		if param.FieldName != "id" || param.Placeholder != "$1" {
			t.Fatalf("unexpected parameter occurrence: %#v", param)
		}
	}
}

func TestBuildQueryPlanNullableSliceFeatures(t *testing.T) {
	query := Query{
		Name:    "Search",
		Command: ":many",
		SQL:     "SELECT id FROM users WHERE id = ANY($1) AND name = coalesce($2, name)",
		Params: []Param{
			{Name: "ids", Position: 1, Type: SqlType{Name: "integer", IsArray: true}, Named: true, Slice: true},
			{Name: "name", Position: 2, Type: SqlType{Name: "text", IsNullable: true}, Named: true},
		},
	}

	plan := BuildQueryPlan(query)

	if !plan.Features.UsesNamedArgs || !plan.Features.UsesNullableArgs || !plan.Features.UsesSlices {
		t.Fatalf("expected named, nullable, and slice features: %#v", plan.Features)
	}
	if !plan.Request.Fields[0].Slice || !plan.Source.Parameters[0].Slice {
		t.Fatalf("expected slice to be represented on request and SQL parameter")
	}
	if !plan.Request.Fields[1].Nullable || !plan.Request.Fields[1].Optional {
		t.Fatalf("expected nullable param to be optional: %#v", plan.Request.Fields[1])
	}
}

func TestBuildQueryPlanEmbedResultShape(t *testing.T) {
	query := Query{
		Name:      "GetOrderWithCustomer",
		Command:   ":one",
		SQL:       "SELECT sqlc.embed(orders), sqlc.embed(customers) FROM orders JOIN customers ON customers.id = orders.customer_id",
		HasEmbeds: true,
		Results: []ResultField{
			{Name: "orders_id", OriginalName: "id", Type: SqlType{Name: "integer"}, Table: "orders", EmbedTable: "orders", IsAliased: true},
			{Name: "orders_status", OriginalName: "status", Type: SqlType{Name: "text"}, Table: "orders", EmbedTable: "orders", IsAliased: true},
			{Name: "customers_id", OriginalName: "id", Type: SqlType{Name: "integer"}, Table: "customers", EmbedTable: "customers", IsAliased: true},
			{Name: "customers_name", OriginalName: "name", Type: SqlType{Name: "text"}, Table: "customers", EmbedTable: "customers", IsAliased: true},
		},
		EmbedGroups: []EmbedGroup{
			{TableName: "orders", Fields: []ResultField{
				{Name: "orders_id", OriginalName: "id", Type: SqlType{Name: "integer"}, Table: "orders", EmbedTable: "orders", IsAliased: true},
				{Name: "orders_status", OriginalName: "status", Type: SqlType{Name: "text"}, Table: "orders", EmbedTable: "orders", IsAliased: true},
			}},
			{TableName: "customers", Fields: []ResultField{
				{Name: "customers_id", OriginalName: "id", Type: SqlType{Name: "integer"}, Table: "customers", EmbedTable: "customers", IsAliased: true},
				{Name: "customers_name", OriginalName: "name", Type: SqlType{Name: "text"}, Table: "customers", EmbedTable: "customers", IsAliased: true},
			}},
		},
	}

	plan := BuildQueryPlan(query)

	if !plan.Features.UsesEmbeds {
		t.Fatalf("expected embed feature")
	}
	if len(plan.Response.Row.Fields) != 4 {
		t.Fatalf("unexpected row fields: %#v", plan.Response.Row.Fields)
	}
	shape := plan.Response.Result.Shape
	if len(shape.Fields) != 2 {
		t.Fatalf("expected two embedded objects, got %#v", shape.Fields)
	}
	orders := shape.Fields[0]
	if orders.Kind != ResultShapeFieldObject || orders.Name != "orders" || orders.Object == nil {
		t.Fatalf("unexpected orders shape: %#v", orders)
	}
	assertValueField(t, orders.Object.Fields[0], "id", "orders_id")
	assertValueField(t, orders.Object.Fields[1], "status", "orders_status")

	customers := shape.Fields[1]
	if customers.Kind != ResultShapeFieldObject || customers.Name != "customers" || customers.Object == nil {
		t.Fatalf("unexpected customers shape: %#v", customers)
	}
	assertValueField(t, customers.Object.Fields[0], "id", "customers_id")
	assertValueField(t, customers.Object.Fields[1], "name", "customers_name")
}

func assertValueField(t *testing.T, field ResultShapeField, name, source string) {
	t.Helper()
	if field.Kind != ResultShapeFieldValue || field.Name != name || field.Source == nil || field.Source.Name != source {
		t.Fatalf("unexpected value field: %#v", field)
	}
}
