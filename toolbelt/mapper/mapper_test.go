package mapper

import (
	"testing"

	"github.com/eikster-dk/sqlc-gen-better-typescript/toolbelt/logger"
	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func TestMapQueriesPreservesParameterMacroMetadata(t *testing.T) {
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{},
		Queries: []*plugin.Query{{
			Name: "GetUsers",
			Cmd:  ":many",
			Text: "SELECT * FROM users WHERE id IN (/*SLICE:ids*/?) AND name = ?",
			Params: []*plugin.Parameter{
				{
					Number: 1,
					Column: &plugin.Column{
						Name:         "ids",
						NotNull:      true,
						IsNamedParam: true,
						IsSqlcSlice:  true,
						Type:         &plugin.Identifier{Name: "integer"},
					},
				},
				{
					Number: 2,
					Column: &plugin.Column{
						Name:         "name",
						NotNull:      false,
						IsNamedParam: true,
						Type:         &plugin.Identifier{Name: "text"},
					},
				},
			},
		}},
	}

	m := New(req, logger.New(false))
	queries := m.MapQueries(req)
	if len(queries) != 1 {
		t.Fatalf("expected one query, got %d", len(queries))
	}

	params := queries[0].Params
	if len(params) != 2 {
		t.Fatalf("expected two params, got %#v", params)
	}

	ids := params[0]
	if ids.Name != "ids" || ids.Position != 1 || !ids.Named || !ids.Slice {
		t.Fatalf("slice param metadata not preserved: %#v", ids)
	}
	if ids.Type.Name != "integer" || ids.Type.IsNullable {
		t.Fatalf("unexpected ids type: %#v", ids.Type)
	}

	name := params[1]
	if name.Name != "name" || name.Position != 2 || !name.Named || name.Slice {
		t.Fatalf("named param metadata not preserved: %#v", name)
	}
	if name.Type.Name != "text" || !name.Type.IsNullable {
		t.Fatalf("expected nullable text param: %#v", name.Type)
	}
}

func TestMapQueriesExpandsEmbedColumns(t *testing.T) {
	req := &plugin.GenerateRequest{
		Catalog: &plugin.Catalog{Schemas: []*plugin.Schema{{Tables: []*plugin.Table{
			{
				Rel: &plugin.Identifier{Name: "orders"},
				Columns: []*plugin.Column{
					{Name: "id", NotNull: true, Type: &plugin.Identifier{Name: "integer"}},
					{Name: "status", NotNull: true, Type: &plugin.Identifier{Name: "text"}},
				},
			},
		}}}},
		Queries: []*plugin.Query{{
			Name: "GetOrder",
			Cmd:  ":one",
			Text: "SELECT sqlc.embed(orders) FROM orders",
			Columns: []*plugin.Column{{
				Name:       "orders",
				EmbedTable: &plugin.Identifier{Name: "orders"},
			}},
		}},
	}

	m := New(req, logger.New(false))
	queries := m.MapQueries(req)
	if len(queries) != 1 {
		t.Fatalf("expected one query, got %d", len(queries))
	}

	query := queries[0]
	if !query.HasEmbeds || len(query.EmbedGroups) != 1 {
		t.Fatalf("expected one embed group: %#v", query)
	}
	if len(query.Results) != 2 {
		t.Fatalf("expected expanded embed results, got %#v", query.Results)
	}
	if query.Results[0].Name != "orders_id" || query.Results[0].EmbedTable != "orders" || !query.Results[0].IsAliased {
		t.Fatalf("unexpected first embed result: %#v", query.Results[0])
	}
	if query.EmbedGroups[0].Fields[1].Name != "orders_status" {
		t.Fatalf("unexpected embed group fields: %#v", query.EmbedGroups[0].Fields)
	}
}
