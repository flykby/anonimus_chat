package db

import (
	"context"
	"os"
	"slices"
	"testing"
)

func TestSchemaTables(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	t.Cleanup(pool.Close)

	expected := []string{
		"users",
		"profiles",
		"premium_subscriptions",
		"personas",
		"photos",
		"dialogs",
		"dialog_messages",
		"dialog_photos_sent",
		"events",
		"deletion_benefits",
	}

	rows, err := pool.Query(ctx, `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public'
		  AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`)
	if err != nil {
		t.Fatalf("query tables: %v", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan: %v", err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}

	for _, table := range expected {
		if !slices.Contains(tables, table) {
			t.Fatalf("missing table %q, got %v", table, tables)
		}
	}
}
