package bayesh

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// Helper for tests: a FileSystem implementation using a function

type mockStatFileSystem struct {
	statFunc func(string) (os.FileInfo, error)
}

func (m mockStatFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.statFunc == nil {
		return nil, os.ErrInvalid
	}
	return m.statFunc(name)
}

func TestNoPermission(t *testing.T) {
	fs := mockStatFileSystem{
		statFunc: func(path string) (os.FileInfo, error) {
			return nil, os.ErrPermission
		},
	}
	ProcessCmd(fs, "echo ./myfile.txt")
}

func TestAnsiColorTokens(t *testing.T) {
	input := "echo <PATH> <STRING>"
	expected := "echo \033[94m<PATH>\033[0m \033[94m<STRING>\033[0m"
	result := AnsiColorTokens(input)
	if result != expected {
		t.Errorf("AnsiColorTokens failed: got %q, want %q", result, expected)
	}
}

func TestProcessCmd_Parametrized(t *testing.T) {
	dataFile := filepath.Join("..", "tests", "data", "processed_bash_commands")
	f, err := os.Open(dataFile)
	if err != nil {
		t.Fatalf("Failed to open test data file: %v", err)
	}
	defer f.Close()

	type CommandPairTestData struct {
		RawCmd        string   `json:"raw_cmd"`
		SanitizedCmd  string   `json:"sanitized_cmd"`
		RequiredPaths []string `json:"required_paths"`
	}

	scanner := bufio.NewScanner(f)
	idx := 0
	for scanner.Scan() {
		var testCase CommandPairTestData
		if err := json.Unmarshal([]byte(scanner.Text()), &testCase); err != nil {
			t.Errorf("Failed to parse JSON on line %d: %v", idx+1, err)
			continue
		}

		fs := mockStatFileSystem{
			statFunc: func(path string) (os.FileInfo, error) {
				for _, p := range testCase.RequiredPaths {
					if p == path {
						return os.FileInfo(nil), nil
					}
				}
				return nil, os.ErrNotExist
			},
		}

		t.Run(testCase.RawCmd[:int(math.Min(10, float64(len(testCase.RawCmd))))], func(t *testing.T) {
			result := ProcessCmd(fs, testCase.RawCmd)
			if result != testCase.SanitizedCmd {
				t.Errorf("Raw: %q\nGot: %q\nWant: %q", testCase.RawCmd, result, testCase.SanitizedCmd)
			}
		})
		idx++
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading test data file: %v", err)
	}
}
