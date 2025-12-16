package bayesh

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
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
	context := context.Background()
	_, err = Setup(context, mockFS)
	var pathError *os.PathError
	if !errors.As(err, &pathError) {
		t.Fatalf("Expected a *os.PathError when BAYESH_DIR is a file, but got %T: %v", err, err)
	}
}

func TestCreateSettings_UserHomeDirError(t *testing.T) {
	expectedErr := "home directory not found"
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
		BayeshDir: "/path/to/bayesh",
		LogLevel:  slog.LevelError,
	}

	expectedJSON := `{
  "BAYESH_DIR": "/path/to/bayesh",
  "BAYESH_LOG_LEVEL": "ERROR"
}`

	jsonString, err := settings.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if jsonString != expectedJSON {
		t.Errorf("Expected JSON %s, got %s", expectedJSON, jsonString)
	}
}
