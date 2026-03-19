package models

// SqlType represents a generic SQL type independent of any builder
type SqlType struct {
	Name       string // e.g., "serial", "text", "user_role"
	Schema     string // e.g., "pg_catalog" or "" for user types
	IsNullable bool
	IsArray    bool // True if this is an array type like text[]
	IsEnum     bool // True if this is a user-defined enum
}

// Enum represents a user-defined enum type
type Enum struct {
	Name   string
	Values []EnumValue
}

// EnumValue represents a single enum constant
type EnumValue struct {
	Name  string // Display name (e.g., "Admin")
	Value string // Stored value (e.g., "admin")
}

// Table represents a database table
type Table struct {
	Name    string
	Columns []Column
}

// Column represents a table column
type Column struct {
	Name     string
	Type     SqlType
	Nullable bool
}

// CompositeType represents a PostgreSQL composite type (prepared but not fully implemented)
type CompositeType struct {
	Name    string
	Columns []Column
}

// Catalog holds all database schema information
type Catalog struct {
	Tables         []Table
	Enums          []Enum
	CompositeTypes []CompositeType // Prepared for future use
}
