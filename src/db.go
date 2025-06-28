package bayesh

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	eventsTable     = "events"
	colCwd          = "cwd"
	colPreviousCmd  = "previous_cmd"
	colCurrentCmd   = "current_cmd"
	colEventCounter = "event_counter"
	colLastModified = "last_modified"
)

type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

type Queries struct {
	db DBTX
}

func New(db *sql.DB) *Queries {
	return &Queries{db: db}
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
	return &Queries{db: tx}
}

func (q *Queries) CreateSchema(ctx context.Context) error {
	schema := `
    CREATE TABLE IF NOT EXISTS ` + eventsTable + ` (
        ` + colCwd + ` TEXT,
        ` + colPreviousCmd + ` TEXT,
        ` + colCurrentCmd + ` TEXT,
        ` + colEventCounter + ` INTEGER CHECK(` + colEventCounter + ` >= 0),
        ` + colLastModified + ` DATETIME,
        PRIMARY KEY (` + colCwd + `, ` + colPreviousCmd + `, ` + colCurrentCmd + `)
    );
    CREATE INDEX IF NOT EXISTS idx_event_counter ON ` + eventsTable + ` (` + colEventCounter + `);`

	_, err := q.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema creation statements: %w", err)
	}
	return nil
}

type Row struct {
	Cwd          string
	PreviousCmd  string
	CurrentCmd   string
	EventCounter int
	LastModified time.Time
}

func (q *Queries) InsertRow(ctx context.Context, arg Row) error {
	query := `
    INSERT INTO ` + eventsTable + ` (
        ` + colCwd + `, ` + colPreviousCmd + `, ` + colCurrentCmd + `, ` + colEventCounter + `, ` + colLastModified + `
    ) VALUES (?, ?, ?, ?, ?)`
	_, err := q.db.ExecContext(ctx, query, arg.Cwd, arg.PreviousCmd, arg.CurrentCmd, arg.EventCounter, arg.LastModified)
	return err
}

func (q *Queries) GetRow(ctx context.Context, cwd, previousCmd, currentCmd string) (Row, error) {
	query := `
    SELECT ` + colCwd + `, ` + colPreviousCmd + `, ` + colCurrentCmd + `, ` + colEventCounter + `, ` + colLastModified + `
    FROM ` + eventsTable + `
    WHERE ` + colCwd + ` = ? AND ` + colPreviousCmd + ` = ? AND ` + colCurrentCmd + ` = ?`

	row := q.db.QueryRowContext(ctx, query, cwd, previousCmd, currentCmd)
	var i Row
	err := row.Scan(
		&i.Cwd,
		&i.PreviousCmd,
		&i.CurrentCmd,
		&i.EventCounter,
		&i.LastModified,
	)
	return i, err
}

func (q *Queries) UpdateRow(ctx context.Context, arg Row) error {
	query := `
    UPDATE ` + eventsTable + `
    SET ` + colEventCounter + ` = ?, ` + colLastModified + ` = ?
    WHERE ` + colCwd + ` = ? AND ` + colPreviousCmd + ` = ? AND ` + colCurrentCmd + ` = ?`
	_, err := q.db.ExecContext(ctx, query, arg.EventCounter, arg.LastModified, arg.Cwd, arg.PreviousCmd, arg.CurrentCmd)
	return err
}

func (q *Queries) InferCurrentCmd(ctx context.Context, cwd, previousCmd string) ([]string, error) {
	query := `
    SELECT ` + colCurrentCmd + `
    FROM ` + eventsTable + `
    WHERE ` + colCwd + ` = ? AND ` + colPreviousCmd + ` = ?
    ORDER BY ` + colEventCounter + ` DESC`

	rows, err := q.db.QueryContext(ctx, query, cwd, previousCmd)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Fatal("Failed to close rows:", err)
		}
	}()

	var items []string
	for rows.Next() {
		var currentCmd string
		if err := rows.Scan(&currentCmd); err != nil {
			return nil, err
		}
		items = append(items, currentCmd)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
