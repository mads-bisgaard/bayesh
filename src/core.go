package bayesh

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

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
	db, err := sql.Open("sqlite3", settings.Db())
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

func (c *Core) InferCommands(ctx context.Context, cwd string, previousCmd string) ([]string, error) {
	queries := New(c.db)
	processedPreviousCmd := ProcessCmd(OsFs{}, previousCmd)
	inferredCmds, err := queries.InferCurrentCmd(ctx, cwd, processedPreviousCmd)
	if err != nil {
		return nil, err
	}

	return inferredCmds, nil
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
