package bayesh

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"slices"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const probabilityWeight float64 = (float64(1.0) / float64(3.0))

type OsFs struct{}

func (OsFs) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (OsFs) Create(name string) (*os.File, error) {
	return os.Create(name)
}
func (OsFs) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}
func (OsFs) Getenv(key string) string {
	return os.Getenv(key)
}
func (OsFs) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

type Core struct {
	Settings *Settings
	db       *sql.DB
}

func NewCore(ctx context.Context, settings *Settings) (*Core, error) {
	db, err := sql.Open("sqlite3", settings.Database)
	if err != nil {
		return nil, err
	}

	return &Core{
		Settings: settings,
		db:       db,
	}, nil
}

// Close should be called to gracefully close the database connection.
func (c *Core) Close() error {
	if c != nil && c.db != nil {
		return c.db.Close()
	}
	return nil
}

func conditionalProbabilities(ctx context.Context, queries *Queries, settings *Settings, cwd *string, processedPreviousCmd *string, channel chan map[string]float64) {

	eventCounts, err := queries.ConditionalEventCounts(ctx, cwd, processedPreviousCmd, &settings.MinRequiredEvents)
	if err != nil {
		slog.Error("Failed to compute conditional events:" + err.Error())
	}

	totalCount := 0
	for _, count := range eventCounts {
		totalCount += count
	}

	probabilities := make(map[string]float64)
	for cmd, count := range eventCounts {
		probabilities[cmd] = float64(count) / float64(totalCount)
	}
	channel <- probabilities

}

func (c *Core) InferCommands(ctx context.Context, cwd string, previousCmd string) ([]string, error) {

	inferredCmdsMap := make(map[string]float64)

	queries := New(c.db)
	processedPreviousCmd := ProcessCmd(OsFs{}, previousCmd)

	chanCwd := make(chan map[string]float64)
	chanPrevCmd := make(chan map[string]float64)
	chanCwdPrevCmd := make(chan map[string]float64)
	go conditionalProbabilities(ctx, queries, c.Settings, &cwd, nil, chanCwd)
	go conditionalProbabilities(ctx, queries, c.Settings, nil, &processedPreviousCmd, chanPrevCmd)
	go conditionalProbabilities(ctx, queries, c.Settings, &cwd, &processedPreviousCmd, chanCwdPrevCmd)

	probCwd := <-chanCwd
	probPrevCmd := <-chanPrevCmd
	probCwdPrevCmd := <-chanCwdPrevCmd

	for cmd, prob := range probCwd {
		inferredCmdsMap[cmd] += probabilityWeight * prob
	}
	for cmd, prob := range probPrevCmd {
		inferredCmdsMap[cmd] += probabilityWeight * prob
	}
	for cmd, prob := range probCwdPrevCmd {
		inferredCmdsMap[cmd] += probabilityWeight * prob
	}

	keys := make([]string, 0, len(inferredCmdsMap))
	for cmd := range inferredCmdsMap {
		keys = append(keys, cmd)
	}

	slices.SortFunc(keys, func(a, b string) int {
		if inferredCmdsMap[a] > inferredCmdsMap[b] {
			return -1
		} else if inferredCmdsMap[a] < inferredCmdsMap[b] {
			return 1
		}
		return 0
	})

	return keys, nil
}

func (c *Core) RecordEvent(ctx context.Context, cwd string, previousCmd string, currentCmd string) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != sql.ErrTxDone {
			log.Fatalf("Error rolling back transaction: %v", err)
		}
	}()

	queries := New(c.db).WithTx(tx)
	processedPreviousCmd := ProcessCmd(OsFs{}, previousCmd)
	processedCurrentCmd := ProcessCmd(OsFs{}, currentCmd)

	row, err := queries.GetRow(ctx, cwd, processedPreviousCmd, processedCurrentCmd)
	if err != nil {
		if err == sql.ErrNoRows {
			newRow := Row{
				Cwd:          cwd,
				PreviousCmd:  processedPreviousCmd,
				CurrentCmd:   processedCurrentCmd,
				EventCounter: 1,
				LastModified: time.Now(),
			}
			if err := queries.InsertRow(ctx, newRow); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		row.EventCounter += 1
		row.LastModified = time.Now()
		if err := queries.UpdateRow(ctx, row); err != nil {
			return err
		}
	}

	return tx.Commit()
}
