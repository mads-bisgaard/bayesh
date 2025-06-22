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

func CreateSettings() (*Settings, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	bayeshDir := os.Getenv(BayeshDirEnvVar)
	if bayeshDir == "" {
		bayeshDir = filepath.Join(home, ".bayesh")
	}
	absDir, err := filepath.Abs(bayeshDir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(absDir, "bayesh.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		f, err := os.Create(dbPath)
		if err != nil {
			return nil, err
		}
		err = f.Close()
		if err != nil {
			return nil, err
		}
	}
	return &Settings{
		BayeshDir: absDir,
		DB:        dbPath,
	}, nil
}
