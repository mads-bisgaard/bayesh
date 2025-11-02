package bayesh

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

const BayeshDirEnvVar = "BAYESH_DIR"
const LogLevelEnvVar = "BAYESH_LOG_LEVEL"

type Settings struct {
	BayeshDir string     `json:"BAYESH_DIR"`
	DB        string     `json:"BAYESH_DB"`
	LogLevel  slog.Level `json:"BAYESH_LOG_LEVEL"`
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

func Setup(context context.Context, fs FileSystem) (*Settings, error) {
	logLevel := slog.LevelError
	logLevelStr := strings.ToLower(fs.Getenv(LogLevelEnvVar))
	switch logLevelStr {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	home, err := fs.UserHomeDir()
	if err != nil {
		return nil, err
	}
	bayeshDir := fs.Getenv(BayeshDirEnvVar)
	if bayeshDir == "" {
		bayeshDir = filepath.Join(home, ".bayesh")
	}
	absDir, err := filepath.Abs(bayeshDir)
	if err != nil {
		return nil, err
	}
	if err := fs.MkdirAll(absDir, 0o755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(absDir, "bayesh.db")
	// If the database file doesn't exist, create it and set up the schema.
	if _, err := fs.Stat(dbPath); os.IsNotExist(err) {
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := db.Close(); err != nil {
				slog.Error("Failed to close DB:", "error", err)
			}
		}()
		queries := New(db)
		if err := queries.CreateSchema(context); err != nil {
			return nil, err
		}
	}
	return &Settings{
		BayeshDir: absDir,
		DB:        dbPath,
		LogLevel:  logLevel,
	}, nil
}
