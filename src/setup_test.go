package bayesh

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type mockFileSystem struct {
	homeDir    string
	homeDirErr error
	envVars    map[string]string
}

func (m mockFileSystem) UserHomeDir() (string, error) {
	if m.homeDirErr != nil {
		return "", m.homeDirErr
	}
	return m.homeDir, nil
}
func (m mockFileSystem) Getenv(key string) string {
	if value, exists := m.envVars[key]; exists {
		return value
	}
	return ""
}
func (m mockFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}
func (m mockFileSystem) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
func (m mockFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func TestBayeshDir(t *testing.T) {
	tests := []struct {
		bayeshDirExists bool
		setEnvVar       bool
	}{

		{
			bayeshDirExists: false,
			setEnvVar:       false,
		},
		{
			bayeshDirExists: true,
			setEnvVar:       false,
		},
		{
			bayeshDirExists: false,
			setEnvVar:       true,
		},
		{
			bayeshDirExists: true,
			setEnvVar:       true,
		},
	}

	for _, tc := range tests {
		t.Run(
			fmt.Sprintf("EnvVar_%v_DirExists_%v", tc.setEnvVar, tc.bayeshDirExists),
			func(t *testing.T) {

				homeDir, err := os.MkdirTemp(os.TempDir(), "")
				if err != nil {
					t.Fatalf("Failed to create temporary home directory: %v", err)
				}
				defer func() {
					if err := os.RemoveAll(homeDir); err != nil {
						t.Errorf("Failed to remove temporary home directory: %v", err)
					}
				}()

				bayeshDir := filepath.Join(homeDir, ".bayesh")
				envVars := map[string]string{}
				if tc.setEnvVar {
					bayeshDir = filepath.Join(homeDir, uuid.NewString())
					envVars[BayeshDirEnvVar] = bayeshDir
				}

				mockFS := mockFileSystem{
					homeDir: homeDir,
					envVars: envVars,
				}
				context := context.Background()
				settings, err := Setup(context, mockFS)
				if err != nil {
					t.Fatalf("CreateSettings failed: %v", err)
				}
				if settings.BayeshDir != bayeshDir {
					t.Errorf("Expected BayeshDir %q, got %q", bayeshDir, settings.BayeshDir)
				}
				if settings.Database != filepath.Join(bayeshDir, "bayesh.db") {
					t.Errorf("Expected DB path %q, got %q", filepath.Join(bayeshDir, "bayesh.db"), settings.Database)
				}
			},
		)
	}
}

func TestBayeshDir_FilePath(t *testing.T) {
	homeDir, err := os.MkdirTemp(os.TempDir(), "")
	if err != nil {
		t.Fatalf("Failed to create temporary home directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(homeDir); err != nil {
			t.Errorf("Failed to remove temporary home directory: %v", err)
		}
	}()

	bayeshDirAsFile := filepath.Join(homeDir, "bayesh_dir_file")
	file, err := os.Create(bayeshDirAsFile)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	defer func() {
		if err := os.Remove(file.Name()); err != nil {
			t.Errorf("Failed to remove file: %v", err)
		}
	}()
	envVars := map[string]string{BayeshDirEnvVar: bayeshDirAsFile}

	mockFS := mockFileSystem{
		homeDir: homeDir,
		envVars: envVars,
	}
	expectedErr := "Could not create directory: '" + bayeshDirAsFile + "'"

	context := context.Background()
	_, err = Setup(context, mockFS)
	if err.Error() != expectedErr {
		t.Fatalf("Expected %s but got %s", expectedErr, err.Error())
	}
}

func TestCreateSettings_UserHomeDirError(t *testing.T) {
	expectedErr := "Could not create directory: ''"
	mockFS := mockFileSystem{
		homeDirErr: errors.New(expectedErr),
	}

	context := context.Background()
	_, err := Setup(context, mockFS)
	if err == nil {
		t.Fatal("Expected an error when UserHomeDir fails, but got nil")
	}

	if err.Error() != expectedErr {
		t.Errorf("Expected error message %q, got %q", expectedErr, err.Error())
	}
}

func TestSettings_ToJSON(t *testing.T) {
	settings := &Settings{
		BayeshDir:           "/path/to/bayesh",
		Database:            "/path/to/bayesh/bayesh.db",
		LogLevel:            slog.LevelError,
		DatabaseTxLock:      "immediate",
		DatabaseBusyTimeout: 5000,
	}

	expectedJSON := `{
  "BAYESH_DIR": "/path/to/bayesh",
  "BAYESH_DATABASE": "/path/to/bayesh/bayesh.db",
  "BAYESH_LOG_LEVEL": "ERROR",
  "BAYESH_DATABASE_TX_LOCK": "immediate",
  "BAYESH_DATABASE_BUSY_TIMEOUT": 5000
}`

	jsonString, err := settings.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if jsonString != expectedJSON {
		t.Errorf("Expected JSON %s, got %s", expectedJSON, jsonString)
	}
}

func TestDataSourceName(t *testing.T) {
	settings := defaultSettings(mockFileSystem{homeDir: "/home/me"})
	want := "/home/me/.bayesh/bayesh.db?_txlock=immediate&_busy_timeout=5000"
	if got := settings.DataSourceName(); got != want {
		t.Errorf("DataSourceName() = %q, want %q", got, want)
	}
}

func TestValidate(t *testing.T) {
	dir := t.TempDir()
	db := filepath.Join(dir, "bayesh.db")
	if err := os.WriteFile(db, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	valid := func() *Settings {
		return &Settings{BayeshDir: dir, Database: db, LogLevel: slog.LevelInfo, DatabaseTxLock: "immediate", DatabaseBusyTimeout: 0}
	}
	if err := valid().validate(OsFs{}); err != nil {
		t.Fatalf("valid settings: %v", err)
	}
	tests := []struct {
		name   string
		modify func(*Settings)
		want   string
	}{
		{"missing dir", func(s *Settings) { s.BayeshDir = filepath.Join(dir, "missing") }, "BAYESH_DIR"},
		{"dir is a file", func(s *Settings) { s.BayeshDir = db }, "BAYESH_DIR"},
		{"missing database", func(s *Settings) { s.Database = filepath.Join(dir, "missing.db") }, "BAYESH_DATABASE"},
		{"database is a dir", func(s *Settings) { s.Database = dir }, "BAYESH_DATABASE"},
		{"unknown tx lock", func(s *Settings) { s.DatabaseTxLock = "bogus" }, "BAYESH_DATABASE_TX_LOCK"},
		{"negative busy timeout", func(s *Settings) { s.DatabaseBusyTimeout = -1 }, "BAYESH_DATABASE_BUSY_TIMEOUT"},
	}
	for _, tt := range tests {
		s := valid()
		tt.modify(s)
		if err := s.validate(OsFs{}); err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: got %v, want an error about %s", tt.name, err, tt.want)
		}
	}
}

func TestSetupAcceptsLogLevelInfo(t *testing.T) {
	mockFS := mockFileSystem{homeDir: t.TempDir(), envVars: map[string]string{LogLevelEnvVar: "INFO"}}
	if _, err := Setup(context.Background(), mockFS); err != nil {
		t.Fatalf("Setup with log level INFO: %v", err)
	}
}
