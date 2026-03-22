package builders

import (
	"testing"

	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/logger"
	"github.com/eikster-dk/sqlc-gen-better-typescript/cmd/plugin/internal/models"
)

func TestSQLTransformer_Transform(t *testing.T) {
	noopLogger := logger.New(false)

	tests := []struct {
		name                 string
		sql                  string
		params               []models.Param
		expected             string
		expectError          bool
		expectedReplacements int
	}{
		{
			name: "simple where clause",
			sql:  "SELECT * FROM customers WHERE id = $1",
			params: []models.Param{
				{Name: "id", Position: 1},
			},
			expected:             "SELECT * FROM customers WHERE id = ${params.id}",
			expectError:          false,
			expectedReplacements: 1,
		},
		{
			name: "multiple params",
			sql:  "INSERT INTO customers (email, name, phone) VALUES ($1, $2, $3)",
			params: []models.Param{
				{Name: "email", Position: 1},
				{Name: "name", Position: 2},
				{Name: "phone", Position: 3},
			},
			expected:             "INSERT INTO customers (email, name, phone) VALUES (${params.email}, ${params.name}, ${params.phone})",
			expectError:          false,
			expectedReplacements: 3,
		},
		{
			name: "type cast with array",
			sql:  "WHERE id = ANY($1::int[])",
			params: []models.Param{
				{Name: "ids", Position: 1},
			},
			expected:             "WHERE id = ANY(${params.ids}::int[])",
			expectError:          false,
			expectedReplacements: 1,
		},
		{
			name: "string concatenation",
			sql:  "WHERE name ILIKE '%' || $1 || '%' ORDER BY name",
			params: []models.Param{
				{Name: "arg1", Position: 1},
			},
			expected:             "WHERE name ILIKE '%' || ${params.arg1} || '%' ORDER BY name",
			expectError:          false,
			expectedReplacements: 1,
		},
		{
			name: "duplicate param usage",
			sql:  "WHERE id = $1 OR parent_id = $1",
			params: []models.Param{
				{Name: "id", Position: 1},
			},
			expected:             "WHERE id = ${params.id} OR parent_id = ${params.id}",
			expectError:          false,
			expectedReplacements: 2,
		},
		{
			name: "complex query from orders.sql",
			sql: `SELECT 
    o.id, o.customer_id, o.status, o.total_cents, o.shipping_address, o.billing_address, o.notes, o.created_at, o.updated_at
FROM orders o
WHERE o.id = $1`,
			params: []models.Param{
				{Name: "id", Position: 1},
			},
			expected: `SELECT 
    o.id, o.customer_id, o.status, o.total_cents, o.shipping_address, o.billing_address, o.notes, o.created_at, o.updated_at
FROM orders o
WHERE o.id = ${params.id}`,
			expectError:          false,
			expectedReplacements: 1,
		},
		{
			name: "LIMIT and OFFSET",
			sql:  "SELECT * FROM customers ORDER BY created_at DESC LIMIT $1 OFFSET $2",
			params: []models.Param{
				{Name: "limit", Position: 1},
				{Name: "offset", Position: 2},
			},
			expected:             "SELECT * FROM customers ORDER BY created_at DESC LIMIT ${params.limit} OFFSET ${params.offset}",
			expectError:          false,
			expectedReplacements: 2,
		},
		{
			name: "missing placeholder should error",
			sql:  "SELECT * FROM customers WHERE id = $1",
			params: []models.Param{
				{Name: "id", Position: 1},
				{Name: "name", Position: 2}, // $2 is expected but not in SQL
			},
			expected:             "", // Will error because $2 not found
			expectError:          true,
			expectedReplacements: 0,
		},
		{
			name: "high position numbers",
			sql:  "VALUES ($1, $2, $10)",
			params: []models.Param{
				{Name: "a", Position: 1},
				{Name: "b", Position: 2},
				{Name: "j", Position: 10},
			},
			expected:             "VALUES (${params.a}, ${params.b}, ${params.j})",
			expectError:          false,
			expectedReplacements: 3,
		},
		{
			name: "many params - $1 should not corrupt $10 $11 $12",
			sql:  "INSERT INTO t (a,b,c,d,e,f,g,h,i,j,k,l) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)",
			params: []models.Param{
				{Name: "a", Position: 1},
				{Name: "b", Position: 2},
				{Name: "c", Position: 3},
				{Name: "d", Position: 4},
				{Name: "e", Position: 5},
				{Name: "f", Position: 6},
				{Name: "g", Position: 7},
				{Name: "h", Position: 8},
				{Name: "i", Position: 9},
				{Name: "j", Position: 10},
				{Name: "k", Position: 11},
				{Name: "l", Position: 12},
			},
			expected:             "INSERT INTO t (a,b,c,d,e,f,g,h,i,j,k,l) VALUES (${params.a},${params.b},${params.c},${params.d},${params.e},${params.f},${params.g},${params.h},${params.i},${params.j},${params.k},${params.l})",
			expectError:          false,
			expectedReplacements: 12,
		},
	}

	transformer := &SQLTransformer{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := transformer.Transform(tt.sql, tt.params, noopLogger)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.TemplateLiteral != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expected, result.TemplateLiteral)
			}

			if result.ReplacementsMade != tt.expectedReplacements {
				t.Errorf("Expected %d replacements, got %d", tt.expectedReplacements, result.ReplacementsMade)
			}

			if result.OriginalSQL != tt.sql {
				t.Errorf("OriginalSQL should be preserved, expected:\n%s\nGot:\n%s", tt.sql, result.OriginalSQL)
			}
		})
	}
}
