package effect4

import (
	"context"
	"strings"
	"testing"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/effect/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
)

func defaultConfig() config.Config {
	return config.Config{}
}

func buildEffect(e *Effect4, catalog *models.Catalog, queries []models.Query, log *logger.Logger, sqlcVersion string) ([]toolbelt.File, error) {
	return e.Build(toolbelt.BuildContext{Context: context.Background(), Catalog: catalog, Queries: queries, Logger: log, SqlcVersion: sqlcVersion})
}

func TestEffect4_Build_RejectsUnsupportedCommands(t *testing.T) {
	e := New(defaultConfig())
	log := logger.New(false)

	unsupported := []string{":copyfrom", ":batchexec", ":batchone", ":batchmany"}
	for _, cmd := range unsupported {
		cmd := cmd
		t.Run("rejects "+cmd, func(t *testing.T) {
			queries := []models.Query{
				{Name: "BadQuery", SQL: "SELECT 1", Command: cmd, Filename: "queries.sql"},
			}
			_, err := buildEffect(e, &models.Catalog{}, queries, log, "1.0.0")
			if err == nil {
				t.Fatalf("expected error for command %s, got nil", cmd)
			}
			if !strings.Contains(err.Error(), cmd) {
				t.Errorf("expected error to mention %q, got: %s", cmd, err.Error())
			}
			if !strings.Contains(err.Error(), "BadQuery") {
				t.Errorf("expected error to mention query name %q, got: %s", "BadQuery", err.Error())
			}
		})
	}
}

func findFile(files []toolbelt.File, name string) *toolbelt.File {
	for _, f := range files {
		if f.Name == name {
			return &f
		}
	}
	return nil
}

func TestEffect4_Build_ExecResult(t *testing.T) {
	e := New(defaultConfig())
	log := logger.New(false)

	queries := []models.Query{
		{Name: "DeleteUser", SQL: "DELETE FROM users WHERE id = $1", Command: ":execresult", Filename: "queries.sql"},
	}
	files, err := buildEffect(e, &models.Catalog{}, queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("repository uses execResult helper not SqlSchema.findAll", func(t *testing.T) {
		repo := findFile(files, "QueriesRepository.ts")
		if repo == nil {
			t.Fatal("expected QueriesRepository.ts in output")
		}
		content := string(repo.Content)
		if strings.Contains(content, "SqlSchema.findAll") {
			t.Error("repository should not use SqlSchema.findAll for :execresult")
		}
		if !strings.Contains(content, "execResult") {
			t.Error("repository should use execResult helper for :execresult")
		}
	})

	t.Run("repository imports execResult from models", func(t *testing.T) {
		repo := findFile(files, "QueriesRepository.ts")
		if repo == nil {
			t.Fatal("expected QueriesRepository.ts in output")
		}
		content := string(repo.Content)
		if !strings.Contains(content, "execResult") || !strings.Contains(content, "models") {
			t.Error("repository should import execResult from models")
		}
	})

	t.Run("models includes SqlExecResult schema and execResult helper", func(t *testing.T) {
		modelsFile := findFile(files, "models.ts")
		if modelsFile == nil {
			t.Fatal("expected models.ts in output")
		}
		content := string(modelsFile.Content)
		if !strings.Contains(content, "SqlExecResult") {
			t.Error("models should define SqlExecResult schema")
		}
		if !strings.Contains(content, "execResult") {
			t.Error("models should define execResult helper")
		}
	})
}

func TestEffect4_Build_AcceptsSupportedCommands(t *testing.T) {
	e := New(defaultConfig())
	log := logger.New(false)

	supported := []string{":one", ":many", ":exec", ":execrows", ":execresult"}
	for _, cmd := range supported {
		cmd := cmd
		t.Run("accepts "+cmd, func(t *testing.T) {
			queries := []models.Query{
				{Name: "GoodQuery", SQL: "SELECT 1", Command: cmd, Filename: "queries.sql"},
			}
			files, err := buildEffect(e, &models.Catalog{}, queries, log, "1.0.0")
			if err != nil {
				t.Fatalf("unexpected error for command %s: %v", cmd, err)
			}
			repositoryFile := findFile(files, "QueriesRepository.ts")
			if repositoryFile == nil {
				t.Fatal("expected QueriesRepository.ts in output")
			}
			repository := string(repositoryFile.Content)
			for _, expected := range []string{
				"goodQuery: Effect.fn(\"QueriesRepository.goodQuery\")(",
			} {
				if !strings.Contains(repository, expected) {
					t.Errorf("expected repository output to contain %q", expected)
				}
			}
			for _, unexpected := range []string{"TaggedErrorClass", "RepositoryError", "Effect.mapError"} {
				if strings.Contains(repository, unexpected) {
					t.Errorf("expected repository output not to contain %q", unexpected)
				}
			}
		})
	}
}

func TestEffect4_Build_UsesBuiltInBigIntSchema(t *testing.T) {
	e := New(defaultConfig())
	log := logger.New(false)
	queries := []models.Query{{
		Name: "FindTotal", Command: ":one", Filename: "totals.sql",
		SQL:    "SELECT total FROM totals WHERE total = $1",
		Params: []models.Param{{Name: "total", Position: 1, Type: models.SqlType{Name: "bigint"}}},
		Results: []models.ResultField{{
			Name: "total", OriginalName: "total", Type: models.SqlType{Name: "bigint"},
		}},
	}}

	files, err := buildEffect(e, &models.Catalog{}, queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, name := range []string{"TotalsRepositoryRequest.ts", "TotalsRepositoryResponse.ts"} {
		file := findFile(files, name)
		if file == nil {
			t.Fatalf("expected %s in output", name)
		}
		content := string(file.Content)
		if !strings.Contains(content, "Schema.BigIntFromString") {
			t.Errorf("expected %s to use Schema.BigIntFromString", name)
		}
		if strings.Contains(content, `from "./models"`) {
			t.Errorf("expected %s not to import bigint schema from models", name)
		}
	}

	modelsFile := findFile(files, "models.ts")
	if modelsFile == nil {
		t.Fatal("expected models.ts in output")
	}
	modelsContent := string(modelsFile.Content)
	for _, unexpected := range []string{"export const BigIntFromString", "SchemaGetter"} {
		if strings.Contains(modelsContent, unexpected) {
			t.Errorf("expected models output not to contain %q", unexpected)
		}
	}
}

func TestEffect4_Build_ProjectsEmbeddedRowsOutsideSchema(t *testing.T) {
	e := New(defaultConfig())
	log := logger.New(false)
	catalog := &models.Catalog{Tables: []models.Table{
		{Name: "orders", Columns: []models.Column{
			{Name: "id", Type: models.SqlType{Name: "integer"}},
			{Name: "note", Type: models.SqlType{Name: "text", IsNullable: true}},
		}},
		{Name: "customers", Columns: []models.Column{
			{Name: "id", Type: models.SqlType{Name: "integer"}},
		}},
	}}

	embedResults := []models.ResultField{
		{Name: "orders_id", OriginalName: "id", Type: models.SqlType{Name: "integer"}, Table: "orders", EmbedTable: "orders", IsAliased: true},
		{Name: "orders_note", OriginalName: "note", Type: models.SqlType{Name: "text", IsNullable: true}, Table: "orders", EmbedTable: "orders", IsAliased: true},
		{Name: "customers_id", OriginalName: "id", Type: models.SqlType{Name: "integer"}, Table: "customers", EmbedTable: "customers", IsAliased: true},
	}
	embedGroups := []models.EmbedGroup{
		{TableName: "orders", Fields: embedResults[:2]},
		{TableName: "customers", Fields: embedResults[2:]},
	}
	queries := []models.Query{
		{
			Name: "GetOrderWithCustomer", Command: ":one", Filename: "embed.sql",
			SQL:     "SELECT orders.id AS orders_id, orders.note AS orders_note, customers.id AS customers_id FROM orders JOIN customers ON customers.id = orders.id WHERE orders.id = $1",
			Params:  []models.Param{{Name: "id", Position: 1, Type: models.SqlType{Name: "integer"}}},
			Results: embedResults, HasEmbeds: true, EmbedGroups: embedGroups,
		},
		{
			Name: "ListOrdersWithCustomer", Command: ":many", Filename: "embed.sql",
			SQL:     "SELECT orders.id AS orders_id, orders.note AS orders_note, customers.id AS customers_id FROM orders JOIN customers ON customers.id = orders.id",
			Results: embedResults, HasEmbeds: true, EmbedGroups: embedGroups,
		},
	}

	files, err := buildEffect(e, catalog, queries, log, "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	responseFile := findFile(files, "EmbedRepositoryResponse.ts")
	if responseFile == nil {
		t.Fatal("expected EmbedRepositoryResponse.ts in output")
	}
	response := string(responseFile.Content)
	for _, expected := range []string{
		"export const GetOrderWithCustomerRow = Schema.Struct({",
		"orders_note: Schema.OptionFromNullOr(Schema.String)",
		"export type GetOrderWithCustomerRow = typeof GetOrderWithCustomerRow.Type",
		"export const GetOrderWithCustomerResult = Schema.Struct({",
		"export const mapGetOrderWithCustomerRowToResult = (",
		"row: GetOrderWithCustomerRow",
		"): GetOrderWithCustomerResult => ({",
		"export const ListOrdersWithCustomerRow = Schema.Struct({",
		"export const mapListOrdersWithCustomerRowToResult = (",
	} {
		if !strings.Contains(response, expected) {
			t.Errorf("expected response output to contain %q", expected)
		}
	}
	for _, unexpected := range []string{"SchemaTransformation", "Schema.decodeTo", "Encode not supported"} {
		if strings.Contains(response, unexpected) {
			t.Errorf("expected response output not to contain %q", unexpected)
		}
	}

	repositoryFile := findFile(files, "EmbedRepository.ts")
	if repositoryFile == nil {
		t.Fatal("expected EmbedRepository.ts in output")
	}
	repository := string(repositoryFile.Content)
	for _, expected := range []string{
		"Result: GetOrderWithCustomerRow",
		"Option.map(mapGetOrderWithCustomerRowToResult)",
		"Result: ListOrdersWithCustomerRow",
		"rows.map(mapListOrdersWithCustomerRowToResult)",
		"getOrderWithCustomer: Effect.fn(\"EmbedRepository.getOrderWithCustomer\")(",
		"listOrdersWithCustomer: Effect.fn(\"EmbedRepository.listOrdersWithCustomer\")(",
	} {
		if !strings.Contains(repository, expected) {
			t.Errorf("expected repository output to contain %q", expected)
		}
	}
	for _, unexpected := range []string{"Result: GetOrderWithCustomerResult", "Result: ListOrdersWithCustomerResult"} {
		if strings.Contains(repository, unexpected) {
			t.Errorf("expected repository output not to contain %q", unexpected)
		}
	}
}
