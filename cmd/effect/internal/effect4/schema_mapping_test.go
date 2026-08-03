package effect4

import (
	"testing"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
)

func TestEffect4_NullableParamUsesRequiredOptionSchema(t *testing.T) {
	e := New(defaultConfig())

	got := e.sqlTypeToEffectSchemaForParams(models.SqlType{Name: "text", IsNullable: true})
	want := "Schema.OptionFromNullOr(Schema.String)"

	if got.Schema != want {
		t.Fatalf("nullable parameter schema = %q, want %q", got.Schema, want)
	}
}

func TestEffect4_BigIntTypesUseBuiltInSchema(t *testing.T) {
	e := New(defaultConfig())

	for _, name := range []string{"bigint", "int8", "bigserial", "serial8"} {
		t.Run(name, func(t *testing.T) {
			got := e.sqlTypeToEffectSchemaBase(models.SqlType{Name: name})
			if got.Schema != "Schema.BigIntFromString" {
				t.Fatalf("bigint schema = %q, want %q", got.Schema, "Schema.BigIntFromString")
			}
			if len(got.ModelImports) != 0 {
				t.Fatalf("expected no model imports, got %#v", got.ModelImports)
			}
		})
	}
}
