package config

import (
	"strings"
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestParse_Defaults(t *testing.T) {
	cfg, err := Parse(&plugin.GenerateRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.ImportExtension != ".js" {
		t.Errorf("ImportExtension = %q, want .js", cfg.ImportExtension)
	}
	if cfg.Driver != "pg" {
		t.Errorf("Driver = %q, want pg", cfg.Driver)
	}
	if cfg.Validator != "zod" {
		t.Errorf("Validator = %q, want zod", cfg.Validator)
	}
	if cfg.Debug {
		t.Error("Debug = true, want false")
	}
	if cfg.DebugDir != "" {
		t.Errorf("DebugDir = %q, want empty", cfg.DebugDir)
	}
}

func TestParse_ExplicitOptions(t *testing.T) {
	req := &plugin.GenerateRequest{PluginOptions: []byte(`{
		"debug": true,
		"debug_dir": "tmp/sqlc-debug",
		"import_extension": ".ts",
		"driver": "pg",
		"validator": "zod"
	}`)}

	cfg, err := Parse(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.Debug {
		t.Error("Debug = false, want true")
	}
	if cfg.DebugDir != "tmp/sqlc-debug" {
		t.Errorf("DebugDir = %q, want tmp/sqlc-debug", cfg.DebugDir)
	}
	if cfg.ImportExtension != ".ts" {
		t.Errorf("ImportExtension = %q, want .ts", cfg.ImportExtension)
	}
	if cfg.Driver != "pg" {
		t.Errorf("Driver = %q, want pg", cfg.Driver)
	}
	if cfg.Validator != "zod" {
		t.Errorf("Validator = %q, want zod", cfg.Validator)
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse(&plugin.GenerateRequest{PluginOptions: []byte(`{"driver":`)})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		req     *plugin.GenerateRequest
		wantErr string
	}{
		{
			name: "valid defaults",
			cfg:  Config{ImportExtension: ".js", Driver: "pg", Validator: "zod"},
			req:  &plugin.GenerateRequest{Settings: &plugin.Settings{Engine: "postgresql"}},
		},
		{
			name: "valid ts imports",
			cfg:  Config{ImportExtension: ".ts", Driver: "pg", Validator: "zod"},
			req:  &plugin.GenerateRequest{Settings: &plugin.Settings{Engine: "postgresql"}},
		},
		{
			name:    "unsupported engine",
			cfg:     Config{ImportExtension: ".js", Driver: "pg", Validator: "zod"},
			req:     &plugin.GenerateRequest{Settings: &plugin.Settings{Engine: "mysql"}},
			wantErr: "engine value is \"mysql\"",
		},
		{
			name:    "unsupported import extension",
			cfg:     Config{ImportExtension: ".mjs", Driver: "pg", Validator: "zod"},
			req:     &plugin.GenerateRequest{Settings: &plugin.Settings{Engine: "postgresql"}},
			wantErr: "import_extension value is .mjs",
		},
		{
			name:    "unsupported driver",
			cfg:     Config{ImportExtension: ".js", Driver: "mysql", Validator: "zod"},
			req:     &plugin.GenerateRequest{Settings: &plugin.Settings{Engine: "postgresql"}},
			wantErr: "driver value is mysql",
		},
		{
			name:    "unsupported validator",
			cfg:     Config{ImportExtension: ".js", Driver: "pg", Validator: "valibot"},
			req:     &plugin.GenerateRequest{Settings: &plugin.Settings{Engine: "postgresql"}},
			wantErr: "validator value is valibot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg, tt.req)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("error = %q, want substring %q", err.Error(), tt.wantErr)
			}
		})
	}
}
