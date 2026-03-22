package models

// Param represents an input parameter to a SQL query ($1, $2, etc.)
type Param struct {
	Name     string // Parameter name (if named) or generated name like "arg1"
	Position int    // 1-based position ($1 = 1)
	Type     SqlType
}

// ResultField represents a column in the query result set
type ResultField struct {
	Name         string // Column name (may be aliased for duplicates: id -> id2)
	OriginalName string // Original column name from SQL (before aliasing)
	Type         SqlType
	Table        string // Source table name (if known)
	IsAliased    bool   // True if this field was auto-aliased due to duplicate
	EmbedTable   string // Table name if this field is part of sqlc.embed(table)
}

// EmbedGroup represents a group of columns from sqlc.embed(table)
type EmbedGroup struct {
	TableName string        // e.g., "orders"
	Fields    []ResultField // Columns belonging to this embed group
}

// Query represents a parsed SQL query with its metadata
type Query struct {
	Name         string        // Function name from annotation (e.g., "GetUser")
	SQL          string        // Original SQL with placeholders
	RewrittenSQL string        // SQL rewritten with explicit column aliases (if needed)
	Command      string        // :one, :many, :exec, :execrows, :execresult, :copyfrom, :batch
	Params       []Param       // Input parameters
	Results      []ResultField // Output columns (empty for :exec)
	Tables       []string      // Tables this query references
	HasEnum      bool          // True if any result field is an enum type
	HasEmbeds    bool          // True if any result field is part of an embed
	EmbedGroups  []EmbedGroup  // Groups of columns from sqlc.embed(table)
	Filename     string        // Source SQL file path (e.g., "queries.sql")
}
