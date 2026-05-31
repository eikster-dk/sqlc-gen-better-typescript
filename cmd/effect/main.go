package main

import (
	"context"

	effectinternal "github.com/eikster-dk/sqlc-gen-better-typescript/cmd/effect/internal"
	"github.com/sqlc-dev/plugin-sdk-go/codegen"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func main() {
	codegen.Run(func(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
		return effectinternal.Generate(ctx, req)
	})
}
