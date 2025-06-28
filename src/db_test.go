package bayesh

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCreateDB(t *testing.T) {
	dbFile := "test_create.db"

	if err := CreateDB(dbFile); err != nil {
		t.Fatalf("CreateDB failed: %v", err)
	}
	defer func() {
		if err := os.Remove(dbFile); err != nil {
			t.Fatalf("Failed to remove test DB file: %v", err)
		}
	}()

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close DB: %v", err)
		}
	}()

	// Check table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", eventsTable).Scan(&tableName)
	if err != nil {
		t.Fatalf("Table %s does not exist: %v", eventsTable, err)
	}

	// Check columns
	cols, err := db.Query("PRAGMA table_info(" + eventsTable + ")")
	if err != nil {
		t.Fatalf("Failed to get table info: %v", err)
	}
	defer func() {
		if err := cols.Close(); err != nil {
			t.Fatalf("Failed to close columns query: %v", err)
		}
	}()

	expected := map[string]bool{
		colCwd: true, colPreviousCmd: true, colCurrentCmd: true, colEventCounter: true, colLastModified: true,
	}
	for cols.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dfltValue interface{}
		if err := cols.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			t.Fatalf("Failed to scan column info: %v", err)
		}
		expected[name] = false
	}
	for col, missing := range expected {
		if missing {
			t.Errorf("Expected column %s not found", col)
		}
	}

	// Check index exists
	var idxName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_event_counter'").Scan(&idxName)
	if err != nil {
		t.Errorf("Index idx_event_counter does not exist: %v", err)
	}
}
