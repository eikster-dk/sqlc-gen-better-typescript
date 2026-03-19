package main

import (
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal"
	"github.com/sqlc-dev/plugin-sdk-go/codegen"
)

func main() {
	codegen.Run(internal.Generate)
}
