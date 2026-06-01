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
			_, err := buildEffect(e, &models.Catalog{}, queries, log, "1.0.0")
			if err != nil {
				t.Fatalf("unexpected error for command %s: %v", cmd, err)
			}
		})
	}
}
