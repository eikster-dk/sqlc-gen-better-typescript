package mapper

import (
	"fmt"
	"strings"

	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/models"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// Mapper converts sqlc plugin types to internal models
type Mapper struct {
	catalog *models.Catalog
	enumSet map[string]bool // Set of enum type names for quick lookup
	logger  *logger.Logger
}

// New creates a new Mapper from the generate request
func New(req *plugin.GenerateRequest, log *logger.Logger) *Mapper {
	catalog := mapCatalog(req.Catalog)

	// Build set of enum names for quick lookup
	enumSet := make(map[string]bool)
	for _, e := range catalog.Enums {
		enumSet[e.Name] = true
	}

	m := &Mapper{
		catalog: catalog,
		enumSet: enumSet,
		logger:  log,
	}

	log.Debug("Mapped catalog",
		logger.F("tables", len(catalog.Tables)),
		logger.F("enums", len(catalog.Enums)),
		logger.F("composite_types", len(catalog.CompositeTypes)))

	for _, table := range catalog.Tables {
		log.Debug("Mapped table",
			logger.F("name", table.Name),
			logger.F("columns", len(table.Columns)))
	}

	for _, enum := range catalog.Enums {
		log.Debug("Mapped enum",
			logger.F("name", enum.Name),
			logger.F("values", len(enum.Values)))
	}

	return m
}

// Catalog returns the mapped catalog
func (m *Mapper) Catalog() *models.Catalog {
	return m.catalog
}

// MapQueries converts sqlc queries to internal model queries
func (m *Mapper) MapQueries(req *plugin.GenerateRequest) []models.Query {
	var queries []models.Query

	for _, q := range req.Queries {
		query := m.mapQuery(q)
		queries = append(queries, query)
	}

	return queries
}

func (m *Mapper) mapQuery(q *plugin.Query) models.Query {
	params := m.mapParams(q.Params)
	results := m.mapResults(q.Columns)

	// Rewrite SQL if needed (adds explicit column aliases for duplicates)
	rewrittenSQL := models.RewriteSQLWithAliases(q.Text, results)
	if rewrittenSQL != q.Text {
		m.logger.Debug("Rewrote SQL with aliases",
			logger.F("query", q.Name),
			logger.F("original", q.Text),
			logger.F("rewritten", rewrittenSQL))
	}

	return models.Query{
		Name:         q.Name,
		SQL:          q.Text,
		RewrittenSQL: rewrittenSQL,
		Command:      q.Cmd,
		Params:       params,
		Results:      results,
		Tables:       extractTables(q.Columns),
		HasEnum:      hasEnumInResults(results, m.enumSet),
		Filename:     q.Filename,
	}
}

func (m *Mapper) mapParams(params []*plugin.Parameter) []models.Param {
	var result []models.Param
	nameCount := make(map[string]int) // Track parameter name occurrences for deduplication

	for i, p := range params {
		name := ""
		if p.Column != nil {
			name = p.Column.Name
		}
		if name == "" {
			name = fmt.Sprintf("arg%d", i+1)
		}

		// Handle duplicate parameter names by suffixing with _2, _3, etc.
		nameCount[name]++
		if nameCount[name] > 1 {
			name = fmt.Sprintf("%s_%d", name, nameCount[name])
		}

		position := int(p.Number)
		if position == 0 {
			position = i + 1
		}

		sqlType := models.SqlType{}
		if p.Column != nil && p.Column.Type != nil {
			sqlType = m.mapSqlTypeFromIdentifier(p.Column.Type, p.Column.NotNull, p.Column.IsArray)
		}

		result = append(result, models.Param{
			Name:     name,
			Position: position,
			Type:     sqlType,
		})
	}

	return result
}

func (m *Mapper) mapResults(columns []*plugin.Column) []models.ResultField {
	var result []models.ResultField
	fieldCount := make(map[string]int) // Track field name occurrences

	for _, col := range columns {
		tableName := ""
		if col.Table != nil {
			tableName = col.Table.Name
		}

		// Get original column name
		originalName := col.Name

		// Check for duplicates and create alias in format table_column
		fieldCount[originalName]++
		uniqueName := originalName
		isAliased := false

		if count := fieldCount[originalName]; count > 1 {
			// Create alias: table_column (e.g., sms_episodes_id)
			uniqueName = fmt.Sprintf("%s_%s", tableName, originalName)
			isAliased = true
			m.logger.Debug("Auto-aliased duplicate column",
				logger.F("original", originalName),
				logger.F("alias", uniqueName),
				logger.F("table", tableName))
		}

		result = append(result, models.ResultField{
			Name:         uniqueName,
			OriginalName: originalName,
			Type:         m.mapSqlTypeFromColumn(col),
			Table:        tableName,
			IsAliased:    isAliased,
		})
	}

	return result
}

func mapCatalog(c *plugin.Catalog) *models.Catalog {
	catalog := &models.Catalog{}

	for _, schema := range c.Schemas {
		for _, table := range schema.Tables {
			catalog.Tables = append(catalog.Tables, mapTable(table))
		}

		for _, enum := range schema.Enums {
			catalog.Enums = append(catalog.Enums, mapEnum(enum))
		}

		// Composite types - prepared but not fully implemented
		for _, ct := range schema.CompositeTypes {
			catalog.CompositeTypes = append(catalog.CompositeTypes, mapCompositeType(ct))
		}
	}

	return catalog
}

func mapTable(t *plugin.Table) models.Table {
	var columns []models.Column

	for _, col := range t.Columns {
		columns = append(columns, models.Column{
			Name:     col.Name,
			Type:     mapSqlTypeFromColumnStatic(col),
			Nullable: !col.NotNull,
		})
	}

	tableName := ""
	if t.Rel != nil {
		tableName = t.Rel.Name
	}

	return models.Table{
		Name:    tableName,
		Columns: columns,
	}
}

func mapEnum(e *plugin.Enum) models.Enum {
	var values []models.EnumValue

	for _, v := range e.Vals {
		values = append(values, models.EnumValue{
			Name:  v,
			Value: v,
		})
	}

	return models.Enum{
		Name:   e.Name,
		Values: values,
	}
}

func mapCompositeType(ct *plugin.CompositeType) models.CompositeType {
	// Prepared for future use - not fully implemented
	// Composite types in sqlc don't expose their attributes currently
	return models.CompositeType{
		Name:    ct.Name,
		Columns: nil,
	}
}

func (m *Mapper) mapSqlTypeFromColumn(col *plugin.Column) models.SqlType {
	if col.Type == nil {
		return models.SqlType{
			Name:       "unknown",
			IsNullable: !col.NotNull,
			IsArray:    col.IsArray,
		}
	}

	return m.mapSqlTypeFromIdentifier(col.Type, col.NotNull, col.IsArray)
}

func (m *Mapper) mapSqlTypeFromIdentifier(id *plugin.Identifier, notNull, isArray bool) models.SqlType {
	typeName := id.GetName()
	schema := id.GetSchema()

	normalized := normalizeTypeName(typeName)

	sqlType := models.SqlType{
		Name:       normalized,
		Schema:     schema,
		IsNullable: !notNull,
		IsArray:    isArray || strings.HasSuffix(typeName, "[]"),
		IsEnum:     m.enumSet[normalized],
	}

	if sqlType.IsEnum {
		m.logger.Debug("Detected enum type",
			logger.F("type", normalized),
			logger.F("schema", schema))
	}

	if sqlType.IsArray {
		m.logger.Debug("Detected array type",
			logger.F("type", normalized),
			logger.F("is_nullable", sqlType.IsNullable))
	}

	return sqlType
}

func mapSqlTypeFromColumnStatic(col *plugin.Column) models.SqlType {
	if col.Type == nil {
		return models.SqlType{
			Name:       "unknown",
			IsNullable: !col.NotNull,
			IsArray:    col.IsArray,
		}
	}

	typeName := col.Type.GetName()
	schema := col.Type.GetSchema()
	normalized := normalizeTypeName(typeName)

	return models.SqlType{
		Name:       normalized,
		Schema:     schema,
		IsNullable: !col.NotNull,
		IsArray:    col.IsArray || strings.HasSuffix(typeName, "[]"),
		IsEnum:     false, // Will be set later if needed
	}
}

func normalizeTypeName(name string) string {
	// Remove array suffix for normalization
	name = strings.TrimSuffix(name, "[]")

	// Strip schema prefix (e.g., "pg_catalog.int4" -> "int4")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}

	// Normalize common aliases
	switch strings.ToLower(name) {
	case "int4", "integer":
		return "integer"
	case "int8", "bigint":
		return "bigint"
	case "int2", "smallint":
		return "smallint"
	case "varchar", "character varying":
		return "varchar"
	case "char", "character":
		return "char"
	default:
		return name
	}
}

func extractTables(columns []*plugin.Column) []string {
	tableMap := make(map[string]bool)

	for _, col := range columns {
		if col.Table != nil && col.Table.Name != "" {
			tableMap[col.Table.Name] = true
		}
	}

	tables := make([]string, 0, len(tableMap))
	for table := range tableMap {
		tables = append(tables, table)
	}

	return tables
}

func hasEnumInResults(results []models.ResultField, enumSet map[string]bool) bool {
	for _, r := range results {
		if enumSet[r.Type.Name] {
			return true
		}
	}

	return false
}
