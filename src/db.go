package bayesh

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
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

func (q *Queries) ConditionalEventCounts(ctx context.Context, cwd *string, previousCmd *string, minEventCount *int) (map[string]int, error) {
	newLine := "\n"

	query := strings.Builder{}
	query.WriteString("SELECT " + colCurrentCmd + ", SUM(" + colEventCounter + ")" + newLine)
	query.WriteString("FROM " + eventsTable + newLine)

	conditions := []string{}
	args := []interface{}{}

	if cwd != nil {
		conditions = append(conditions, colCwd+" = ?")
		args = append(args, *cwd)
	}
	if previousCmd != nil {
		conditions = append(conditions, colPreviousCmd+" = ?")
		args = append(args, *previousCmd)
	}
	if minEventCount != nil {
		conditions = append(conditions, colEventCounter+" >= ?")
		args = append(args, *minEventCount)
	}
	if len(conditions) > 0 {
		query.WriteString("WHERE " + strings.Join(conditions, " AND ") + newLine)
	}
	query.WriteString("GROUP BY " + colCurrentCmd + newLine)

	slog.Debug("Inferring current command with",
		"query", query.String(),
		"cwd", cwd,
		"previousCmd", previousCmd,
	)

	rows, err := q.db.QueryContext(ctx, query.String(), args...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Failed to close rows:", "error", err)
		}
	}()

	result := make(map[string]int)
	for rows.Next() {
		var currentCmd string
		var eventCounter int
		if err := rows.Scan(&currentCmd, &eventCounter); err != nil {
			return nil, err
		}
		result[currentCmd] = eventCounter
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
