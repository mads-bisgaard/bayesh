package bayesh

import (
	"os"
	"path/filepath"
)

const BayeshDirEnvVar = "BAYESH_DIR"

type Settings struct {
	BayeshDir string
	DB        string
}

type FileSystem interface {
	UserHomeDir() (string, error)
	Getenv(key string) string
	MkdirAll(path string, perm os.FileMode) error
	Stat(name string) (os.FileInfo, error)
	Create(name string) (*os.File, error)
}

func CreateSettings(fs FileSystem) (*Settings, error) {
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
	if _, err := fs.Stat(dbPath); os.IsNotExist(err) {
		f, err := fs.Create(dbPath)
		if err != nil {
			return nil, err
		}
		if f != nil {
			if err = f.Close(); err != nil {
				return nil, err
			}
		}
	}
	return &Settings{
		BayeshDir: absDir,
		DB:        dbPath,
	}, nil
}
