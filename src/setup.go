package bayesh

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
)

const BayeshDirEnvVar = "BAYESH_DIR"
const LogLevelEnvVar = "BAYESH_LOG_LEVEL"

type Settings struct {
	BayeshDir string     `json:"BAYESH_DIR"`
	Database  string     `json:"BAYESH_DATABASE"`
	LogLevel  slog.Level `json:"BAYESH_LOG_LEVEL"`
	// When a transaction takes the write lock. WAL mode lets readers run
	// alongside a writer, but writers still take turns. A transaction which
	// starts reading and then writes can't wait for another writer which
	// committed in between, so writers take the lock when they begin.
	DatabaseTxLock string `json:"BAYESH_DATABASE_TX_LOCK"`
	// How long, in milliseconds, to wait for another writer before failing
	DatabaseBusyTimeout int `json:"BAYESH_DATABASE_BUSY_TIMEOUT"`
}

// validate checks that the directory and database exist and that the database options are valid.
func (s *Settings) validate(fs FileSystem) error {
	if info, err := fs.Stat(s.BayeshDir); err != nil || !info.IsDir() {
		return fmt.Errorf("BAYESH_DIR %q is not a directory", s.BayeshDir)
	}
	if info, err := fs.Stat(s.Database); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("BAYESH_DATABASE %q is not a file", s.Database)
	}
	if !slices.Contains([]string{"deferred", "immediate", "exclusive"}, s.DatabaseTxLock) {
		return fmt.Errorf("BAYESH_DATABASE_TX_LOCK %q is not deferred, immediate or exclusive", s.DatabaseTxLock)
	}
	if s.DatabaseBusyTimeout < 0 {
		return fmt.Errorf("BAYESH_DATABASE_BUSY_TIMEOUT %d is negative", s.DatabaseBusyTimeout)
	}
	return nil
}

// DataSourceName returns the name to open the database with, including its options.
func (s *Settings) DataSourceName() string {
	return fmt.Sprintf("%s?_txlock=%s&_busy_timeout=%d", s.Database, s.DatabaseTxLock, s.DatabaseBusyTimeout)
}

func (s *Settings) setupDatabase() error {
	db, err := sql.Open("sqlite3", s.Database)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("Failed to close DB: %v", err)
		}
	}()
	queries := New(db)
	if err := queries.CreateSchema(context.Background()); err != nil {
		return err
	}
	return nil
}

func (s *Settings) ToJSON() (string, error) {
	jsonBytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

type FileSystem interface {
	UserHomeDir() (string, error)
	Getenv(key string) string
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
}

func defaultSettings(fs FileSystem) *Settings {
	var bayeshDir = ""
	var bayeshDb = ""

	userHomeDir, err := fs.UserHomeDir()
	if err == nil {
		bayeshDir = filepath.Join(userHomeDir, ".bayesh")
		bayeshDb = filepath.Join(bayeshDir, "bayesh.db")
	}

	return &Settings{
		BayeshDir:           bayeshDir,
		Database:            bayeshDb,
		LogLevel:            slog.LevelError,
		DatabaseTxLock:      "immediate",
		DatabaseBusyTimeout: 5000,
	}
}

func Setup(context context.Context, fs FileSystem) (*Settings, error) {
	settings := defaultSettings(fs)
	if logLevelStr := fs.Getenv(LogLevelEnvVar); logLevelStr != "" {
		var logLevel slog.Level
		err := logLevel.UnmarshalText([]byte(logLevelStr))
		if err != nil {
			return nil, err
		}
		settings.LogLevel = logLevel
	}

	if bayeshDir := fs.Getenv(BayeshDirEnvVar); bayeshDir != "" {
		settings.BayeshDir = bayeshDir
		settings.Database = filepath.Join(bayeshDir, "bayesh.db")
	}
	if err := fs.MkdirAll(settings.BayeshDir, 0o755); err != nil {
		return nil, errors.New("Could not create directory: '" + settings.BayeshDir + "'")
	}

	dbPath := settings.Database
	if _, err := fs.Stat(dbPath); os.IsNotExist(err) {
		err := settings.setupDatabase()
		if err != nil {
			return nil, err
		}
	}

	if err := settings.validate(fs); err != nil {
		return nil, err
	}
	logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: settings.LogLevel})
	slog.SetDefault(slog.New(logHandler))
	return settings, nil
}
