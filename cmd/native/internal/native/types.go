package native

import (
	"fmt"
	"sort"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
)

// QueryView holds template data for a single query.
type QueryView struct {
	Name              string
	NamePascal        string
	NameCamel         string
	Command           string
	HasParams         bool
	HasResults        bool
	HasResultMappings bool
	ParamFields       []ZodField
	ResultFields      []ZodField
	ResultMappings    []ResultMapping
	SQL               string
	SQLComment        string // SQL with each line prefixed by "// "
	ParamList         string // comma-separated "params.foo, params.bar"
	QueryParamList    string // comma-separated query param specs for slice expansion
	HasSlices         bool
	EnumImports       []string
}

// ZodField holds a single field with its Zod schema expression.
type ZodField struct {
	Name       string
	Schema     string
	EnumImport string
}

type ZodExpr struct {
	Schema     string
	ModelNames []string
}

type EnumView struct {
	NamePascal string
	Schema     string
}

type TableRowView struct {
	NamePascal string
	Fields     []ZodField
}

type ResultMapping struct {
	Name       string
	RowField   string
	Object     []ResultMapping
	Schema     string
	IsObject   bool
	IsLast     bool
	Indent     string
	ParentPath string
}

// QueriesData is passed to the Queries template.
type QueriesData struct {
	FileStem        string // e.g. "customers" — used for import paths
	ImportExt       string // e.g. ".js"
	NeedsExecResult bool
	NeedsSlices     bool
	QueryViews      []QueryView
	SqlcVersion     string
	PluginVersion   string
}

type ModelsData struct {
	SqlcVersion   string
	PluginVersion string
	Enums         []EnumView
	TableRows     []TableRowView
}

// RequestsData is passed to the Requests template.
type RequestsData struct {
	SqlcVersion   string
	PluginVersion string
	QueryViews    []QueryView
	EnumImports   []string
	ImportExt     string
}

// ResponsesData is passed to the Responses template.
type ResponsesData struct {
	SqlcVersion   string
	PluginVersion string
	QueryViews    []QueryView
	EnumImports   []string
	ImportExt     string
}

// zodEnumUnion builds a Zod union, literal, or never type from a slice of enum values.
// 0 values → z.never(), 1 value → z.literal("x"), 2+ values → z.union([z.literal("a"), ...]).
func zodEnumUnion(values []string) string {
	switch len(values) {
	case 0:
		return "z.never()"
	case 1:
		return fmt.Sprintf(`z.literal(%q)`, values[0])
	default:
		parts := make([]string, len(values))
		for i, v := range values {
			parts[i] = fmt.Sprintf(`z.literal(%q)`, v)
		}
		return fmt.Sprintf("z.union([%s])", strings.Join(parts, ", "))
	}
}

func uniqueSorted(values []string) []string {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		set[value] = struct{}{}
	}
	result := make([]string, 0, len(set))
	for value := range set {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func appendUniqueSorted(base []string, values ...[]string) []string {
	all := append([]string{}, base...)
	for _, set := range values {
		all = append(all, set...)
	}
	return uniqueSorted(all)
}

// zodBaseType maps a SqlType to its base Zod expression (no nullable/optional modifier).
func (n *Native) zodBaseType(t models.SqlType) ZodExpr {
	switch strings.ToLower(t.Name) {
	case "serial", "serial4", "smallserial", "serial2",
		"integer", "int", "int4", "smallint", "int2",
		"float", "double precision", "float8", "real", "float4":
		return ZodExpr{Schema: "z.number()"}
	case "bigserial", "serial8", "bigint", "int8":
		return ZodExpr{Schema: "z.coerce.bigint()"}
	case "text", "varchar", "char", "bpchar", "citext",
		"numeric", "money", "time", "timetz", "interval",
		"inet", "cidr", "macaddr", "macaddr8", "ltree", "lquery", "ltxtquery":
		return ZodExpr{Schema: "z.string()"}
	case "uuid":
		return ZodExpr{Schema: "z.string().uuid()"}
	case "boolean", "bool":
		return ZodExpr{Schema: "z.boolean()"}
	case "json", "jsonb":
		return ZodExpr{Schema: "z.unknown()"}
	case "bytea", "blob":
		return ZodExpr{Schema: "z.instanceof(Buffer)"}
	case "date", "timestamp", "timestamptz":
		return ZodExpr{Schema: "z.coerce.date()"}
	default:
		if t.IsEnum {
			if _, ok := n.enumValues[t.Name]; ok {
				name := toPascalCase(t.Name)
				return ZodExpr{Schema: name, ModelNames: []string{name}}
			}
			// Fallback when enum is not in the catalog (e.g. unknown type source).
			return ZodExpr{Schema: "z.string()"}
		}
		return ZodExpr{Schema: "z.unknown()"}
	}
}

// scalarType returns a copy of t with IsArray and IsNullable cleared,
// preserving IsEnum and Schema so zodBaseType can dispatch correctly.
func scalarType(t models.SqlType) models.SqlType {
	return models.SqlType{
		Name:   t.Name,
		Schema: t.Schema,
		IsEnum: t.IsEnum,
	}
}

// zodTypeForParam builds the Zod expression for a query input parameter.
// Nullable params become optional.
func (n *Native) zodTypeForParam(t models.SqlType) string {
	base := n.zodBaseType(scalarType(t)).Schema
	if t.IsArray {
		base = fmt.Sprintf("z.array(%s)", base)
	}
	if t.IsNullable {
		return base + ".optional()"
	}
	return base
}

// zodTypeForResult builds the Zod expression for a query output column.
// Nullable result columns become nullable.
func (n *Native) zodTypeForResult(t models.SqlType) string {
	base := n.zodBaseType(scalarType(t)).Schema
	if t.IsArray {
		base = fmt.Sprintf("z.array(%s)", base)
	}
	if t.IsNullable {
		return base + ".nullable()"
	}
	return base
}
