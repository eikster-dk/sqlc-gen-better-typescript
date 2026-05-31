package models

// SqlType represents a generic SQL type independent of any builder.
type SqlType struct {
	Name       string
	Schema     string
	IsNullable bool
	IsArray    bool
	IsEnum     bool
}

// Enum represents a user-defined enum type.
type Enum struct {
	Name   string
	Values []EnumValue
}

// EnumValue represents a single enum constant.
type EnumValue struct {
	Name  string
	Value string
}

// Table represents a database table.
type Table struct {
	Name    string
	Columns []Column
}

// Column represents a table column.
type Column struct {
	Name     string
	Type     SqlType
	Nullable bool
}

// CompositeType represents a PostgreSQL composite type.
type CompositeType struct {
	Name    string
	Columns []Column
}

// Catalog holds mapped database schema information.
type Catalog struct {
	Tables         []Table
	Enums          []Enum
	CompositeTypes []CompositeType
}
