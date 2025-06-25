package bayesh

import (
	"os"
	"path/filepath"
	"testing"
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
	if _, exists := m.existingPaths[path]; !exists {
		m.existingPaths[path] = true
	}
	return nil
}
func (m mockFileSystem) Stat(name string) (os.FileInfo, error) {
	if _, exists := m.existingPaths[name]; exists {
		return nil, nil
	}
	return nil, os.ErrNotExist
}
func (m mockFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}

func TestBayeshDir(t *testing.T) {
	tests := []struct {
		name string
		fs   mockFileSystem
	}{
		{
			name: "bayesh dir exists",
			fs: mockFileSystem{
				homeDir:       "/home/user",
				envVars:       map[string]string{},
				existingPaths: map[string]bool{"/home/user/.bayesh": true},
			},
		},
		{
			name: "bayesh dir does not exist (should be created)",
			fs: mockFileSystem{
				homeDir:       "/home/user",
				envVars:       map[string]string{},
				existingPaths: map[string]bool{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {

			expectedDir := filepath.Join(tc.fs.homeDir, ".bayesh")
			expectedDB := filepath.Join(expectedDir, "bayesh.db")

			settings, err := CreateSettings(tc.fs)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if settings.BayeshDir != expectedDir {
				t.Errorf("expected BayeshDir to be %s, got %s", expectedDir, settings.BayeshDir)
			}
			if settings.DB != expectedDB {
				t.Errorf("expected DB to be %s, got %s", expectedDB, settings.DB)
			}
		})
	}
}
