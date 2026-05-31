package models

import "testing"

func TestRewriteSQLWithAliases(t *testing.T) {
	sql := "SELECT users.id, posts.id FROM users JOIN posts ON posts.user_id = users.id"
	rewritten := RewriteSQLWithAliases(sql, []ResultField{
		{Name: "id", OriginalName: "id", Table: "users"},
		{Name: "posts_id", OriginalName: "id", Table: "posts", IsAliased: true},
	})
	expected := "SELECT users.id, posts.id AS posts_id FROM users JOIN posts ON posts.user_id = users.id"
	if rewritten != expected {
		t.Fatalf("expected %q, got %q", expected, rewritten)
	}
}
