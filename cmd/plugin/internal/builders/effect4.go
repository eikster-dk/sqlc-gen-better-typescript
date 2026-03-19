package builders

import (
	"embed"
	"fmt"
	"strings"

	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/config"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/models"
)

//go:embed templates/**/*.gotmpl
var templates embed.FS

type Effect4 struct {
	cfg config.Config
}

// NewEffect4 creates a new Effect4 builder
func NewEffect4(cfg config.Config) *Effect4 {
	return &Effect4{cfg: cfg}
}

// Build implements builders.Builder
func (e *Effect4) Build(catalog *models.Catalog, queries []models.Query, log *logger.Logger) ([]File, error) {
	log.Info("Starting Effect4 code generation", logger.F("builder", "effect-v4-unstable"))
	log.Debug("Catalog info", logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)))

	var files []File

	for i, query := range queries {
		log.Info("Generating query",
			logger.F("index", i),
			logger.F("name", query.Name),
			logger.F("cmd", query.Command),
			logger.F("params", len(query.Params)),
			logger.F("results", len(query.Results)))

		// Log type mapping for complex types
		for _, param := range query.Params {
			if param.Type.IsEnum || param.Type.IsArray {
				log.Debug("Param type mapping",
					logger.F("param", param.Name),
					logger.F("sql_type", param.Type.Name),
					logger.F("is_array", param.Type.IsArray),
					logger.F("is_enum", param.Type.IsEnum))
			}
		}

		for _, result := range query.Results {
			if result.Type.IsEnum || result.Type.IsArray {
				log.Debug("Result type mapping",
					logger.F("field", result.Name),
					logger.F("sql_type", result.Type.Name),
					logger.F("is_array", result.Type.IsArray),
					logger.F("is_enum", result.Type.IsEnum))
			}
		}

		// Generate code for each query
		content := e.generateQueryCode(query, catalog, log)

		file := File{
			Name:    fmt.Sprintf("%s.ts", query.Name),
			Content: []byte(content),
		}
		files = append(files, file)

		log.Info("Generated file",
			logger.F("name", file.Name),
			logger.F("size", len(file.Content)))
	}

	log.Info("Effect4 code generation complete", logger.F("files", len(files)))

	return files, nil
}

func (e *Effect4) generateQueryCode(query models.Query, catalog *models.Catalog, log *logger.Logger) string {
	// This is a placeholder implementation
	// In the real implementation, this would:
	// 1. Generate Effect Schema types for params
	// 2. Generate Effect Schema types for results
	// 3. Generate the query function with proper Effect types

	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("// Query: %s\n", query.Name))
	builder.WriteString(fmt.Sprintf("// Command: %s\n", query.Command))
	builder.WriteString(fmt.Sprintf("// SQL: %s\n\n", query.SQL))

	// Generate param schemas
	if len(query.Params) > 0 {
		builder.WriteString("// Parameters:\n")
		for _, param := range query.Params {
			schema := e.sqlTypeToEffectSchema(param.Type, log)
			builder.WriteString(fmt.Sprintf("//   %s: %s (position %d)\n",
				param.Name, schema, param.Position))
		}
		builder.WriteString("\n")
	}

	// Generate result schemas
	if len(query.Results) > 0 {
		builder.WriteString("// Results:\n")
		for _, result := range query.Results {
			schema := e.sqlTypeToEffectSchema(result.Type, log)
			builder.WriteString(fmt.Sprintf("//   %s: %s\n",
				result.Name, schema))
		}
	}

	return builder.String()
}

// sqlTypeToEffectSchema converts internal SqlType to Effect Schema expression
func (e *Effect4) sqlTypeToEffectSchema(t models.SqlType, log *logger.Logger) string {
	var baseSchema string

	switch strings.ToLower(t.Name) {
	// serials
	case "serial", "serial4":
		baseSchema = "Schema.Int"
	case "bigserial", "serial8":
		baseSchema = "Schema.BigInt"
	case "smallserial", "serial2":
		baseSchema = "Schema.Int"

	// ints
	case "integer", "int", "int4":
		baseSchema = "Schema.Int"
	case "bigint", "int8":
		baseSchema = "Schema.BigInt"
	case "smallint", "int2":
		baseSchema = "Schema.Int"

	// floats
	case "float", "double precision", "float8":
		baseSchema = "Schema.Number"
	case "real", "float4":
		baseSchema = "Schema.Number"

	// numeric / money
	case "numeric":
		baseSchema = "Schema.String"
	case "money":
		baseSchema = "Schema.String"

	// boolean
	case "boolean", "bool":
		baseSchema = "Schema.Boolean"

	// json
	case "json", "jsonb":
		baseSchema = "Schema.Unknown"

	// bytes
	case "bytea", "blob":
		baseSchema = "Schema.Uint8Array"

	// dates/times
	case "date":
		baseSchema = "Schema.Date"
	case "time", "timetz":
		baseSchema = "Schema.String"
	case "timestamp", "timestamptz":
		baseSchema = "Schema.Date"
	case "interval":
		baseSchema = "Schema.String"

	// strings
	case "text", "varchar", "bpchar", "string", "citext":
		baseSchema = "Schema.String"

	// uuid
	case "uuid":
		baseSchema = "Schema.String"

	// network-ish
	case "inet", "cidr", "macaddr", "macaddr8":
		baseSchema = "Schema.String"

	// ltree family
	case "ltree", "lquery", "ltxtquery":
		baseSchema = "Schema.String"

	// Enums
	default:
		if t.IsEnum {
			baseSchema = fmt.Sprintf("%sSchema", t.Name) // Reference to generated enum schema
			log.Debug("Using enum schema", logger.F("enum", t.Name))
		} else {
			log.Warn("Unknown SQL type, using Schema.Unknown", logger.F("type", t.Name))
			baseSchema = "Schema.Unknown"
		}
	}

	// Handle arrays
	if t.IsArray {
		baseSchema = fmt.Sprintf("Schema.Array(%s)", baseSchema)
	}

	// Handle nullability
	if t.IsNullable {
		baseSchema = fmt.Sprintf("Schema.optional(%s)", baseSchema)
	}

	return baseSchema
}

var _ Builder = (*Effect4)(nil)
