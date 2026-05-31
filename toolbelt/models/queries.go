package models

// Param represents an input parameter to a SQL query ($1, $2, etc.).
type Param struct {
	Name     string
	Position int
	Type     SqlType
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
