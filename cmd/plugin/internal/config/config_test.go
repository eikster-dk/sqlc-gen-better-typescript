package config

import (
	"encoding/json"
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func makeRequest(options map[string]any) *plugin.GenerateRequest {
	raw, err := json.Marshal(options)
	if err != nil {
		panic(err)
	}
	return &plugin.GenerateRequest{PluginOptions: raw}
}

func makeEmptyRequest() *plugin.GenerateRequest {
	return &plugin.GenerateRequest{}
}

func TestGetConfig_Defaults(t *testing.T) {
	t.Run("empty options apply all defaults", func(t *testing.T) {
		cfg, err := GetConfig(makeEmptyRequest())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Builder != "native" {
			t.Errorf("expected default builder %q, got %q", "native", cfg.Builder)
		}
		if cfg.ImportExtension == nil || *cfg.ImportExtension != ".js" {
			got := "<nil>"
			if cfg.ImportExtension != nil {
				got = *cfg.ImportExtension
			}
			t.Errorf("expected default import_extension %q, got %q", ".js", got)
		}
		if cfg.Driver != "pg" {
			t.Errorf("expected default driver %q, got %q", "pg", cfg.Driver)
		}
		if cfg.Validator != "zod" {
			t.Errorf("expected default validator %q, got %q", "zod", cfg.Validator)
		}
	})

	t.Run("explicit builder is preserved", func(t *testing.T) {
		cfg, err := GetConfig(makeRequest(map[string]any{"builder": "effect-v4-unstable"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Builder != "effect-v4-unstable" {
			t.Errorf("expected builder %q, got %q", "effect-v4-unstable", cfg.Builder)
		}
	})

	t.Run("explicit native builder is preserved", func(t *testing.T) {
		cfg, err := GetConfig(makeRequest(map[string]any{"builder": "native"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Builder != "native" {
			t.Errorf("expected builder %q, got %q", "native", cfg.Builder)
		}
	})

	t.Run("explicit import_extension .ts is preserved", func(t *testing.T) {
		cfg, err := GetConfig(makeRequest(map[string]any{"import_extension": ".ts"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ImportExtension == nil || *cfg.ImportExtension != ".ts" {
			got := "<nil>"
			if cfg.ImportExtension != nil {
				got = *cfg.ImportExtension
			}
			t.Errorf("expected import_extension %q, got %q", ".ts", got)
		}
	})

	t.Run("explicit import_extension .js is preserved", func(t *testing.T) {
		cfg, err := GetConfig(makeRequest(map[string]any{"import_extension": ".js"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ImportExtension == nil || *cfg.ImportExtension != ".js" {
			got := "<nil>"
			if cfg.ImportExtension != nil {
				got = *cfg.ImportExtension
			}
			t.Errorf("expected import_extension %q, got %q", ".js", got)
		}
	})

	t.Run("explicit empty import_extension is preserved", func(t *testing.T) {
		// Setting import_extension: "" explicitly should NOT be overridden to ".js".
		// Empty string is a valid value meaning no extension.
		cfg, err := GetConfig(makeRequest(map[string]any{"import_extension": ""}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ImportExtension == nil || *cfg.ImportExtension != "" {
			got := "<nil>"
			if cfg.ImportExtension != nil {
				got = *cfg.ImportExtension
			}
			t.Errorf("explicit empty import_extension should be preserved as %q, got %q", "", got)
		}
	})

	t.Run("omitted import_extension defaults to .js", func(t *testing.T) {
		cfg, err := GetConfig(makeEmptyRequest())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ImportExtension == nil || *cfg.ImportExtension != ".js" {
			got := "<nil>"
			if cfg.ImportExtension != nil {
				got = *cfg.ImportExtension
			}
			t.Errorf("omitted import_extension should default to %q, got %q", ".js", got)
		}
	})

	t.Run("explicit driver is preserved", func(t *testing.T) {
		cfg, err := GetConfig(makeRequest(map[string]any{"driver": "pg2"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Driver != "pg2" {
			t.Errorf("expected driver %q, got %q", "pg2", cfg.Driver)
		}
	})

	t.Run("explicit validator is preserved", func(t *testing.T) {
		cfg, err := GetConfig(makeRequest(map[string]any{"validator": "valibot"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Validator != "valibot" {
			t.Errorf("expected validator %q, got %q", "valibot", cfg.Validator)
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		req := &plugin.GenerateRequest{PluginOptions: []byte(`{not valid json`)}
		_, err := GetConfig(req)
		if err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})

	t.Run("nil plugin options uses defaults", func(t *testing.T) {
		req := &plugin.GenerateRequest{PluginOptions: nil}
		cfg, err := GetConfig(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Builder != "native" {
			t.Errorf("expected default builder %q, got %q", "native", cfg.Builder)
		}
	})

	t.Run("empty JSON object uses defaults", func(t *testing.T) {
		req := &plugin.GenerateRequest{PluginOptions: []byte(`{}`)}
		cfg, err := GetConfig(req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Builder != "native" {
			t.Errorf("expected default builder %q, got %q", "native", cfg.Builder)
		}
		if cfg.Driver != "pg" {
			t.Errorf("expected default driver %q, got %q", "pg", cfg.Driver)
		}
		if cfg.Validator != "zod" {
			t.Errorf("expected default validator %q, got %q", "zod", cfg.Validator)
		}
	})

	t.Run("debug flag is preserved", func(t *testing.T) {
		cfg, err := GetConfig(makeRequest(map[string]any{"debug": true}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.Debug {
			t.Error("expected debug=true to be preserved")
		}
	})

	t.Run("disable_template_literals flag is preserved", func(t *testing.T) {
		cfg, err := GetConfig(makeRequest(map[string]any{"disable_template_literals": true}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.DisableTemplateLiterals {
			t.Error("expected disable_template_literals=true to be preserved")
		}
	})
}
