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
	"strconv"

	"github.com/asaskevich/govalidator"
)

const BayeshDirEnvVar = "BAYESH_DIR"
const LogLevelEnvVar = "BAYESH_LOG_LEVEL"
const MinRequiredEventsEnvVar = "BAYESH_MIN_REQUIRED_EVENTS"

type Settings struct {
	BayeshDir         string     `json:"BAYESH_DIR" validate:"required,dir"`
	Database          string     `json:"BAYESH_DATABASE" validate:"required,file"`
	LogLevel          slog.Level `json:"BAYESH_LOG_LEVEL" validate:"required"`
	MinRequiredEvents int        `json:"BAYESH_MIN_REQUIRED_EVENTS" validate:"required,min=0"`
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
		BayeshDir:         bayeshDir,
		Database:          bayeshDb,
		LogLevel:          slog.LevelError,
		MinRequiredEvents: 1,
	}
}

func Setup(context context.Context, fs FileSystem) (*Settings, error) {
	settings := defaultSettings(fs)
	if logLevelStr := fs.Getenv(LogLevelEnvVar); logLevelStr != "" {
		var logLevel slog.Level
		err := logLevel.UnmarshalText([]byte(logLevelStr))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid log level %q: %v\n", logLevelStr, err)
			return nil, err
		}
		settings.LogLevel = logLevel
	}
	logHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: settings.LogLevel})
	slog.SetDefault(slog.New(logHandler))

	if minRequiredEventsStr := fs.Getenv(MinRequiredEventsEnvVar); minRequiredEventsStr != "" {
		minRequiredEvents, err := strconv.Atoi(minRequiredEventsStr)
		if err != nil {
			slog.Error("Error: Invalid min required events %q: %v\n", minRequiredEventsStr, err)
			return nil, err
		}
		settings.MinRequiredEvents = minRequiredEvents
	}

	if bayeshDir := fs.Getenv(BayeshDirEnvVar); bayeshDir != "" {
		settings.BayeshDir = bayeshDir
	}
	if err := fs.MkdirAll(settings.BayeshDir, 0o755); err != nil {
		return nil, err
	}

	dbPath := settings.Database
	if _, err := fs.Stat(dbPath); os.IsNotExist(err) {
		err := settings.setupDatabase()
		if err != nil {
			return nil, err
		}
	}

	result, err := govalidator.ValidateStruct(settings)
	if err != nil {
		return nil, err
	}
	if !result {
		return nil, errors.New("Could not validate settings: " + govalidator.ToString(&settings))
	}
	return settings, nil
}
