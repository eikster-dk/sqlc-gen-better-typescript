package debug

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/builders"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-effect/cmd/plugin/internal/models"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// GenerateArtifacts creates debug files as builders.File objects
// These are returned through the plugin protocol instead of written to disk
func GenerateArtifacts(catalog *models.Catalog, queries []models.Query, req *plugin.GenerateRequest, log *logger.Logger, dir string) []builders.File {
	if !log.IsEnabled() {
		return nil
	}

	if dir == "" {
		dir = "debug"
	}

	var files []builders.File

	// Generate logs.txt
	if logsFile := generateLogsFile(log, dir); logsFile != nil {
		files = append(files, *logsFile)
	}

	// Generate ast.json
	if astFile := generateASTFile(catalog, queries, dir); astFile != nil {
		files = append(files, *astFile)
	}

	// Generate request.json
	if reqFile := generateRequestFile(req, dir); reqFile != nil {
		files = append(files, *reqFile)
	}

	log.Info("Debug artifacts generated", logger.F("count", len(files)), logger.F("dir", dir))

	return files
}

func generateLogsFile(log *logger.Logger, dir string) *builders.File {
	entries := log.GetEntries()
	if len(entries) == 0 {
		return nil
	}

	var sb strings.Builder
	for _, entry := range entries {
		timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf("%s [%s] %s", timestamp, entry.Level, entry.Message))

		if len(entry.Fields) > 0 {
			var fieldStrs []string
			for k, v := range entry.Fields {
				fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%v", k, v))
			}
			sb.WriteString(" - " + strings.Join(fieldStrs, ", "))
		}

		sb.WriteString("\n")
	}

	return &builders.File{
		Name:    dir + "/logs.txt",
		Content: []byte(sb.String()),
	}
}

func generateASTFile(catalog *models.Catalog, queries []models.Query, dir string) *builders.File {
	ast := struct {
		GeneratedAt string          `json:"generated_at"`
		Catalog     *models.Catalog `json:"catalog"`
		Queries     []models.Query  `json:"queries"`
	}{
		GeneratedAt: time.Now().Format(time.RFC3339),
		Catalog:     catalog,
		Queries:     queries,
	}

	data, err := json.MarshalIndent(ast, "", "  ")
	if err != nil {
		return nil
	}

	return &builders.File{
		Name:    dir + "/ast.json",
		Content: data,
	}
}

func generateRequestFile(req *plugin.GenerateRequest, dir string) *builders.File {
	// Create a simplified version of the request for debugging
	// We don't include all protobuf fields to keep it readable
	simplified := struct {
		GeneratedAt   string           `json:"generated_at"`
		SqlcVersion   string           `json:"sqlc_version"`
		Settings      *plugin.Settings `json:"settings"`
		Catalog       *plugin.Catalog  `json:"catalog"`
		Queries       []*plugin.Query  `json:"queries"`
		PluginOptions string           `json:"plugin_options"`
		GlobalOptions string           `json:"global_options"`
	}{
		GeneratedAt:   time.Now().Format(time.RFC3339),
		SqlcVersion:   req.SqlcVersion,
		Settings:      req.Settings,
		Catalog:       req.Catalog,
		Queries:       req.Queries,
		PluginOptions: string(req.PluginOptions),
		GlobalOptions: string(req.GlobalOptions),
	}

	data, err := json.MarshalIndent(simplified, "", "  ")
	if err != nil {
		return nil
	}

	return &builders.File{
		Name:    dir + "/request.json",
		Content: data,
	}
}

// LoadAST loads an AST from a JSON file (useful for testing)
func LoadAST(data []byte) (*models.Catalog, []models.Query, error) {
	var ast struct {
		Catalog *models.Catalog `json:"catalog"`
		Queries []models.Query  `json:"queries"`
	}

	if err := json.Unmarshal(data, &ast); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal AST: %w", err)
	}

	return ast.Catalog, ast.Queries, nil
}
