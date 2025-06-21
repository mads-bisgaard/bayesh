package src

import (
	"bufio"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
)

// Helper for tests: a FileSystem implementation using a function

type mockFileSystem struct {
	existsFunc func(string) bool
}

func (m mockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.existsFunc != nil && m.existsFunc(name) {
		return nil, nil
	}
	return nil, os.ErrNotExist
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

		fs := mockFileSystem{
			existsFunc: func(path string) bool {
				for _, p := range testCase.RequiredPaths {
					if p == path {
						return true
					}
				}
				return false
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
