package main

import (
	"context"
	"encoding/json"

	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func main() {
	codegen.Run(Generate)
}

func Generate(_ context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {

	bytes, _ := json.MarshalIndent(req.Queries, "", "  ")
	test := &plugin.File{
		Name:     "test.json",
		Contents: bytes,
	}

	return &plugin.GenerateResponse{
		Files: []*plugin.File{test},
	}, nil
}
