package native

import (
	"strings"
	"testing"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

// oneQuery returns a simple :one query with one param and one result.
func oneQuery() models.Query {
	return models.Query{
		Name:         "GetCustomer",
		SQL:          "SELECT id, name FROM customers WHERE id = $1",
		RewrittenSQL: "SELECT id, name FROM customers WHERE id = $1",
		Command:      ":one",
		Filename:     "customers.sql",
		Params: []models.Param{
			{Name: "id", Position: 1, Type: models.SqlType{Name: "integer", IsNullable: false}},
		},
		Results: []models.ResultField{
			{Name: "id", OriginalName: "id", Type: models.SqlType{Name: "integer", IsNullable: false}},
			{Name: "name", OriginalName: "name", Type: models.SqlType{Name: "text", IsNullable: false}},
		},
	}
}

// TestNative_Build_GeneratesQueryFiles verifies that query files are generated when
// queries are provided to Build.
func TestNative_Build_GeneratesQueryFiles(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	queries := []models.Query{oneQuery()}

	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expect models.ts + {name}Requests.ts + {name}Responses.ts + {name}Queries.ts
	if len(files) != 4 {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Name
		}
		t.Fatalf("expected 4 files, got %d: %v", len(files), names)
	}
}

func TestNative_Build_RequestsFile(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	queries := []models.Query{oneQuery()}
	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var requestsFile *File
	for i := range files {
		if files[i].Name == "customersRequests.ts" {
			requestsFile = &files[i]
			break
		}
	}
	if requestsFile == nil {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Name
		}
		t.Fatalf("expected customersRequests.ts, got files: %v", names)
	}

	content := string(requestsFile.Content)

	t.Run("imports z from zod", func(t *testing.T) {
		if !strings.Contains(content, `import { z } from "zod"`) {
			t.Errorf("expected z import, got:\n%s", content)
		}
	})

	t.Run("exports GetCustomer param schema", func(t *testing.T) {
		if !strings.Contains(content, "export const GetCustomerParams") {
			t.Errorf("expected GetCustomerParams schema, got:\n%s", content)
		}
	})

	t.Run("exports GetCustomer param type", func(t *testing.T) {
		if !strings.Contains(content, "export type GetCustomerParams") {
			t.Errorf("expected GetCustomerParams type, got:\n%s", content)
		}
	})

	t.Run("maps integer param to z.number()", func(t *testing.T) {
		if !strings.Contains(content, "z.number()") {
			t.Errorf("expected z.number() for integer param, got:\n%s", content)
		}
	})
}

func TestNative_Build_ResponsesFile(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	queries := []models.Query{oneQuery()}
	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var responsesFile *File
	for i := range files {
		if files[i].Name == "customersResponses.ts" {
			responsesFile = &files[i]
			break
		}
	}
	if responsesFile == nil {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Name
		}
		t.Fatalf("expected customersResponses.ts, got files: %v", names)
	}

	content := string(responsesFile.Content)

	t.Run("imports z from zod", func(t *testing.T) {
		if !strings.Contains(content, `import { z } from "zod"`) {
			t.Errorf("expected z import, got:\n%s", content)
		}
	})

	t.Run("exports GetCustomer result schema", func(t *testing.T) {
		if !strings.Contains(content, "export const GetCustomerResult") {
			t.Errorf("expected GetCustomerResult schema, got:\n%s", content)
		}
	})

	t.Run("exports GetCustomer result type", func(t *testing.T) {
		if !strings.Contains(content, "export type GetCustomerResult") {
			t.Errorf("expected GetCustomerResult type, got:\n%s", content)
		}
	})

	t.Run("maps text result field to z.string()", func(t *testing.T) {
		if !strings.Contains(content, "z.string()") {
			t.Errorf("expected z.string() for text result field, got:\n%s", content)
		}
	})
}

func TestNative_Build_QueriesFile(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	queries := []models.Query{oneQuery()}
	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "customersQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Name
		}
		t.Fatalf("expected customersQueries.ts, got files: %v", names)
	}

	content := string(queriesFile.Content)

	t.Run("imports SqlClient and QueryResult from models", func(t *testing.T) {
		if !strings.Contains(content, "SqlClient") {
			t.Errorf("expected SqlClient import, got:\n%s", content)
		}
		if !strings.Contains(content, "QueryResult") {
			t.Errorf("expected QueryResult import, got:\n%s", content)
		}
	})

	t.Run("exports async getCustomer function", func(t *testing.T) {
		if !strings.Contains(content, "export async function getCustomer") {
			t.Errorf("expected exported async getCustomer function, got:\n%s", content)
		}
	})

	t.Run("function accepts SqlClient and params", func(t *testing.T) {
		if !strings.Contains(content, "client: SqlClient") {
			t.Errorf("expected client: SqlClient param, got:\n%s", content)
		}
		if !strings.Contains(content, "params: GetCustomerParams") {
			t.Errorf("expected params: GetCustomerParams, got:\n%s", content)
		}
	})

	t.Run("returns Promise<QueryResult<GetCustomerResult | null>>", func(t *testing.T) {
		if !strings.Contains(content, "Promise<QueryResult<GetCustomerResult | null>>") {
			t.Errorf("expected Promise<QueryResult<GetCustomerResult | null>> return type, got:\n%s", content)
		}
	})

	t.Run("validates input with safeParse", func(t *testing.T) {
		if !strings.Contains(content, "safeParse") {
			t.Errorf("expected safeParse call for input validation, got:\n%s", content)
		}
	})

	t.Run("includes phase input for validation failure", func(t *testing.T) {
		if !strings.Contains(content, `"input"`) {
			t.Errorf(`expected phase: "input" in validation failure, got:\n%s`, content)
		}
	})

	t.Run("includes phase output for result validation failure", func(t *testing.T) {
		if !strings.Contains(content, `"output"`) {
			t.Errorf(`expected phase: "output" in result validation failure, got:\n%s`, content)
		}
	})

	t.Run("uses parameterized query with $1", func(t *testing.T) {
		if !strings.Contains(content, "$1") {
			t.Errorf("expected parameterized query with $1, got:\n%s", content)
		}
	})
}

func TestNative_Build_QueriesFile_SQLComment(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	queries := []models.Query{oneQuery()}
	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "customersQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		t.Fatal("expected customersQueries.ts")
	}

	content := string(queriesFile.Content)

	t.Run("SQL comment appears above function", func(t *testing.T) {
		// Each line of the SQL should appear as a // comment before the function
		if !strings.Contains(content, "// SELECT id, name FROM customers WHERE id = $1") {
			t.Errorf("expected SQL as comment above function, got:\n%s", content)
		}
	})

	t.Run("SQL comment precedes the export keyword", func(t *testing.T) {
		commentIdx := strings.Index(content, "// SELECT id, name FROM customers WHERE id = $1")
		funcIdx := strings.Index(content, "export async function getCustomer")
		if commentIdx == -1 || funcIdx == -1 || commentIdx > funcIdx {
			t.Errorf("expected SQL comment to appear before function declaration, got:\n%s", content)
		}
	})

	t.Run("multiline SQL has each line commented", func(t *testing.T) {
		multilineQuery := models.Query{
			Name:         "GetCustomer",
			SQL:          "SELECT id, name\nFROM customers\nWHERE id = $1",
			RewrittenSQL: "SELECT id, name\nFROM customers\nWHERE id = $1",
			Command:      ":one",
			Filename:     "customers.sql",
			Params:       []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}},
			Results: []models.ResultField{
				{Name: "id", Type: models.SqlType{Name: "integer"}},
				{Name: "name", Type: models.SqlType{Name: "text"}},
			},
		}
		files2, err := n.Build(defaultCatalog(), []models.Query{multilineQuery}, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var qf *File
		for i := range files2 {
			if files2[i].Name == "customersQueries.ts" {
				qf = &files2[i]
				break
			}
		}
		if qf == nil {
			t.Fatal("expected customersQueries.ts")
		}
		c := string(qf.Content)
		if !strings.Contains(c, "// SELECT id, name") {
			t.Errorf("expected first SQL line as comment, got:\n%s", c)
		}
		if !strings.Contains(c, "// FROM customers") {
			t.Errorf("expected second SQL line as comment, got:\n%s", c)
		}
		if !strings.Contains(c, "// WHERE id = $1") {
			t.Errorf("expected third SQL line as comment, got:\n%s", c)
		}
	})
}

func TestNative_Build_OneQuery_NoParamQuery(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	// Query with no params (e.g., list all)
	noParamQuery := models.Query{
		Name:         "ListCustomers",
		SQL:          "SELECT id, name FROM customers ORDER BY id",
		RewrittenSQL: "SELECT id, name FROM customers ORDER BY id",
		Command:      ":one",
		Filename:     "customers.sql",
		Params:       []models.Param{},
		Results: []models.ResultField{
			{Name: "id", OriginalName: "id", Type: models.SqlType{Name: "integer", IsNullable: false}},
			{Name: "name", OriginalName: "name", Type: models.SqlType{Name: "text", IsNullable: false}},
		},
	}

	files, err := n.Build(defaultCatalog(), []models.Query{noParamQuery}, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "customersQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		t.Fatal("expected customersQueries.ts")
	}

	content := string(queriesFile.Content)

	t.Run("function has no params argument when no params", func(t *testing.T) {
		// Function should only accept client, no params arg
		if strings.Contains(content, "GetCustomerParams") {
			t.Errorf("expected no GetCustomerParams for no-param query, got:\n%s", content)
		}
	})

	t.Run("no input validation when no params", func(t *testing.T) {
		// No safeParse on input side when there are no params
		if strings.Contains(content, `phase: "input"`) {
			t.Errorf("expected no input phase validation for no-param query, got:\n%s", content)
		}
	})
}

func TestNative_ZodTypeMapping(t *testing.T) {
	n := New(defaultConfig())

	tests := []struct {
		typeName string
		want     string
	}{
		{"integer", "z.number()"},
		{"serial", "z.number()"},
		{"smallint", "z.number()"},
		{"int", "z.number()"},
		{"int2", "z.number()"},
		{"int4", "z.number()"},
		{"serial2", "z.number()"},
		{"serial4", "z.number()"},
		{"smallserial", "z.number()"},
		{"float", "z.number()"},
		{"float4", "z.number()"},
		{"float8", "z.number()"},
		{"double precision", "z.number()"},
		{"real", "z.number()"},
		{"text", "z.string()"},
		{"varchar", "z.string()"},
		{"char", "z.string()"},
		{"bpchar", "z.string()"},
		{"citext", "z.string()"},
		{"numeric", "z.string()"},
		{"money", "z.string()"},
		{"time", "z.string()"},
		{"timetz", "z.string()"},
		{"interval", "z.string()"},
		{"inet", "z.string()"},
		{"cidr", "z.string()"},
		{"macaddr", "z.string()"},
		{"macaddr8", "z.string()"},
		{"ltree", "z.string()"},
		{"lquery", "z.string()"},
		{"ltxtquery", "z.string()"},
		{"boolean", "z.boolean()"},
		{"bool", "z.boolean()"},
		{"bigint", "z.coerce.bigint()"},
		{"int8", "z.coerce.bigint()"},
		{"bigserial", "z.coerce.bigint()"},
		{"serial8", "z.coerce.bigint()"},
		{"uuid", "z.string().uuid()"},
		{"json", "z.unknown()"},
		{"jsonb", "z.unknown()"},
		{"bytea", "z.instanceof(Buffer)"},
		{"blob", "z.instanceof(Buffer)"},
		{"date", "z.coerce.date()"},
		{"timestamp", "z.coerce.date()"},
		{"timestamptz", "z.coerce.date()"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			got := n.zodBaseType(models.SqlType{Name: tt.typeName})
			if got != tt.want {
				t.Errorf("zodBaseType(%q) = %q, want %q", tt.typeName, got, tt.want)
			}
		})
	}

	t.Run("unknown type maps to z.unknown()", func(t *testing.T) {
		got := n.zodBaseType(models.SqlType{Name: "some_unknown_type"})
		if got != "z.unknown()" {
			t.Errorf("zodBaseType(unknown) = %q, want %q", got, "z.unknown()")
		}
	})

	t.Run("enum type without catalog falls back to z.string()", func(t *testing.T) {
		got := n.zodBaseType(models.SqlType{Name: "user_role", IsEnum: true})
		if got != "z.string()" {
			t.Errorf("zodBaseType(enum without catalog) = %q, want %q", got, "z.string()")
		}
	})

	t.Run("case insensitive mapping", func(t *testing.T) {
		for _, name := range []string{"INTEGER", "Integer", "TEXT", "Text", "BOOLEAN", "Boolean"} {
			got := n.zodBaseType(models.SqlType{Name: name})
			if got == "z.unknown()" {
				t.Errorf("zodBaseType(%q) = z.unknown(), expected a known type", name)
			}
		}
	})
}

func TestNative_ZodTypeMapping_Nullable(t *testing.T) {
	n := New(defaultConfig())

	t.Run("nullable param uses .optional()", func(t *testing.T) {
		sqlType := models.SqlType{Name: "text", IsNullable: true}
		got := n.zodTypeForParam(sqlType)
		want := "z.string().optional()"
		if got != want {
			t.Errorf("zodTypeForParam(nullable text) = %q, want %q", got, want)
		}
	})

	t.Run("nullable result uses .nullable()", func(t *testing.T) {
		sqlType := models.SqlType{Name: "text", IsNullable: true}
		got := n.zodTypeForResult(sqlType)
		want := "z.string().nullable()"
		if got != want {
			t.Errorf("zodTypeForResult(nullable text) = %q, want %q", got, want)
		}
	})

	t.Run("non-nullable result does not use .nullable()", func(t *testing.T) {
		sqlType := models.SqlType{Name: "text", IsNullable: false}
		got := n.zodTypeForResult(sqlType)
		want := "z.string()"
		if got != want {
			t.Errorf("zodTypeForResult(non-nullable text) = %q, want %q", got, want)
		}
	})
}

func TestNative_ZodTypeMapping_Array(t *testing.T) {
	n := New(defaultConfig())

	t.Run("array of text", func(t *testing.T) {
		got := n.zodTypeForParam(models.SqlType{Name: "text", IsArray: true})
		if got != "z.array(z.string())" {
			t.Errorf("zodTypeForParam(text[]) = %q, want %q", got, "z.array(z.string())")
		}
	})

	t.Run("array of integer", func(t *testing.T) {
		got := n.zodTypeForResult(models.SqlType{Name: "integer", IsArray: true})
		if got != "z.array(z.number())" {
			t.Errorf("zodTypeForResult(integer[]) = %q, want %q", got, "z.array(z.number())")
		}
	})

	t.Run("nullable array param", func(t *testing.T) {
		got := n.zodTypeForParam(models.SqlType{Name: "text", IsArray: true, IsNullable: true})
		if got != "z.array(z.string()).optional()" {
			t.Errorf("zodTypeForParam(nullable text[]) = %q, want %q", got, "z.array(z.string()).optional()")
		}
	})

	t.Run("nullable array result", func(t *testing.T) {
		got := n.zodTypeForResult(models.SqlType{Name: "text", IsArray: true, IsNullable: true})
		if got != "z.array(z.string()).nullable()" {
			t.Errorf("zodTypeForResult(nullable text[]) = %q, want %q", got, "z.array(z.string()).nullable()")
		}
	})

	t.Run("array of enum preserves enum flag", func(t *testing.T) {
		got := n.zodTypeForParam(models.SqlType{Name: "user_role", IsArray: true, IsEnum: true})
		want := "z.array(z.string())"
		if got != want {
			t.Errorf("zodTypeForParam(enum[]) = %q, want %q", got, want)
		}
	})

	t.Run("nullable array of enum preserves enum flag", func(t *testing.T) {
		got := n.zodTypeForResult(models.SqlType{Name: "user_role", IsArray: true, IsEnum: true, IsNullable: true})
		want := "z.array(z.string()).nullable()"
		if got != want {
			t.Errorf("zodTypeForResult(nullable enum[]) = %q, want %q", got, want)
		}
	})
}

func TestNative_Build_MultipleQueriesSameFile(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	queries := []models.Query{
		{
			Name:         "GetCustomer",
			SQL:          "SELECT id FROM customers WHERE id = $1",
			RewrittenSQL: "SELECT id FROM customers WHERE id = $1",
			Command:      ":one",
			Filename:     "customers.sql",
			Params:       []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}},
			Results:      []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		},
		{
			Name:         "GetCustomerByEmail",
			SQL:          "SELECT id FROM customers WHERE email = $1",
			RewrittenSQL: "SELECT id FROM customers WHERE email = $1",
			Command:      ":one",
			Filename:     "customers.sql",
			Params:       []models.Param{{Name: "email", Position: 1, Type: models.SqlType{Name: "text"}}},
			Results:      []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		},
	}

	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Still 4 files: models.ts + 3 for customers
	if len(files) != 4 {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Name
		}
		t.Fatalf("expected 4 files, got %d: %v", len(files), names)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "customersQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		t.Fatal("expected customersQueries.ts")
	}

	content := string(queriesFile.Content)

	t.Run("contains both functions", func(t *testing.T) {
		if !strings.Contains(content, "getCustomer") {
			t.Errorf("expected getCustomer function, got:\n%s", content)
		}
		if !strings.Contains(content, "getCustomerByEmail") {
			t.Errorf("expected getCustomerByEmail function, got:\n%s", content)
		}
	})
}

func TestNative_Build_MultipleFiles(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	queries := []models.Query{
		{
			Name:         "GetCustomer",
			SQL:          "SELECT id FROM customers WHERE id = $1",
			RewrittenSQL: "SELECT id FROM customers WHERE id = $1",
			Command:      ":one",
			Filename:     "customers.sql",
			Params:       []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}},
			Results:      []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		},
		{
			Name:         "GetProduct",
			SQL:          "SELECT id FROM products WHERE id = $1",
			RewrittenSQL: "SELECT id FROM products WHERE id = $1",
			Command:      ":one",
			Filename:     "products.sql",
			Params:       []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}},
			Results:      []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		},
	}

	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// models.ts + 3 per file * 2 files = 7
	if len(files) != 7 {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Name
		}
		t.Fatalf("expected 7 files, got %d: %v", len(files), names)
	}
}

func TestNative_Build_EmptyQuerySlice(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	files, err := n.Build(defaultCatalog(), []models.Query{}, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Empty query slice should produce only models.ts
	if len(files) != 1 {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Name
		}
		t.Fatalf("expected 1 file, got %d: %v", len(files), names)
	}
	if files[0].Name != "models.ts" {
		t.Errorf("expected models.ts, got %q", files[0].Name)
	}
}

func TestNative_Build_NilImportExtension(t *testing.T) {
	cfg := config.Config{Builder: "native", Driver: "pg", Validator: "zod", ImportExtension: nil}
	n := New(cfg)
	log := logger.New(false)

	queries := []models.Query{oneQuery()}
	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "customersQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		t.Fatal("expected customersQueries.ts")
	}

	content := string(queriesFile.Content)

	// With nil import extension, imports should have no extension
	if !strings.Contains(content, `"./models"`) {
		t.Errorf("expected import from ./models (no ext), got:\n%s", content)
	}
}

func TestNative_Build_EmptyImportExtension(t *testing.T) {
	ext := ""
	cfg := config.Config{Builder: "native", Driver: "pg", Validator: "zod", ImportExtension: &ext}
	n := New(cfg)
	log := logger.New(false)

	queries := []models.Query{oneQuery()}
	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "customersQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		t.Fatal("expected customersQueries.ts")
	}

	content := string(queriesFile.Content)

	// With empty import extension, imports should have no extension
	if !strings.Contains(content, `"./models"`) {
		t.Errorf("expected import from ./models (no ext), got:\n%s", content)
	}
}

func TestNative_Build_RewrittenSQLFallback(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	// Query with empty RewrittenSQL should fall back to SQL
	q := models.Query{
		Name:         "GetCustomer",
		SQL:          "SELECT id FROM customers WHERE id = $1",
		RewrittenSQL: "",
		Command:      ":one",
		Filename:     "customers.sql",
		Params:       []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}},
		Results:      []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
	}

	files, err := n.Build(defaultCatalog(), []models.Query{q}, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "customersQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		t.Fatal("expected customersQueries.ts")
	}

	content := string(queriesFile.Content)

	if !strings.Contains(content, "SELECT id FROM customers WHERE id = $1") {
		t.Errorf("expected SQL fallback when RewrittenSQL is empty, got:\n%s", content)
	}
}

func TestNative_Build_QueryWithNoResults(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	q := models.Query{
		Name:         "GetEmpty",
		SQL:          "SELECT 1 WHERE false",
		RewrittenSQL: "SELECT 1 WHERE false",
		Command:      ":one",
		Filename:     "empty.sql",
		Params:       []models.Param{},
		Results:      []models.ResultField{},
	}

	files, err := n.Build(defaultCatalog(), []models.Query{q}, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "emptyQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		t.Fatal("expected emptyQueries.ts")
	}

	content := string(queriesFile.Content)

	// No results means return type should be null, not a type | null
	if !strings.Contains(content, "Promise<QueryResult<null>>") {
		t.Errorf("expected Promise<QueryResult<null>> for no-result query, got:\n%s", content)
	}

	// Should not have output validation
	if strings.Contains(content, `phase: "output"`) {
		t.Errorf("expected no output validation for no-result query, got:\n%s", content)
	}
}

func TestNative_CaseConversion(t *testing.T) {
	tests := []struct {
		input      string
		wantPascal string
		wantCamel  string
	}{
		{"get_customer", "GetCustomer", "getCustomer"},
		{"GetCustomer", "GetCustomer", "getCustomer"},
		{"customer", "Customer", "customer"},
		{"a", "A", "a"},
		{"", "", ""},
		{"get_customer_by_id", "GetCustomerById", "getCustomerById"},
		{"_leading", "Leading", "leading"},
		{"trailing_", "Trailing", "trailing"},
		{"__double__under__", "DoubleUnder", "doubleUnder"},
		{"ALL_CAPS", "ALLCAPS", "aLLCAPS"},
	}

	for _, tt := range tests {
		t.Run("pascal_"+tt.input, func(t *testing.T) {
			got := toPascalCase(tt.input)
			if got != tt.wantPascal {
				t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, got, tt.wantPascal)
			}
		})
		t.Run("camel_"+tt.input, func(t *testing.T) {
			got := toCamelCase(tt.input)
			if got != tt.wantCamel {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, got, tt.wantCamel)
			}
		})
	}
}

func TestNative_FilenameToStem(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"customers.sql", "customers"},
		{"queries.sql", "queries"},
		{"my_queries.sql", "my_queries"},
		{"path/to/queries.sql", "queries"},
		{"noext", "noext"},
		{".hidden", ""}, // filepath.Ext(".hidden") = ".hidden", so stem is empty
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := filenameToStem(tt.input)
			if got != tt.want {
				t.Errorf("filenameToStem(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNative_BuildParamList(t *testing.T) {
	t.Run("empty params", func(t *testing.T) {
		got := buildParamList(nil)
		if got != "" {
			t.Errorf("buildParamList(nil) = %q, want %q", got, "")
		}
	})

	t.Run("single param", func(t *testing.T) {
		params := []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}}
		got := buildParamList(params)
		if got != "inputParsed.data.id" {
			t.Errorf("buildParamList = %q, want %q", got, "inputParsed.data.id")
		}
	})

	t.Run("multiple params", func(t *testing.T) {
		params := []models.Param{
			{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}},
			{Name: "name", Position: 2, Type: models.SqlType{Name: "text"}},
		}
		got := buildParamList(params)
		if got != "inputParsed.data.id, inputParsed.data.name" {
			t.Errorf("buildParamList = %q, want %q", got, "inputParsed.data.id, inputParsed.data.name")
		}
	})

	t.Run("snake_case param names are camelCased", func(t *testing.T) {
		params := []models.Param{
			{Name: "user_id", Position: 1, Type: models.SqlType{Name: "integer"}},
		}
		got := buildParamList(params)
		if got != "inputParsed.data.userId" {
			t.Errorf("buildParamList = %q, want %q", got, "inputParsed.data.userId")
		}
	})
}

func TestNative_Build_QueryImportExtension(t *testing.T) {
	ext := ".ts"
	cfg := config.Config{Builder: "native", Driver: "pg", Validator: "zod", ImportExtension: &ext}
	n := New(cfg)
	log := logger.New(false)

	queries := []models.Query{oneQuery()}
	files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var queriesFile *File
	for i := range files {
		if files[i].Name == "customersQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		t.Fatal("expected customersQueries.ts")
	}

	content := string(queriesFile.Content)

	// .ts extension should be applied to all local imports
	if !strings.Contains(content, `"./models.ts"`) {
		t.Errorf("expected .ts extension on models import, got:\n%s", content)
	}
	if !strings.Contains(content, `"./customersRequests.ts"`) {
		t.Errorf("expected .ts extension on requests import, got:\n%s", content)
	}
	if !strings.Contains(content, `"./customersResponses.ts"`) {
		t.Errorf("expected .ts extension on responses import, got:\n%s", content)
	}
}

func TestNative_Build_EnumUnionGeneration(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	catalog := &models.Catalog{
		Enums: []models.Enum{
			{
				Name: "order_status",
				Values: []models.EnumValue{
					{Name: "Pending", Value: "pending"},
					{Name: "Confirmed", Value: "confirmed"},
					{Name: "Cancelled", Value: "cancelled"},
				},
			},
		},
	}

	q := models.Query{
		Name:         "ListOrdersByStatus",
		SQL:          "SELECT id, status FROM orders WHERE status = $1",
		RewrittenSQL: "SELECT id, status FROM orders WHERE status = $1",
		Command:      ":many",
		Filename:     "orders.sql",
		Params: []models.Param{
			{Name: "status", Position: 1, Type: models.SqlType{Name: "order_status", IsEnum: true}},
		},
		Results: []models.ResultField{
			{Name: "id", Type: models.SqlType{Name: "integer"}},
			{Name: "status", Type: models.SqlType{Name: "order_status", IsEnum: true}},
		},
	}

	files, err := n.Build(catalog, []models.Query{q}, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var requestsFile, responsesFile *File
	for i := range files {
		switch files[i].Name {
		case "ordersRequests.ts":
			requestsFile = &files[i]
		case "ordersResponses.ts":
			responsesFile = &files[i]
		}
	}
	if requestsFile == nil {
		t.Fatal("expected ordersRequests.ts")
	}
	if responsesFile == nil {
		t.Fatal("expected ordersResponses.ts")
	}

	wantUnion := `z.union([z.literal("pending"), z.literal("confirmed"), z.literal("cancelled")])`

	t.Run("enum param generates z.union literal", func(t *testing.T) {
		if !strings.Contains(string(requestsFile.Content), wantUnion) {
			t.Errorf("expected enum union in requests, got:\n%s", requestsFile.Content)
		}
	})

	t.Run("enum result generates z.union literal", func(t *testing.T) {
		if !strings.Contains(string(responsesFile.Content), wantUnion) {
			t.Errorf("expected enum union in responses, got:\n%s", responsesFile.Content)
		}
	})
}

func TestNative_Build_EnumUnion_EdgeCases(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	t.Run("single enum value generates z.literal", func(t *testing.T) {
		catalog := &models.Catalog{
			Enums: []models.Enum{
				{Name: "singleton", Values: []models.EnumValue{{Name: "Only", Value: "only"}}},
			},
		}
		q := models.Query{
			Name: "GetSingleton", SQL: "SELECT id FROM t WHERE e = $1",
			RewrittenSQL: "SELECT id FROM t WHERE e = $1",
			Command:      ":one", Filename: "t.sql",
			Params:  []models.Param{{Name: "e", Position: 1, Type: models.SqlType{Name: "singleton", IsEnum: true}}},
			Results: []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		}
		files, err := n.Build(catalog, []models.Query{q}, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var rf *File
		for i := range files {
			if files[i].Name == "tRequests.ts" {
				rf = &files[i]
			}
		}
		if rf == nil {
			t.Fatal("expected tRequests.ts")
		}
		if !strings.Contains(string(rf.Content), `z.literal("only")`) {
			t.Errorf("expected z.literal for single-value enum, got:\n%s", rf.Content)
		}
	})

	t.Run("zero enum values generates z.never", func(t *testing.T) {
		catalog := &models.Catalog{
			Enums: []models.Enum{
				{Name: "empty_enum", Values: []models.EnumValue{}},
			},
		}
		q := models.Query{
			Name: "GetEmpty", SQL: "SELECT id FROM t WHERE e = $1",
			RewrittenSQL: "SELECT id FROM t WHERE e = $1",
			Command:      ":one", Filename: "t.sql",
			Params:  []models.Param{{Name: "e", Position: 1, Type: models.SqlType{Name: "empty_enum", IsEnum: true}}},
			Results: []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		}
		files, err := n.Build(catalog, []models.Query{q}, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var rf *File
		for i := range files {
			if files[i].Name == "tRequests.ts" {
				rf = &files[i]
			}
		}
		if rf == nil {
			t.Fatal("expected tRequests.ts")
		}
		if !strings.Contains(string(rf.Content), `z.never()`) {
			t.Errorf("expected z.never() for zero-value enum, got:\n%s", rf.Content)
		}
	})

	t.Run("nullable enum param generates .optional()", func(t *testing.T) {
		catalog := &models.Catalog{
			Enums: []models.Enum{
				{Name: "order_status", Values: []models.EnumValue{
					{Name: "Pending", Value: "pending"},
					{Name: "Done", Value: "done"},
				}},
			},
		}
		q := models.Query{
			Name: "FilterOrders", SQL: "SELECT id FROM t WHERE status = $1",
			RewrittenSQL: "SELECT id FROM t WHERE status = $1",
			Command:      ":many", Filename: "t.sql",
			Params: []models.Param{
				{Name: "status", Position: 1, Type: models.SqlType{Name: "order_status", IsEnum: true, IsNullable: true}},
			},
			Results: []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		}
		files, err := n.Build(catalog, []models.Query{q}, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var rf *File
		for i := range files {
			if files[i].Name == "tRequests.ts" {
				rf = &files[i]
			}
		}
		if rf == nil {
			t.Fatal("expected tRequests.ts")
		}
		content := string(rf.Content)
		if !strings.Contains(content, `z.union([z.literal("pending"), z.literal("done")]).optional()`) {
			t.Errorf("expected z.union(...).optional() for nullable enum param, got:\n%s", content)
		}
	})

	t.Run("nullable enum result generates .nullable()", func(t *testing.T) {
		catalog := &models.Catalog{
			Enums: []models.Enum{
				{Name: "order_status", Values: []models.EnumValue{
					{Name: "Pending", Value: "pending"},
					{Name: "Done", Value: "done"},
				}},
			},
		}
		q := models.Query{
			Name: "GetStatus", SQL: "SELECT status FROM t WHERE id = $1",
			RewrittenSQL: "SELECT status FROM t WHERE id = $1",
			Command:      ":one", Filename: "t.sql",
			Params: []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}},
			Results: []models.ResultField{
				{Name: "status", Type: models.SqlType{Name: "order_status", IsEnum: true, IsNullable: true}},
			},
		}
		files, err := n.Build(catalog, []models.Query{q}, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var rf *File
		for i := range files {
			if files[i].Name == "tResponses.ts" {
				rf = &files[i]
			}
		}
		if rf == nil {
			t.Fatal("expected tResponses.ts")
		}
		content := string(rf.Content)
		if !strings.Contains(content, `z.union([z.literal("pending"), z.literal("done")]).nullable()`) {
			t.Errorf("expected z.union(...).nullable() for nullable enum result, got:\n%s", content)
		}
	})

	t.Run("array of enum generates z.array(z.union(...))", func(t *testing.T) {
		catalog := &models.Catalog{
			Enums: []models.Enum{
				{Name: "order_status", Values: []models.EnumValue{
					{Name: "Pending", Value: "pending"},
					{Name: "Done", Value: "done"},
				}},
			},
		}
		q := models.Query{
			Name: "GetByStatuses", SQL: "SELECT id FROM t WHERE status = ANY($1)",
			RewrittenSQL: "SELECT id FROM t WHERE status = ANY($1)",
			Command:      ":many", Filename: "t.sql",
			Params: []models.Param{
				{Name: "statuses", Position: 1, Type: models.SqlType{Name: "order_status", IsEnum: true, IsArray: true}},
			},
			Results: []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		}
		files, err := n.Build(catalog, []models.Query{q}, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var rf *File
		for i := range files {
			if files[i].Name == "tRequests.ts" {
				rf = &files[i]
			}
		}
		if rf == nil {
			t.Fatal("expected tRequests.ts")
		}
		content := string(rf.Content)
		if !strings.Contains(content, `z.array(z.union([z.literal("pending"), z.literal("done")]))`) {
			t.Errorf("expected z.array(z.union(...)) for array enum, got:\n%s", content)
		}
	})

	t.Run("enum not in catalog falls back to z.string()", func(t *testing.T) {
		emptyCatalog := &models.Catalog{}
		q := models.Query{
			Name: "GetThing", SQL: "SELECT id FROM t WHERE e = $1",
			RewrittenSQL: "SELECT id FROM t WHERE e = $1",
			Command:      ":one", Filename: "t.sql",
			Params:  []models.Param{{Name: "e", Position: 1, Type: models.SqlType{Name: "unknown_enum", IsEnum: true}}},
			Results: []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
		}
		files, err := n.Build(emptyCatalog, []models.Query{q}, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		var rf *File
		for i := range files {
			if files[i].Name == "tRequests.ts" {
				rf = &files[i]
			}
		}
		if rf == nil {
			t.Fatal("expected tRequests.ts")
		}
		if !strings.Contains(string(rf.Content), `z.string()`) {
			t.Errorf("expected z.string() fallback for unknown enum, got:\n%s", rf.Content)
		}
	})
}

func TestNative_Build_QueryWithEmptyFilename(t *testing.T) {
	n := New(defaultConfig())
	log := logger.New(false)

	q := models.Query{
		Name:         "GetThing",
		SQL:          "SELECT id FROM things WHERE id = $1",
		RewrittenSQL: "SELECT id FROM things WHERE id = $1",
		Command:      ":one",
		Filename:     "", // Empty filename
		Params:       []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}},
		Results:      []models.ResultField{{Name: "id", Type: models.SqlType{Name: "integer"}}},
	}

	files, err := n.Build(defaultCatalog(), []models.Query{q}, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should default to "queries.sql" -> "queries" stem
	var queriesFile *File
	for i := range files {
		if files[i].Name == "queriesQueries.ts" {
			queriesFile = &files[i]
			break
		}
	}
	if queriesFile == nil {
		names := make([]string, len(files))
		for i, f := range files {
			names[i] = f.Name
		}
		t.Fatalf("expected queriesQueries.ts for empty filename, got files: %v", names)
	}
}
