package bayesh

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupCoreWithTempDB(t *testing.T) (*Core, string) {
	t.Helper()

	tempDir := t.TempDir()
	t.Setenv(BayeshDirEnvVar, tempDir)

	settings, err := Setup(context.Background(), OsFs{})
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	core, err := NewCore(context.Background(), settings)
	if err != nil {
		t.Fatalf("NewCore failed: %v", err)
	}

	return core, settings.Database
}

func TestRecordEvent_NewEvent(t *testing.T) {
	core, dbPath := setupCoreWithTempDB(t)
	defer func() {
		if err := core.Close(); err != nil {
			t.Errorf("core.Close() failed: %v", err)
		}
	}()

	ctx := context.Background()
	cwd := "/home/user/project"
	previousCmd := "ls -l"
	currentCmd := "git status"

	err := core.RecordEvent(ctx, cwd, previousCmd, currentCmd)
	if err != nil {
		t.Fatalf("RecordEvent failed: %v", err)
	}

	// Verify the data was inserted correctly
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open DB for verification: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	}()

	var eventCounter int
	err = db.QueryRow("SELECT event_counter FROM events WHERE cwd = ? AND previous_cmd = ? AND current_cmd = ?", cwd, ProcessCmd(OsFs{}, previousCmd), ProcessCmd(OsFs{}, currentCmd)).Scan(&eventCounter)
	if err != nil {
		t.Fatalf("Failed to query for new event: %v", err)
	}

	if eventCounter != 1 {
		t.Errorf("expected event_counter to be 1, got %d", eventCounter)
	}
}

func TestRecordEvent_ExistingEvent(t *testing.T) {
	core, dbPath := setupCoreWithTempDB(t)
	defer func() {
		if err := core.Close(); err != nil {
			t.Errorf("core.Close() failed: %v", err)
		}
	}()

	ctx := context.Background()
	cwd := "/home/user/project"
	previousCmd := "ls -l"
	currentCmd := "git status"

	// Record the event for the first time
	if err := core.RecordEvent(ctx, cwd, previousCmd, currentCmd); err != nil {
		t.Fatalf("First RecordEvent failed: %v", err)
	}

	// Verify initial state and get timestamp
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open DB for verification: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	}()

	var initialTimestamp time.Time
	processedPrevCmd := ProcessCmd(OsFs{}, previousCmd)
	processedCurrCmd := ProcessCmd(OsFs{}, currentCmd)
	err = db.QueryRow("SELECT last_modified FROM events WHERE cwd = ? AND previous_cmd = ? AND current_cmd = ?", cwd, processedPrevCmd, processedCurrCmd).Scan(&initialTimestamp)
	if err != nil {
		t.Fatalf("Failed to query for initial timestamp: %v", err)
	}

	// Wait a moment to ensure the timestamp will be different
	time.Sleep(2 * time.Millisecond)

	// Record the same event again
	if err := core.RecordEvent(ctx, cwd, previousCmd, currentCmd); err != nil {
		t.Fatalf("Second RecordEvent failed: %v", err)
	}

	// Verify the event was updated
	var updatedCounter int
	var updatedTimestamp time.Time
	err = db.QueryRow("SELECT event_counter, last_modified FROM events WHERE cwd = ? AND previous_cmd = ? AND current_cmd = ?", cwd, processedPrevCmd, processedCurrCmd).Scan(&updatedCounter, &updatedTimestamp)
	if err != nil {
		t.Fatalf("Failed to query for updated event: %v", err)
	}

	if updatedCounter != 2 {
		t.Errorf("expected event_counter to be 2, got %d", updatedCounter)
	}
	if !updatedTimestamp.After(initialTimestamp) {
		t.Errorf("expected updated timestamp %v to be after initial timestamp %v", updatedTimestamp, initialTimestamp)
	}
}

func TestInferCommands(t *testing.T) {
	core, dbPath := setupCoreWithTempDB(t)
	defer func() {
		if err := core.Close(); err != nil {
			t.Errorf("core.Close() failed: %v", err)
		}
	}()

	ctx := context.Background()
	cwd := "/home/user/project"
	previousCmd := "git status"
	expectedCurrentCmd := "git add ."

	// Manually insert a row for testing inference
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open DB for data insertion: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() failed: %v", err)
		}
	}()

	queries := New(db)
	insertRow := Row{
		Cwd:          cwd,
		PreviousCmd:  ProcessCmd(OsFs{}, previousCmd),
		CurrentCmd:   expectedCurrentCmd,
		EventCounter: 1,
		LastModified: time.Now(),
	}
	if err := queries.InsertRow(ctx, insertRow); err != nil {
		t.Fatalf("Failed to insert test row: %v", err)
	}

	inferredCmds, err := core.InferCommands(ctx, cwd, previousCmd)
	if err != nil {
		t.Fatalf("InferCommands failed: %v", err)
	}

	if len(inferredCmds) != 1 || inferredCmds[0] != expectedCurrentCmd {
		t.Errorf("expected inferred command [%q], but got %q", expectedCurrentCmd, inferredCmds)
	}
}
