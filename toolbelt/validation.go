package toolbelt

import (
	"fmt"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

// RequireEngine validates that the sqlc request uses one of the supported engines.
func RequireEngine(req *plugin.GenerateRequest, engines ...string) error {
	engine := req.GetSettings().GetEngine()
	for _, supported := range engines {
		if engine == supported {
			return nil
		}
	}
	return fmt.Errorf("Option: engine value is %q but this plugin currently only supports %q", engine, strings.Join(engines, ", "))
}

// RequireOneOf validates that a named value is one of the allowed values.
func RequireOneOf(name, value string, allowed ...string) error {
	for _, candidate := range allowed {
		if value == candidate {
			return nil
		}
	}
	return fmt.Errorf("Option: %s value is %s but can only be one of %q", name, value, allowed)
}

// RequireImportExtension validates relative TypeScript import extension options.
func RequireImportExtension(value string) error {
	return RequireOneOf("import_extension", value, "", ".js", ".ts")
}
