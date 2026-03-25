package effect4

type File struct {
	Name    string
	Content []byte
}

type Imports map[string][]string

type SchemaField struct {
	Name         string
	Schema       string
	ModelImports []string
}

type SchemaExpr struct {
	Schema       string
	ModelImports []string
}

type EnumView struct {
	NamePascal string
	Values     []string
}

type TableRowView struct {
	NamePascal string
	Fields     []SchemaField
}

type FieldMap struct {
	RowFieldName   string
	EmbedFieldName string
}

type EmbedGroupView struct {
	TableName    string
	FieldName    string
	RowSchema    string
	SchemaName   string
	Fields       []SchemaField
	FieldMapping []FieldMap
}

type QueryView struct {
	Name                string
	NamePascal          string
	NameCamel           string
	Command             string
	HasParams           bool
	HasResults          bool
	HasEmbeds           bool
	ReturnType          string
	SqlSchemaMethod     string
	RequestSchema       string
	ParamFields         []SchemaField
	ResultFields        []SchemaField
	EmbedGroups         []EmbedGroupView
	RowFields           []SchemaField
	OriginalSQL         string
	SQLTemplateLiteral  string
	ParamList           string
	UseTemplateLiterals bool
}

type RepositoryData struct {
	RepositoryName      string
	RepositoryNameCamel string
	Filename            string
	Imports             Imports
	QueryViews          []QueryView
	SqlcVersion         string
	PluginVersion       string
}

type RequestData struct {
	RepositoryName string
	Imports        Imports
	QueryViews     []QueryView
	SqlcVersion    string
	PluginVersion  string
}

type ResponseData struct {
	RepositoryName string
	Imports        Imports
	QueryViews     []QueryView
	SqlcVersion    string
	PluginVersion  string
}

type ModelsData struct {
	Imports       Imports
	Enums         []EnumView
	TableRows     []TableRowView
	NeedsBigInt   bool
	NeedsExecRows bool
	SqlcVersion   string
	PluginVersion string
}
