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
