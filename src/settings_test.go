package bayesh

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

type mockFileSystem struct {
	homeDir       string
	envVars       map[string]string
	existingPaths map[string]bool
}

func (m mockFileSystem) UserHomeDir() (string, error) {
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
					homeDir:       homeDir,
					envVars:       envVars,
					existingPaths: map[string]bool{},
				}

				settings, err := CreateSettings(mockFS)
				if err != nil {
					t.Fatalf("CreateSettings failed: %v", err)
				}
				if settings.BayeshDir != bayeshDir {
					t.Errorf("Expected BayeshDir %q, got %q", bayeshDir, settings.BayeshDir)
				}
			},
		)
	}
}
