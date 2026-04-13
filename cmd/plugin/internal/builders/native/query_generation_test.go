package native

import (
	"strings"
	"testing"

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
		{"text", "z.string()"},
		{"varchar", "z.string()"},
		{"boolean", "z.boolean()"},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			got := n.zodBaseType(models.SqlType{Name: tt.typeName})
			if got != tt.want {
				t.Errorf("zodBaseType(%q) = %q, want %q", tt.typeName, got, tt.want)
			}
		})
	}
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
