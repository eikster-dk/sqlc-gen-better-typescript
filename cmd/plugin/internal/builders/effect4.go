package builders

import (
	"embed"
	"strings"
)

//go:embed templates/**/*.gotmpl
var templates embed.FS

type Effect4 struct {
}

func newEffect4() *Effect4 {
	return &Effect4{}
}

// Build implements builders.Builder.
func (e *Effect4) Build() ([]File, error) {
	panic("unimplemented")
}

// postgresqlTypeToEffectSchema maps PostgreSQL type names to Effect Schema expressions (TypeScript-side),
// returned as strings (e.g. "Schema.String", "Schema.BigInt").
//
// Notes:
// - bigint/int8 is mapped to Schema.BigInt (safer than Schema.Number).
// - numeric/money are mapped to Schema.String (common driver behavior to preserve precision).
// - json/jsonb are mapped to Schema.Unknown.
// - timestamp/timestamptz/date are mapped to Schema.Date (if you prefer strings, change accordingly).
func (e *Effect4) postgresqlTypeToEffectSchema(dbType string) string {
	switch strings.ToLower(dbType) {
	// serials
	case "serial", "serial4", "pg_catalog.serial":
		return "Schema.Int"
	case "bigserial", "serial8", "pg_catalog.serial8":
		return "Schema.BigInt"
	case "smallserial", "serial2", "pg_catalog.serial2":
		return "Schema.Int"

	// ints
	case "integer", "int", "int4", "pg_catalog.int4":
		return "Schema.Int"
	case "bigint", "int8", "pg_catalog.int8":
		return "Schema.BigInt"
	case "smallint", "int2", "pg_catalog.int2":
		return "Schema.Int"

	// floats
	case "float", "double precision", "float8", "pg_catalog.float8":
		return "Schema.Number"
	case "real", "float4", "pg_catalog.float4":
		return "Schema.Number"

	// numeric / money
	case "numeric", "pg_catalog.numeric":
		return "Schema.String"
	case "money":
		return "Schema.String"

	// boolean
	case "boolean", "bool", "pg_catalog.bool":
		return "Schema.Boolean"

	// json
	case "json", "jsonb":
		return "Schema.Unknown"

	// bytes (depends on your JS runtime/driver)
	case "bytea", "blob", "pg_catalog.bytea":
		return "Schema.Uint8Array"

	// dates/times
	case "date":
		return "Schema.Date"
	case "pg_catalog.time", "pg_catalog.timetz":
		return "Schema.String"
	case "pg_catalog.timestamp", "pg_catalog.timestamptz", "timestamptz":
		return "Schema.Date"
	case "interval", "pg_catalog.interval":
		return "Schema.String"

	// strings
	case "text", "pg_catalog.varchar", "pg_catalog.bpchar", "string", "citext":
		return "Schema.String"

	// uuid
	case "uuid":
		// If you want a check: `Schema.String.check(Schema.isUUID())`
		return "Schema.String"

	// network-ish
	case "inet", "cidr":
		return "Schema.String"
	case "macaddr", "macaddr8":
		return "Schema.String"

	// ltree family
	case "ltree", "lquery", "ltxtquery":
		return "Schema.String"

	default:
		return "Schema.Unknown"
	}
}

var _ Builder = (*Effect4)(nil)
