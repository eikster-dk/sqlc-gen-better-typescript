package native

import (
	"testing"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

func defaultCatalog() *models.Catalog {
	return &models.Catalog{}
}

func TestNative_Build_Stub(t *testing.T) {
	cfg := config.Config{Builder: "native", Driver: "pg", Validator: "zod"}
	n := New(cfg)
	log := logger.New(false)

	t.Run("returns empty file list without error", func(t *testing.T) {
		files, err := n.Build(defaultCatalog(), nil, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})

	t.Run("empty queries slice returns no error", func(t *testing.T) {
		files, err := n.Build(defaultCatalog(), []models.Query{}, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})

	t.Run("catalog with tables and enums returns no error", func(t *testing.T) {
		catalog := &models.Catalog{
			Tables: []models.Table{
				{Name: "users", Columns: []models.Column{{Name: "id", Type: models.SqlType{Name: "serial"}}}},
			},
			Enums: []models.Enum{
				{Name: "role", Values: []models.EnumValue{{Name: "Admin", Value: "admin"}}},
			},
		}
		files, err := n.Build(catalog, nil, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files in stub phase, got %d", len(files))
		}
	})

	t.Run("non-empty queries slice returns no error", func(t *testing.T) {
		queries := []models.Query{
			{Name: "GetUser", SQL: "SELECT id FROM users WHERE id = $1", Command: ":one"},
			{Name: "ListUsers", SQL: "SELECT id FROM users", Command: ":many"},
		}
		files, err := n.Build(defaultCatalog(), queries, log, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files in stub phase, got %d", len(files))
		}
	})

	t.Run("sqlc version is accepted without error", func(t *testing.T) {
		for _, version := range []string{"", "1.0.0", "2.3.4-beta", "v1.27.0"} {
			_, err := n.Build(defaultCatalog(), nil, log, version)
			if err != nil {
				t.Errorf("unexpected error for sqlcVersion=%q: %v", version, err)
			}
		}
	})

	t.Run("debug logging enabled does not cause error", func(t *testing.T) {
		debugLog := logger.New(true)
		files, err := n.Build(defaultCatalog(), nil, debugLog, "1.0.0")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %d", len(files))
		}
	})
}

func TestNew(t *testing.T) {
	t.Run("New returns a non-nil Native", func(t *testing.T) {
		cfg := config.Config{Builder: "native"}
		n := New(cfg)
		if n == nil {
			t.Fatal("expected non-nil *Native")
		}
	})

	t.Run("config is stored correctly", func(t *testing.T) {
		cfg := config.Config{Builder: "native", Driver: "pg2", Validator: "valibot"}
		n := New(cfg)
		if n.cfg.Driver != "pg2" {
			t.Errorf("expected driver %q, got %q", "pg2", n.cfg.Driver)
		}
		if n.cfg.Validator != "valibot" {
			t.Errorf("expected validator %q, got %q", "valibot", n.cfg.Validator)
		}
	})
}
