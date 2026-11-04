package bayesh

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a new temporary database for testing and returns the queries object and the db object.
// It also registers a cleanup function to close and remove the database file.
func setupTestDB(t *testing.T) (*Queries, *sql.DB) {
	t.Helper()

	tempFile, err := os.CreateTemp(t.TempDir(), "test-*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file for db: %v", err)
	}
	dbFile := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		t.Fatalf("Failed to open DB: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("Failed to close DB: %v", err)
		}
		// os.Remove is not needed here because t.TempDir() handles cleanup
	})

	queries := New(db)
	if err := queries.CreateSchema(context.Background()); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}

	return queries, db
}

func TestCreateDB(t *testing.T) {
	_, db := setupTestDB(t)

	// Check table exists
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", eventsTable).Scan(&tableName)
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

func TestInsertRow(t *testing.T) {
	queries, db := setupTestDB(t)
	ctx := context.Background()

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM " + eventsTable).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}
	if count != 0 {
		t.Fatalf("Expected 0 rows, got %d", count)
	}

	rowToInsert := Row{
		Cwd:          "/tmp",
		PreviousCmd:  "ls -l",
		CurrentCmd:   "cat file.txt",
		EventCounter: 123,
		LastModified: time.Now(),
	}

	if err := queries.InsertRow(ctx, rowToInsert); err != nil {
		t.Fatalf("Failed to insert row: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM " + eventsTable).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("Expected 1 row, got %d", count)
	}

	var fetchedRow Row
	err = db.QueryRow("SELECT cwd, previous_cmd, current_cmd, event_counter, last_modified FROM "+eventsTable).Scan(
		&fetchedRow.Cwd,
		&fetchedRow.PreviousCmd,
		&fetchedRow.CurrentCmd,
		&fetchedRow.EventCounter,
		&fetchedRow.LastModified,
	)
	if err != nil {
		t.Fatalf("Failed to fetch inserted row: %v", err)
	}

	if fetchedRow.Cwd != rowToInsert.Cwd {
		t.Errorf("Expected Cwd to be %s, got %s", rowToInsert.Cwd, fetchedRow.Cwd)
	}
	if fetchedRow.PreviousCmd != rowToInsert.PreviousCmd {
		t.Errorf("Expected PreviousCmd to be %s, got %s", rowToInsert.PreviousCmd, fetchedRow.PreviousCmd)
	}
	if fetchedRow.CurrentCmd != rowToInsert.CurrentCmd {
		t.Errorf("Expected CurrentCmd to be %s, got %s", rowToInsert.CurrentCmd, fetchedRow.CurrentCmd)
	}
	if fetchedRow.EventCounter != rowToInsert.EventCounter {
		t.Errorf("Expected EventCounter to be %d, got %d", rowToInsert.EventCounter, fetchedRow.EventCounter)
	}
}

func TestInsertDuplicateRow(t *testing.T) {
	queries, _ := setupTestDB(t)
	ctx := context.Background()

	rowToInsert := Row{
		Cwd:          "/tmp",
		PreviousCmd:  "ls -l",
		CurrentCmd:   "cat file.txt",
		EventCounter: 123,
		LastModified: time.Now(),
	}

	if err := queries.InsertRow(ctx, rowToInsert); err != nil {
		t.Fatalf("Failed to insert first row: %v", err)
	}

	// Create a new row with the same primary key but different non-primary key data

	duplicateRow := Row{
		Cwd:          rowToInsert.Cwd,
		PreviousCmd:  rowToInsert.PreviousCmd,
		CurrentCmd:   rowToInsert.CurrentCmd,
		EventCounter: 456,                         // Different event counter
		LastModified: time.Now().Add(time.Second), // Different timestamp
	}

	if err := queries.InsertRow(ctx, duplicateRow); err == nil {
		t.Fatal("Expected error when inserting row with duplicate primary key, but got nil")
	}
}

func TestGetRow(t *testing.T) {
	queries, _ := setupTestDB(t)
	ctx := context.Background()

	// 1. Test getting a non-existent row
	_, err := queries.GetRow(ctx, "non-existent", "non-existent", "non-existent")
	if err == nil {
		t.Fatal("Expected an error when getting a non-existent row, but got nil")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("Expected sql.ErrNoRows, but got %v", err)
	}

	// 2. Test getting an existing row

	rowToInsert := Row{
		Cwd:          "/tmp",
		PreviousCmd:  "ls -l",
		CurrentCmd:   "cat file.txt",
		EventCounter: 123,
		LastModified: time.Now().Truncate(time.Second), // Truncate for reliable comparison
	}

	if err := queries.InsertRow(ctx, rowToInsert); err != nil {
		t.Fatalf("Failed to insert row for testing GetRow: %v", err)
	}

	fetchedRow, err := queries.GetRow(ctx, rowToInsert.Cwd, rowToInsert.PreviousCmd, rowToInsert.CurrentCmd)
	if err != nil {
		t.Fatalf("Expected no error when getting an existing row, but got: %v", err)
	}

	// Compare the fields
	if fetchedRow.Cwd != rowToInsert.Cwd {
		t.Errorf("Expected Cwd to be %s, got %s", rowToInsert.Cwd, fetchedRow.Cwd)
	}
	if fetchedRow.PreviousCmd != rowToInsert.PreviousCmd {
		t.Errorf("Expected PreviousCmd to be %s, got %s", rowToInsert.PreviousCmd, fetchedRow.PreviousCmd)
	}
	if fetchedRow.CurrentCmd != rowToInsert.CurrentCmd {
		t.Errorf("Expected CurrentCmd to be %s, got %s", rowToInsert.CurrentCmd, fetchedRow.CurrentCmd)
	}
	if fetchedRow.EventCounter != rowToInsert.EventCounter {
		t.Errorf("Expected EventCounter to be %d, got %d", rowToInsert.EventCounter, fetchedRow.EventCounter)
	}
	if !fetchedRow.LastModified.Truncate(time.Second).Equal(rowToInsert.LastModified) {
		t.Errorf("Expected LastModified to be %v, got %v", rowToInsert.LastModified, fetchedRow.LastModified)
	}
}

func TestUpdateRow(t *testing.T) {
	queries, _ := setupTestDB(t)
	ctx := context.Background()

	// 1. Insert a row to be updated

	rowToInsert := Row{
		Cwd:          "/tmp",
		PreviousCmd:  "ls -l",
		CurrentCmd:   "cat file.txt",
		EventCounter: 123,
		LastModified: time.Now().Truncate(time.Second),
	}
	if err := queries.InsertRow(ctx, rowToInsert); err != nil {
		t.Fatalf("Failed to insert initial row: %v", err)
	}

	// 2. Update the row with new data

	updatedRow := Row{
		Cwd:          rowToInsert.Cwd,
		PreviousCmd:  rowToInsert.PreviousCmd,
		CurrentCmd:   rowToInsert.CurrentCmd,
		EventCounter: 456,
		LastModified: time.Now().Truncate(time.Second).Add(time.Hour),
	}
	if err := queries.UpdateRow(ctx, updatedRow); err != nil {
		t.Fatalf("Failed to update row: %v", err)
	}

	// 3. Get the row again and verify the changes

	fetchedRow, err := queries.GetRow(ctx, rowToInsert.Cwd, rowToInsert.PreviousCmd, rowToInsert.CurrentCmd)
	if err != nil {
		t.Fatalf("Failed to get updated row: %v", err)
	}

	if fetchedRow.EventCounter != updatedRow.EventCounter {
		t.Errorf("Expected EventCounter to be %d, got %d", updatedRow.EventCounter, fetchedRow.EventCounter)
	}

	if !fetchedRow.LastModified.Equal(updatedRow.LastModified) {
		t.Errorf("Expected LastModified to be %v, got %v", updatedRow.LastModified, fetchedRow.LastModified)
	}
}

func TestEvents(t *testing.T) {
	queries, _ := setupTestDB(t)
	ctx := context.Background()

	events, err := queries.Events(ctx)
	if err != nil {
		t.Fatalf("Events failed on empty database: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("Expected no events, got %d", len(events))
	}

	rows := []Row{
		{Cwd: "/a", PreviousCmd: "ls", CurrentCmd: "cd b", EventCounter: 2, LastModified: time.Now().UTC()},
		{Cwd: "/b", PreviousCmd: "cd b", CurrentCmd: `git commit -m "a b"`, EventCounter: 1, LastModified: time.Now().UTC()},
	}
	for _, row := range rows {
		if err := queries.InsertRow(ctx, row); err != nil {
			t.Fatalf("Failed to insert row: %v", err)
		}
	}

	events, err = queries.Events(ctx)
	if err != nil {
		t.Fatalf("Events failed: %v", err)
	}
	if len(events) != len(rows) {
		t.Fatalf("Expected %d events, got %d", len(rows), len(events))
	}
	for i, row := range rows {
		got := events[i]
		if got.Cwd != row.Cwd || got.PreviousCmd != row.PreviousCmd || got.CurrentCmd != row.CurrentCmd || got.EventCounter != row.EventCounter {
			t.Errorf("Event %d: got %+v, want %+v", i, got, row)
		}
	}
}
