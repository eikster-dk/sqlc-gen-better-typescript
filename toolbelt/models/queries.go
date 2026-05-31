package models

// Param represents an input parameter to a SQL query ($1, $2, etc.).
type Param struct {
	Name     string
	Position int
	Type     SqlType
	Named    bool
	Slice    bool
}

// ResultField represents a column in the query result set.
type ResultField struct {
	Name         string
	OriginalName string
	Type         SqlType
	Table        string
	IsAliased    bool
	EmbedTable   string
}

// EmbedGroup represents a group of columns from sqlc.embed(table).
type EmbedGroup struct {
	TableName string
	Fields    []ResultField
}

// Query represents a parsed SQL query with normalized sqlc metadata.
type Query struct {
	Name         string
	SQL          string
	RewrittenSQL string
	Command      string
	Params       []Param
	Results      []ResultField
	Tables       []string
	HasEnum      bool
	HasEmbeds    bool
	EmbedGroups  []EmbedGroup
	Filename     string
}

// QueryPlan is the builder-facing query contract derived from Query.
// It separates SQL execution, request fields, database row shape, and public result shape.
type QueryPlan struct {
	Name     string
	Command  string
	Filename string
	Source   QuerySourcePlan
	Request  RequestPlan
	Response ResponsePlan
	Features QueryFeatures
}

// QueryFeatures summarizes advanced sqlc behaviors normalized by toolbelt.
type QueryFeatures struct {
	UsesNamedArgs    bool
	UsesNullableArgs bool
	UsesSlices       bool
	UsesEmbeds       bool
	RewritesSQL      bool
}

// QuerySourcePlan contains SQL variants and placeholder occurrences for execution.
type QuerySourcePlan struct {
	OriginalSQL  string
	CanonicalSQL string
	ExecSQL      string
	Parameters   []SQLParameter
}

// SQLParameter is a placeholder occurrence in executable SQL.
type SQLParameter struct {
	Position    int
	FieldName   string
	Placeholder string
	Slice       bool
}

// RequestPlan describes the logical input object a builder should expose.
type RequestPlan struct {
	Fields []RequestField
}

// RequestField is a logical input field. One field can be referenced by multiple SQL parameters.
type RequestField struct {
	Name     string
	Param    Param
	Type     SqlType
	Optional bool
	Nullable bool
	Slice    bool
	Named    bool
}

// ResponsePlan separates the raw database row shape from the public generated result shape.
type ResponsePlan struct {
	Row    RowPlan
	Result ResultPlan
}

// RowPlan describes the database row returned by the execution SQL.
type RowPlan struct {
	Fields []RowField
}

// RowField is a field in the database row shape.
type RowField struct {
	Name         string
	OriginalName string
	Type         SqlType
	Table        string
	EmbedTable   string
	Aliased      bool
}

// ResultPlan describes the public result shape exposed by generated code.
type ResultPlan struct {
	Shape ResultShape
}

// ResultShape is a tree representing the public result object.
type ResultShape struct {
	Fields []ResultShapeField
}

type ResultShapeFieldKind string

const (
	ResultShapeFieldValue  ResultShapeFieldKind = "value"
	ResultShapeFieldObject ResultShapeFieldKind = "object"
)

// ResultShapeField is either a leaf sourced from a row field or a nested object.
type ResultShapeField struct {
	Kind   ResultShapeFieldKind
	Name   string
	Source *RowFieldRef
	Object *ResultShape
}

// RowFieldRef points at a field in ResponsePlan.Row.
type RowFieldRef struct {
	Name string
}
