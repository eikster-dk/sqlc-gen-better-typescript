package native

import (
	"fmt"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

// QueryView holds template data for a single query.
type QueryView struct {
	Name         string
	NamePascal   string
	NameCamel    string
	Command      string
	HasParams    bool
	HasResults   bool
	ParamFields  []ZodField
	ResultFields []ZodField
	SQL          string
	ParamList    string // comma-separated "params.foo, params.bar"
}

// ZodField holds a single field with its Zod schema expression.
type ZodField struct {
	Name   string
	Schema string
}

// QueriesData is passed to the Queries template.
type QueriesData struct {
	FileStem      string // e.g. "customers" — used for import paths
	ImportExt     string // e.g. ".js"
	QueryViews    []QueryView
	SqlcVersion   string
	PluginVersion string
}

// RequestsData is passed to the Requests template.
type RequestsData struct {
	SqlcVersion   string
	PluginVersion string
	QueryViews    []QueryView
}

// ResponsesData is passed to the Responses template.
type ResponsesData struct {
	SqlcVersion   string
	PluginVersion string
	QueryViews    []QueryView
}

// zodBaseType maps a SqlType to its base Zod expression (no nullable/optional modifier).
func (n *Native) zodBaseType(t models.SqlType) string {
	switch strings.ToLower(t.Name) {
	case "serial", "serial4", "smallserial", "serial2",
		"integer", "int", "int4", "smallint", "int2",
		"float", "double precision", "float8", "real", "float4":
		return "z.number()"
	case "bigserial", "serial8", "bigint", "int8":
		return "z.coerce.bigint()"
	case "text", "varchar", "char", "bpchar", "citext",
		"numeric", "money", "time", "timetz", "interval",
		"inet", "cidr", "macaddr", "macaddr8", "ltree", "lquery", "ltxtquery":
		return "z.string()"
	case "uuid":
		return "z.string().uuid()"
	case "boolean", "bool":
		return "z.boolean()"
	case "json", "jsonb":
		return "z.unknown()"
	case "bytea", "blob":
		return "z.instanceof(Buffer)"
	case "date", "timestamp", "timestamptz":
		return "z.coerce.date()"
	default:
		if t.IsEnum {
			// Will be handled in later phase
			return "z.string()"
		}
		return "z.unknown()"
	}
}

// zodTypeForParam builds the Zod expression for a query input parameter.
// Nullable params become optional.
func (n *Native) zodTypeForParam(t models.SqlType) string {
	base := n.zodBaseType(t)
	if t.IsArray {
		base = fmt.Sprintf("z.array(%s)", n.zodBaseType(models.SqlType{Name: t.Name}))
	}
	if t.IsNullable {
		return base + ".optional()"
	}
	return base
}

// zodTypeForResult builds the Zod expression for a query output column.
// Nullable result columns become nullable.
func (n *Native) zodTypeForResult(t models.SqlType) string {
	base := n.zodBaseType(t)
	if t.IsArray {
		base = fmt.Sprintf("z.array(%s)", n.zodBaseType(models.SqlType{Name: t.Name}))
	}
	if t.IsNullable {
		return base + ".nullable()"
	}
	return base
}
