package toolbelt

import (
	"context"
	"testing"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type testBuilder struct{}

func (testBuilder) Build(ctx BuildContext) ([]File, error) {
	if ctx.Catalog == nil {
		return nil, nil
	}
	return []File{{Name: "out.ts", Content: []byte("ok")}}, nil
}

func TestGenerate(t *testing.T) {
	validated := false
	resp, err := Generate(context.Background(), &plugin.GenerateRequest{
		Settings: &plugin.Settings{Engine: "postgresql"},
		Catalog:  &plugin.Catalog{},
	}, Options[struct{}]{
		ParseConfig: func(req *plugin.GenerateRequest) (struct{}, error) { return struct{}{}, nil },
		ValidateConfig: func(cfg struct{}, req *plugin.GenerateRequest) error {
			validated = true
			return nil
		},
		NewBuilder: func(cfg struct{}) (Builder, error) { return testBuilder{}, nil },
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if !validated {
		t.Fatalf("expected ValidateConfig to be called")
	}
	if len(resp.Files) != 1 || resp.Files[0].Name != "out.ts" || string(resp.Files[0].Contents) != "ok" {
		t.Fatalf("unexpected response files: %#v", resp.Files)
	}
}
