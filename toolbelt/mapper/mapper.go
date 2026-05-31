package mapper

import (
	"fmt"
	"strings"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/models"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// Mapper converts sqlc plugin types to the public toolbelt IR.
type Mapper struct {
	catalog *models.Catalog
	enumSet map[string]bool
	logger  *logger.Logger
}

// New creates a new Mapper from the generate request.
func New(req *plugin.GenerateRequest, log *logger.Logger) *Mapper {
	catalog := mapCatalog(req.Catalog)
	enumSet := make(map[string]bool)
	for _, e := range catalog.Enums {
		enumSet[e.Name] = true
	}

	for i := range catalog.Tables {
		for j := range catalog.Tables[i].Columns {
			col := &catalog.Tables[i].Columns[j]
			if enumSet[col.Type.Name] {
				col.Type.IsEnum = true
			}
		}
	}

	m := &Mapper{catalog: catalog, enumSet: enumSet, logger: log}
	log.Debug("Mapped catalog", logger.F("tables", len(catalog.Tables)), logger.F("enums", len(catalog.Enums)), logger.F("composite_types", len(catalog.CompositeTypes)))
	return m
}

// Catalog returns the mapped catalog.
func (m *Mapper) Catalog() *models.Catalog {
	return m.catalog
}

// MapQueries converts sqlc queries to toolbelt IR queries.
func (m *Mapper) MapQueries(req *plugin.GenerateRequest) []models.Query {
	queries := make([]models.Query, 0, len(req.Queries))
	for _, q := range req.Queries {
		queries = append(queries, m.mapQuery(q))
	}
	return queries
}

func (m *Mapper) mapQuery(q *plugin.Query) models.Query {
	params := m.mapParams(q.Params)
	results, embedGroups := m.mapResults(q.Columns)
	rewrittenSQL := models.RewriteSQLWithAliases(q.Text, results)
	if rewrittenSQL != q.Text {
		m.logger.Debug("Rewrote SQL with aliases", logger.F("query", q.Name), logger.F("original", q.Text), logger.F("rewritten", rewrittenSQL))
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
		HasEmbeds:    len(embedGroups) > 0,
		EmbedGroups:  embedGroups,
		Filename:     q.Filename,
	}
}

func (m *Mapper) mapParams(params []*plugin.Parameter) []models.Param {
	result := make([]models.Param, 0, len(params))
	nameCount := make(map[string]int)
	for i, p := range params {
		name := ""
		if p.Column != nil {
			name = p.Column.Name
		}
		if name == "" {
			name = fmt.Sprintf("arg%d", i+1)
		}

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

		result = append(result, models.Param{Name: name, Position: position, Type: sqlType})
	}
	return result
}

func (m *Mapper) mapResults(columns []*plugin.Column) ([]models.ResultField, []models.EmbedGroup) {
	var result []models.ResultField
	fieldCount := make(map[string]int)
	embedMap := make(map[string]*models.EmbedGroup)
	var embedOrder []string

	for _, col := range columns {
		tableName := ""
		if col.Table != nil {
			tableName = col.Table.Name
		}

		embedTableName := ""
		if embedTable := col.GetEmbedTable(); embedTable != nil {
			embedTableName = embedTable.GetName()
			if _, exists := embedMap[embedTableName]; !exists {
				embedMap[embedTableName] = &models.EmbedGroup{TableName: embedTableName}
				embedOrder = append(embedOrder, embedTableName)
			}

			embedFields := m.expandEmbedColumns(embedTableName)
			for _, field := range embedFields {
				field.Name = fmt.Sprintf("%s_%s", embedTableName, field.OriginalName)
				field.IsAliased = true
				field.EmbedTable = embedTableName
				result = append(result, field)
				embedMap[embedTableName].Fields = append(embedMap[embedTableName].Fields, field)
			}
			continue
		}

		originalName := col.Name
		fieldCount[originalName]++
		uniqueName := originalName
		isAliased := false
		if count := fieldCount[originalName]; count > 1 {
			uniqueName = fmt.Sprintf("%s_%s", tableName, originalName)
			isAliased = true
			m.logger.Debug("Auto-aliased duplicate column", logger.F("original", originalName), logger.F("alias", uniqueName), logger.F("table", tableName))
		}

		result = append(result, models.ResultField{
			Name:         uniqueName,
			OriginalName: originalName,
			Type:         m.mapSqlTypeFromColumn(col),
			Table:        tableName,
			IsAliased:    isAliased,
			EmbedTable:   embedTableName,
		})
	}

	embedGroups := make([]models.EmbedGroup, 0, len(embedOrder))
	for _, tableName := range embedOrder {
		embedGroups = append(embedGroups, *embedMap[tableName])
	}
	return result, embedGroups
}

func (m *Mapper) expandEmbedColumns(tableName string) []models.ResultField {
	var fields []models.ResultField
	for _, table := range m.catalog.Tables {
		if table.Name != tableName {
			continue
		}
		for _, col := range table.Columns {
			fields = append(fields, models.ResultField{Name: col.Name, OriginalName: col.Name, Type: col.Type, Table: tableName})
		}
		break
	}
	if len(fields) == 0 {
		m.logger.Warn("Could not find embed table in catalog", logger.F("table", tableName))
	}
	return fields
}

func mapCatalog(c *plugin.Catalog) *models.Catalog {
	catalog := &models.Catalog{}
	if c == nil {
		return catalog
	}
	for _, schema := range c.Schemas {
		for _, table := range schema.Tables {
			catalog.Tables = append(catalog.Tables, mapTable(table))
		}
		for _, enum := range schema.Enums {
			catalog.Enums = append(catalog.Enums, mapEnum(enum))
		}
		for _, ct := range schema.CompositeTypes {
			catalog.CompositeTypes = append(catalog.CompositeTypes, mapCompositeType(ct))
		}
	}
	return catalog
}

func mapTable(t *plugin.Table) models.Table {
	columns := make([]models.Column, 0, len(t.Columns))
	for _, col := range t.Columns {
		columns = append(columns, models.Column{Name: col.Name, Type: mapSqlTypeFromColumnStatic(col), Nullable: !col.NotNull})
	}

	tableName := ""
	if t.Rel != nil {
		tableName = t.Rel.Name
	}
	return models.Table{Name: tableName, Columns: columns}
}

func mapEnum(e *plugin.Enum) models.Enum {
	values := make([]models.EnumValue, 0, len(e.Vals))
	for _, v := range e.Vals {
		values = append(values, models.EnumValue{Name: v, Value: v})
	}
	return models.Enum{Name: e.Name, Values: values}
}

func mapCompositeType(ct *plugin.CompositeType) models.CompositeType {
	return models.CompositeType{Name: ct.Name, Columns: nil}
}

func (m *Mapper) mapSqlTypeFromColumn(col *plugin.Column) models.SqlType {
	if col.Type == nil {
		return models.SqlType{Name: "unknown", IsNullable: !col.NotNull, IsArray: col.IsArray}
	}
	return m.mapSqlTypeFromIdentifier(col.Type, col.NotNull, col.IsArray)
}

func (m *Mapper) mapSqlTypeFromIdentifier(id *plugin.Identifier, notNull, isArray bool) models.SqlType {
	typeName := id.GetName()
	normalized := normalizeTypeName(typeName)
	sqlType := models.SqlType{
		Name:       normalized,
		Schema:     id.GetSchema(),
		IsNullable: !notNull,
		IsArray:    isArray || strings.HasSuffix(typeName, "[]"),
		IsEnum:     m.enumSet[normalized],
	}
	if sqlType.IsEnum {
		m.logger.Debug("Detected enum type", logger.F("type", normalized), logger.F("schema", sqlType.Schema))
	}
	return sqlType
}

func mapSqlTypeFromColumnStatic(col *plugin.Column) models.SqlType {
	if col.Type == nil {
		return models.SqlType{Name: "unknown", IsNullable: !col.NotNull, IsArray: col.IsArray}
	}
	typeName := col.Type.GetName()
	return models.SqlType{
		Name:       normalizeTypeName(typeName),
		Schema:     col.Type.GetSchema(),
		IsNullable: !col.NotNull,
		IsArray:    col.IsArray || strings.HasSuffix(typeName, "[]"),
		IsEnum:     false,
	}
}

func normalizeTypeName(name string) string {
	name = strings.TrimSuffix(name, "[]")
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
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
