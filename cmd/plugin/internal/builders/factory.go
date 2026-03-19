package builders

import (
	"fmt"

	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/config"
)

// NewBuilder creates a builder based on the config
func NewBuilder(cfg config.Config) (Builder, error) {
	switch cfg.Builder {
	case "effect-v4-unstable":
		return NewEffect4(cfg), nil
	default:
		return nil, fmt.Errorf("unknown builder: %s", cfg.Builder)
	}
}
