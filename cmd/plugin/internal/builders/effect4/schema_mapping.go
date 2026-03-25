package effect4

import (
	"fmt"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

func (e *Effect4) sqlTypeToEffectSchemaBase(t models.SqlType) SchemaExpr {
	var baseSchema string
	modelImports := []string{}

	switch strings.ToLower(t.Name) {
	case "serial", "serial4", "smallserial", "serial2", "integer", "int", "int4", "smallint", "int2":
		baseSchema = "Schema.Int"
	case "bigserial", "serial8", "bigint", "int8":
		baseSchema = "BigIntFromString"
		modelImports = append(modelImports, "BigIntFromString")
	case "float", "double precision", "float8", "real", "float4":
		baseSchema = "Schema.Number"
	case "numeric", "money", "time", "timetz", "interval", "text", "varchar", "bpchar", "string", "citext", "uuid", "inet", "cidr", "macaddr", "macaddr8", "ltree", "lquery", "ltxtquery":
		baseSchema = "Schema.String"
	case "boolean", "bool":
		baseSchema = "Schema.Boolean"
	case "json", "jsonb":
		baseSchema = "Schema.Unknown"
	case "bytea", "blob":
		baseSchema = "Schema.Uint8Array"
	case "date", "timestamp", "timestamptz":
		baseSchema = "Schema.Date"
	default:
		if t.IsEnum {
			baseSchema = fmt.Sprintf("%sSchema", toPascalCase(t.Name))
			modelImports = append(modelImports, baseSchema)
		} else {
			baseSchema = "Schema.String"
		}
	}

	if t.IsArray {
		baseSchema = fmt.Sprintf("Schema.Array(%s)", baseSchema)
	}

	return SchemaExpr{Schema: baseSchema, ModelImports: uniqueSorted(modelImports)}
}

func (e *Effect4) sqlTypeToEffectSchemaForParams(t models.SqlType) SchemaExpr {
	baseExpr := e.sqlTypeToEffectSchemaBase(t)
	schema := baseExpr.Schema
	if t.IsNullable {
		schema = fmt.Sprintf("Schema.optional(%s)", schema)
	}
	return SchemaExpr{Schema: schema, ModelImports: baseExpr.ModelImports}
}

func (e *Effect4) sqlTypeToEffectSchemaForResults(t models.SqlType) SchemaExpr {
	baseExpr := e.sqlTypeToEffectSchemaBase(t)
	schema := baseExpr.Schema
	if t.IsNullable {
		schema = fmt.Sprintf("Schema.OptionFromNullOr(%s)", schema)
	}
	return SchemaExpr{Schema: schema, ModelImports: baseExpr.ModelImports}
}
