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

	"github.com/asaskevich/govalidator"
)

const BayeshDirEnvVar = "BAYESH_DIR"
const LogLevelEnvVar = "BAYESH_LOG_LEVEL"

type Settings struct {
	BayeshDir string     `json:"BAYESH_DIR" validate:"required,dir"`
	LogLevel  slog.Level `json:"BAYESH_LOG_LEVEL" validate:"required"`
}

func (s *Settings) Db() string {
	return filepath.Join(s.BayeshDir, "bayesh.db")
}

func (s *Settings) setupDatabase() error {
	dbPath := s.Db()
	db, err := sql.Open("sqlite3", dbPath)
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

func Setup(context context.Context, fs FileSystem) (*Settings, error) {
	logLevelStr := fs.Getenv(LogLevelEnvVar)
	if logLevelStr == "" {
		logLevelStr = "ERROR"
	}
	var logLevel slog.Level
	err := logLevel.UnmarshalText([]byte(logLevelStr))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Invalid log level %q: %v\n", logLevelStr, err)
		return nil, err
	}

	bayeshDir := fs.Getenv(BayeshDirEnvVar)
	if bayeshDir == "" {
		home, err := fs.UserHomeDir()
		if err != nil {
			return nil, err
		}
		bayeshDir = filepath.Join(home, ".bayesh")
	}
	if err := fs.MkdirAll(bayeshDir, 0o755); err != nil {
		return nil, err
	}
	settings := Settings{
		BayeshDir: bayeshDir,
		LogLevel:  logLevel,
	}

	result, err := govalidator.ValidateStruct(&settings)
	if err != nil {
		return nil, err
	}
	if !result {
		return nil, errors.New("Could not validate settings: " + govalidator.ToString(&settings))
	}

	dbPath := settings.Db()
	if _, err := fs.Stat(dbPath); os.IsNotExist(err) {
		err := settings.setupDatabase()
		if err != nil {
			return nil, err
		}
	}
	return &settings, nil
}
