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

	// The templates of previous_cmd and current_cmd, kept up to date when events are recorded
	colPreviousTemplate = "previous_template"
	colCurrentTemplate  = "current_template"

	// The number of runs of each value in each argument slot, see templater
	slotsTable = "slots"
	colSlot    = "slot"
	colValue   = "value"
	colRuns    = "runs"
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
        ` + colPreviousTemplate + ` TEXT,
        ` + colCurrentTemplate + ` TEXT,
        ` + colEventCounter + ` INTEGER CHECK(` + colEventCounter + ` >= 0),
        ` + colLastModified + ` DATETIME,
        PRIMARY KEY (` + colCwd + `, ` + colPreviousCmd + `, ` + colCurrentCmd + `)
    );
    CREATE INDEX IF NOT EXISTS idx_event_counter ON ` + eventsTable + ` (` + colEventCounter + `);
    CREATE INDEX IF NOT EXISTS idx_cwd_template ON ` + eventsTable + ` (` + colCwd + `, ` + colCurrentTemplate + `, ` + colEventCounter + `);
    CREATE INDEX IF NOT EXISTS idx_previous_template ON ` + eventsTable + ` (` + colPreviousTemplate + `, ` + colCurrentTemplate + `, ` + colEventCounter + `);
    CREATE TABLE IF NOT EXISTS ` + slotsTable + ` (
        ` + colSlot + ` TEXT,
        ` + colValue + ` TEXT,
        ` + colRuns + ` INTEGER,
        PRIMARY KEY (` + colSlot + `, ` + colValue + `)
    );`

	_, err := q.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to execute schema creation statements: %w", err)
	}
	return nil
}

type Row struct {
	Cwd              string
	PreviousCmd      string
	CurrentCmd       string
	PreviousTemplate string
	CurrentTemplate  string
	EventCounter     int
	LastModified     time.Time
}

func (q *Queries) InsertRow(ctx context.Context, arg Row) error {
	query := `
    INSERT INTO ` + eventsTable + ` (
        ` + colCwd + `, ` + colPreviousCmd + `, ` + colCurrentCmd + `, ` + colPreviousTemplate + `, ` + colCurrentTemplate + `, ` + colEventCounter + `, ` + colLastModified + `
    ) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := q.db.ExecContext(ctx, query, arg.Cwd, arg.PreviousCmd, arg.CurrentCmd, arg.PreviousTemplate, arg.CurrentTemplate, arg.EventCounter, arg.LastModified)
	return err
}

func (q *Queries) GetRow(ctx context.Context, cwd, previousCmd, currentCmd string) (Row, error) {
	query := `
    SELECT ` + colCwd + `, ` + colPreviousCmd + `, ` + colCurrentCmd + `, ` + colPreviousTemplate + `, ` + colCurrentTemplate + `, ` + colEventCounter + `, ` + colLastModified + `
    FROM ` + eventsTable + `
    WHERE ` + colCwd + ` = ? AND ` + colPreviousCmd + ` = ? AND ` + colCurrentCmd + ` = ?`

	row := q.db.QueryRowContext(ctx, query, cwd, previousCmd, currentCmd)
	var i Row
	err := row.Scan(
		&i.Cwd,
		&i.PreviousCmd,
		&i.CurrentCmd,
		&i.PreviousTemplate,
		&i.CurrentTemplate,
		&i.EventCounter,
		&i.LastModified,
	)
	return i, err
}

func (q *Queries) UpdateRow(ctx context.Context, arg Row) error {
	query := `
    UPDATE ` + eventsTable + `
    SET ` + colPreviousTemplate + ` = ?, ` + colCurrentTemplate + ` = ?, ` + colEventCounter + ` = ?, ` + colLastModified + ` = ?
    WHERE ` + colCwd + ` = ? AND ` + colPreviousCmd + ` = ? AND ` + colCurrentCmd + ` = ?`
	_, err := q.db.ExecContext(ctx, query, arg.PreviousTemplate, arg.CurrentTemplate, arg.EventCounter, arg.LastModified, arg.Cwd, arg.PreviousCmd, arg.CurrentCmd)
	return err
}

// TemplateCounts returns the number of events of each current template, given
// the cwd and/or the previous template if not nil. Templates with fewer than
// minEventCount events are left out.
func (q *Queries) TemplateCounts(ctx context.Context, cwd *string, previousTemplate *string, minEventCount int) (map[string]int, error) {
	var conditions []string
	var args []any
	if cwd != nil {
		conditions = append(conditions, colCwd+" = ?")
		args = append(args, *cwd)
	}
	if previousTemplate != nil {
		conditions = append(conditions, colPreviousTemplate+" = ?")
		args = append(args, *previousTemplate)
	}
	query := "SELECT " + colCurrentTemplate + ", SUM(" + colEventCounter + ") FROM " + eventsTable
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " GROUP BY " + colCurrentTemplate + " HAVING SUM(" + colEventCounter + ") >= ?"
	args = append(args, minEventCount)

	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)
	result := map[string]int{}
	for rows.Next() {
		var template string
		var count int
		if err := rows.Scan(&template, &count); err != nil {
			return nil, err
		}
		result[template] = count
	}
	return result, rows.Err()
}

// SlotDistinct returns the number of distinct values seen in slot.
func (q *Queries) SlotDistinct(ctx context.Context, slot string) (int, error) {
	var n int
	err := q.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+slotsTable+" WHERE "+colSlot+" = ?", slot).Scan(&n)
	return n, err
}

// SlotRuns returns the number of runs of value in slot.
func (q *Queries) SlotRuns(ctx context.Context, slot, value string) (int, error) {
	var n int
	err := q.db.QueryRowContext(ctx, "SELECT "+colRuns+" FROM "+slotsTable+" WHERE "+colSlot+" = ? AND "+colValue+" = ?", slot, value).Scan(&n)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return n, err
}

// AddSlotRuns adds n runs of value in slot.
func (q *Queries) AddSlotRuns(ctx context.Context, slot, value string, n int) error {
	query := `
    INSERT INTO ` + slotsTable + ` (` + colSlot + `, ` + colValue + `, ` + colRuns + `) VALUES (?, ?, ?)
    ON CONFLICT (` + colSlot + `, ` + colValue + `) DO UPDATE SET ` + colRuns + ` = ` + colRuns + ` + excluded.` + colRuns
	_, err := q.db.ExecContext(ctx, query, slot, value, n)
	return err
}

// TemplatedRow holds the commands of an event and their stored templates.
type TemplatedRow struct {
	ID                                int64
	PreviousCmd, CurrentCmd           string
	PreviousTemplate, CurrentTemplate string
}

// RowsMentioning returns the events whose previous or current command contains any of words.
func (q *Queries) RowsMentioning(ctx context.Context, words []string) ([]TemplatedRow, error) {
	var conditions []string
	var args []any
	for _, w := range words {
		conditions = append(conditions, "instr("+colPreviousCmd+", ?) > 0 OR instr("+colCurrentCmd+", ?) > 0")
		args = append(args, w, w)
	}
	query := "SELECT rowid, " + colPreviousCmd + ", " + colCurrentCmd + ", " + colPreviousTemplate + ", " + colCurrentTemplate +
		" FROM " + eventsTable + " WHERE " + strings.Join(conditions, " OR ")
	rows, err := q.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer closeRows(rows)
	var result []TemplatedRow
	for rows.Next() {
		var r TemplatedRow
		if err := rows.Scan(&r.ID, &r.PreviousCmd, &r.CurrentCmd, &r.PreviousTemplate, &r.CurrentTemplate); err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

// SetTemplates sets the templates of the event with the given id.
func (q *Queries) SetTemplates(ctx context.Context, id int64, previousTemplate, currentTemplate string) error {
	query := "UPDATE " + eventsTable + " SET " + colPreviousTemplate + " = ?, " + colCurrentTemplate + " = ? WHERE rowid = ?"
	_, err := q.db.ExecContext(ctx, query, previousTemplate, currentTemplate, id)
	return err
}

func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		slog.Error("Failed to close rows:", "error", err)
	}
}
