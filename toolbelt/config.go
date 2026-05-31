package toolbelt

import (
	"encoding/json"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// ParseJSONConfig unmarshals sqlc plugin options into the plugin-specific config type.
func ParseJSONConfig[T any](req *plugin.GenerateRequest) (T, error) {
	var cfg T
	if len(req.PluginOptions) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(req.PluginOptions, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
