package bayesh

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

const (
	eventsTable     = "events"
	colCwd          = "cwd"
	colPreviousCmd  = "previous_cmd"
	colCurrentCmd   = "current_cmd"
	colEventCounter = "event_counter"
	colLastModified = "last_modified"
)

func CreateDB(dbPath string) error {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("Warning: failed to close database connection: %v\n", closeErr)
		}
	}()

	schema := "CREATE TABLE " + eventsTable + " (" +
		colCwd + " TEXT, " +
		colPreviousCmd + " TEXT, " +
		colCurrentCmd + " TEXT, " +
		colEventCounter + " INTEGER CHECK(" + colEventCounter + " >= 0), " +
		colLastModified + " DATETIME, " +
		"PRIMARY KEY (" + colCwd + ", " + colPreviousCmd + ", " + colCurrentCmd + ")" +
		");" +
		"CREATE INDEX idx_event_counter ON " + eventsTable +
		" (" + colEventCounter + ");"

	_, err = db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema creation statements: %w", err)
	}

	return nil
}
