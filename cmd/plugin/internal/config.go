package internal

import (
	"encoding/json"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

const (
	enginePostgres = "postgresql"
)

type Config struct {
	Output string `json:"output"`
}

func GetConfig(req *plugin.GenerateRequest) (Config, error) {
	var conf Config
	if len(req.PluginOptions) > 0 {
		if err := json.Unmarshal(req.PluginOptions, &conf); err != nil {
			return conf, err
		}
	}
	return conf, nil
}
