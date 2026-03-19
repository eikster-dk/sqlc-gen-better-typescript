package main

import (
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal"
	"github.com/sqlc-dev/plugin-sdk-go/codegen"
)

func main() {
	codegen.Run(internal.Generate)
}
