package bayesh

import (
	"context"
	"database/sql"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
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
		colCwd: true, colPreviousCmd: true, colCurrentCmd: true, colPreviousTemplate: true, colCurrentTemplate: true, colEventCounter: true, colLastModified: true,
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

	// One covering index per ranking query, see TestTemplateCountsUsesCoveringIndexes
	if got, want := indexes(t, db), []string{"idx_cwd", "idx_cwd_previous", "idx_previous"}; !slices.Equal(got, want) {
		t.Errorf("indexes = %q, want %q", got, want)
	}
}

// indexes returns the names of the indexes created on the events table.
func indexes(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type = 'index' AND tbl_name = ? AND sql IS NOT NULL ORDER BY name", eventsTable)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		names = append(names, name)
	}
	return names
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
	err = db.QueryRow("SELECT cwd, previous_cmd, current_cmd, event_counter FROM "+eventsTable).Scan(
		&fetchedRow.Cwd,
		&fetchedRow.PreviousCmd,
		&fetchedRow.CurrentCmd,
		&fetchedRow.EventCounter,
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
		Cwd:              rowToInsert.Cwd,
		PreviousCmd:      rowToInsert.PreviousCmd,
		CurrentCmd:       rowToInsert.CurrentCmd,
		PreviousTemplate: "ls -l",
		CurrentTemplate:  "cat <ARG>",
		EventCounter:     456,
		LastModified:     time.Now().Truncate(time.Second).Add(time.Hour),
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
	if fetchedRow.PreviousTemplate != updatedRow.PreviousTemplate || fetchedRow.CurrentTemplate != updatedRow.CurrentTemplate {
		t.Errorf("Expected templates %q, %q, got %q, %q", updatedRow.PreviousTemplate, updatedRow.CurrentTemplate, fetchedRow.PreviousTemplate, fetchedRow.CurrentTemplate)
	}

	if !fetchedRow.LastModified.Equal(updatedRow.LastModified) {
		t.Errorf("Expected LastModified to be %v, got %v", updatedRow.LastModified, fetchedRow.LastModified)
	}
}

func TestTemplateCounts(t *testing.T) {
	queries, _ := setupTestDB(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	recent := now.Add(-time.Minute)
	for _, row := range []Row{
		{Cwd: "/a", PreviousCmd: "cd x", CurrentCmd: "ls", PreviousTemplate: "cd <ARG>", CurrentTemplate: "ls", EventCounter: 2, LastModified: recent},
		{Cwd: "/a", PreviousCmd: "vim y", CurrentCmd: "cd y", PreviousTemplate: "vim <ARG>", CurrentTemplate: "cd <ARG>", EventCounter: 1, LastModified: recent},
		{Cwd: "/a", PreviousCmd: "vim z", CurrentCmd: "cd z", PreviousTemplate: "vim <ARG>", CurrentTemplate: "cd <ARG>", EventCounter: 1, LastModified: recent},
		{Cwd: "/b", PreviousCmd: "cd w", CurrentCmd: "ls", PreviousTemplate: "cd <ARG>", CurrentTemplate: "ls", EventCounter: 1, LastModified: recent},
	} {
		if err := queries.InsertRow(ctx, row); err != nil {
			t.Fatalf("Failed to insert row: %v", err)
		}
	}
	a, cd, vim := "/a", "cd <ARG>", "vim <ARG>"
	tests := []struct {
		name     string
		cwd      *string
		previous *string
		want     map[string]float64
	}{
		{"by cwd", &a, nil, map[string]float64{"ls": 8, "cd <ARG>": 8}},
		{"by previous", nil, &cd, map[string]float64{"ls": 12}},
		{"by both", &a, &vim, map[string]float64{"cd <ARG>": 8}},
		{"no events", &a, &cd, map[string]float64{"ls": 8}},
	}
	for _, tt := range tests {
		got, err := queries.TemplateCounts(ctx, now, tt.cwd, tt.previous)
		if err != nil {
			t.Fatalf("%s: %v", tt.name, err)
		}
		if !maps.Equal(got, tt.want) {
			t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
		}
	}
}

// TestTemplateCountsFrecency checks that each row is weighted by the time
// since it was last recorded, like z weights directories.
func TestTemplateCountsFrecency(t *testing.T) {
	queries, _ := setupTestDB(t)
	ctx := context.Background()
	now := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	// Recorded in another time zone than now, which unix time is independent of
	zone := time.FixedZone("UTC+5", 5*60*60)
	for i, r := range []struct {
		age time.Duration
		n   int
	}{
		{0, 1},                    // 4
		{59 * time.Minute, 1},     // 4
		{61 * time.Minute, 2},     // 2*2
		{23 * time.Hour, 1},       // 2
		{25 * time.Hour, 4},       // 4*0.5
		{6 * 24 * time.Hour, 2},   // 2*0.5
		{8 * 24 * time.Hour, 4},   // 4*0.25
		{365 * 24 * time.Hour, 8}, // 8*0.25
	} {
		row := Row{Cwd: "/a", PreviousCmd: fmt.Sprint(i), CurrentCmd: "ls", CurrentTemplate: "ls", EventCounter: r.n, LastModified: now.Add(-r.age).In(zone)}
		if err := queries.InsertRow(ctx, row); err != nil {
			t.Fatal(err)
		}
	}
	// A row without a timestamp counts as old
	if _, err := queries.db.ExecContext(ctx, "INSERT INTO "+eventsTable+" (cwd, previous_cmd, current_cmd, current_template, event_counter) VALUES ('/a', 'x', 'ls', 'ls', 4)"); err != nil {
		t.Fatal(err)
	}
	a := "/a"
	got, err := queries.TemplateCounts(ctx, now, &a, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := map[string]float64{"ls": 4 + 4 + 4 + 2 + 2 + 1 + 1 + 2 + 1}; !maps.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestTemplateCountsUsesCoveringIndexes checks that each ranking query reads
// only the index rows of its matching events, so opening Bayesh stays fast.
func TestTemplateCountsUsesCoveringIndexes(t *testing.T) {
	_, db := setupTestDB(t)
	a, ls := "/a", "ls"
	for _, tt := range []struct {
		cwd, previous *string
		index         string
	}{{&a, nil, "idx_cwd"}, {nil, &ls, "idx_previous"}, {&a, &ls, "idx_cwd_previous"}} {
		query, args := templateCountsQuery(time.Now(), tt.cwd, tt.previous)
		rows, err := db.Query("EXPLAIN QUERY PLAN "+query, args...)
		if err != nil {
			t.Fatal(err)
		}
		var steps []string
		for rows.Next() {
			var id, parent, notused int
			var step string
			if err := rows.Scan(&id, &parent, &notused, &step); err != nil {
				t.Fatal(err)
			}
			steps = append(steps, step)
		}
		rows.Close()
		detail := strings.Join(steps, "; ")
		if want := "USING COVERING INDEX " + tt.index + " "; !strings.Contains(detail, want) || strings.Contains(detail, "TEMP B-TREE") {
			t.Errorf("plan %q, want %q without sorting", detail, want)
		}
	}
}

func TestSlotStatistics(t *testing.T) {
	queries, _ := setupTestDB(t)
	ctx := context.Background()
	if n, err := queries.SlotDistinct(ctx, "cd#1"); err != nil || n != 0 {
		t.Fatalf("SlotDistinct of an unseen slot = %d, %v", n, err)
	}
	if n, err := queries.SlotRuns(ctx, "cd#1", "a"); err != nil || n != 0 {
		t.Fatalf("SlotRuns of an unseen value = %d, %v", n, err)
	}
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, v := range []struct {
		slot, value string
		n           int
	}{{"cd#1", "a", 1}, {"cd#1", "b", 1}, {"cd#1", "a", 2}, {"vim#1", "a", 1}} {
		if err := queries.AddSlotRuns(ctx, v.slot, v.value, v.n, start.Add(time.Duration(i)*time.Hour)); err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range []struct {
		slot, value string
		want        int
	}{{"cd#1", "a", 3}, {"cd#1", "b", 1}, {"cd#1", "c", 0}, {"vim#1", "a", 1}} {
		if n, err := queries.SlotRuns(ctx, v.slot, v.value); err != nil || n != v.want {
			t.Errorf("SlotRuns(%q, %q) = %d, %v, want %d", v.slot, v.value, n, err, v.want)
		}
	}
	if n, err := queries.SlotDistinct(ctx, "cd#1"); err != nil || n != 2 {
		t.Errorf("SlotDistinct(cd#1) = %d, %v, want 2", n, err)
	}
	// last_modified is when a value was last run (unix time, like in the events
	// table), so stale values can be cleaned up
	for _, v := range []struct {
		slot, value string
		want        int64
	}{{"cd#1", "a", start.Add(2 * time.Hour).Unix()}, {"cd#1", "b", start.Add(time.Hour).Unix()}} {
		var got int64
		var typ string
		err := queries.db.QueryRowContext(ctx, "SELECT "+colLastModified+", typeof("+colLastModified+") FROM "+slotsTable+" WHERE "+colSlot+" = ? AND "+colValue+" = ?", v.slot, v.value).Scan(&got, &typ)
		if err != nil || got != v.want || typ != "integer" {
			t.Errorf("last_modified of %q in %q = %v (%s), %v, want integer %v", v.value, v.slot, got, typ, err, v.want)
		}
	}
}

func TestRowsMentioningAndSetTemplates(t *testing.T) {
	queries, _ := setupTestDB(t)
	ctx := context.Background()
	for _, row := range []Row{
		{Cwd: "/a", PreviousCmd: "ls", CurrentCmd: "cd x", PreviousTemplate: "ls", CurrentTemplate: "cd x"},
		{Cwd: "/a", PreviousCmd: "cd y", CurrentCmd: "ls", PreviousTemplate: "cd y", CurrentTemplate: "ls"},
		{Cwd: "/a", PreviousCmd: "ls", CurrentCmd: "vim z", PreviousTemplate: "ls", CurrentTemplate: "vim z"},
	} {
		if err := queries.InsertRow(ctx, row); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := queries.RowsMentioning(ctx, []string{"cd"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].CurrentCmd != "cd x" || rows[1].PreviousCmd != "cd y" {
		t.Fatalf("got %+v, want the two rows with cd", rows)
	}
	if err := queries.SetTemplates(ctx, rows[0].ID, "ls", "cd <ARG>"); err != nil {
		t.Fatal(err)
	}
	row, err := queries.GetRow(ctx, "/a", "ls", "cd x")
	if err != nil {
		t.Fatal(err)
	}
	if row.PreviousTemplate != "ls" || row.CurrentTemplate != "cd <ARG>" {
		t.Errorf("templates not set: %+v", row)
	}
}
