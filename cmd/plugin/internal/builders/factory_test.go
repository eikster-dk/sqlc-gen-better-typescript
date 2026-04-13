package builders

import (
	"testing"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/config"
)

func TestNewBuilder(t *testing.T) {
	t.Run("native builder is created successfully", func(t *testing.T) {
		cfg := config.Config{Builder: "native"}
		b, err := NewBuilder(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if b == nil {
			t.Fatal("expected non-nil builder")
		}
	})

	t.Run("effect-v4-unstable builder is created successfully", func(t *testing.T) {
		cfg := config.Config{Builder: "effect-v4-unstable"}
		b, err := NewBuilder(cfg)
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if b == nil {
			t.Fatal("expected non-nil builder")
		}
	})

	t.Run("unknown builder returns error", func(t *testing.T) {
		cfg := config.Config{Builder: "does-not-exist"}
		b, err := NewBuilder(cfg)
		if err == nil {
			t.Error("expected error for unknown builder, got nil")
		}
		if b != nil {
			t.Error("expected nil builder on error")
		}
	})

	t.Run("empty builder string returns error", func(t *testing.T) {
		cfg := config.Config{Builder: ""}
		b, err := NewBuilder(cfg)
		if err == nil {
			t.Error("expected error for empty builder, got nil")
		}
		if b != nil {
			t.Error("expected nil builder on error")
		}
	})

	t.Run("error message contains the unknown builder name", func(t *testing.T) {
		cfg := config.Config{Builder: "my-custom-builder"}
		_, err := NewBuilder(cfg)
		if err == nil {
			t.Fatal("expected error")
		}
		if got := err.Error(); got == "" {
			t.Error("expected non-empty error message")
		}
	})
}
