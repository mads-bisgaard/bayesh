package src

import (
	"os"
	"testing"
)

func TestAnsiColorTokens(t *testing.T) {
	input := "echo <PATH> <STRING>"
	expected := "echo \033[94m<PATH>\033[0m \033[94m<STRING>\033[0m"
	result := AnsiColorTokens(input)
	if result != expected {
		t.Errorf("AnsiColorTokens failed: got %q, want %q", result, expected)
	}
}

func TestProcessCmd_PathAndString(t *testing.T) {
	// Create a temp file to simulate a path
	tmpfile, err := os.CreateTemp("", "testfile*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	cmd := "cat " + tmpfile.Name() + " 'some string'"
	expected := "cat <PATH> <STRING>"
	result := ProcessCmd(cmd)
	if result != expected {
		t.Errorf("ProcessCmd failed: got %q, want %q", result, expected)
	}
}

func TestProcessCmd_PermissionError(t *testing.T) {
	// Simulate a permission error by using a path that cannot exist
	cmd := "cat /root/this_file_should_not_exist_1234567890"
	_ = ProcessCmd(cmd) // Should not panic or error
}

func TestProcessCmd_Quotes(t *testing.T) {
	cmd := "echo 'hello world'"
	expected := "echo <STRING>"
	result := ProcessCmd(cmd)
	if result != expected {
		t.Errorf("ProcessCmd with quotes failed: got %q, want %q", result, expected)
	}
}

func TestProcessCmd_MultiplePaths(t *testing.T) {
	// Create two temp files
	tmp1, err1 := os.CreateTemp("", "file1*")
	tmp2, err2 := os.CreateTemp("", "file2*")
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to create temp files")
	}
	defer os.Remove(tmp1.Name())
	defer os.Remove(tmp2.Name())

	cmd := "diff " + tmp1.Name() + " " + tmp2.Name()
	expected := "diff <PATH> <PATH>"
	result := ProcessCmd(cmd)
	if result != expected {
		t.Errorf("ProcessCmd with multiple paths failed: got %q, want %q", result, expected)
	}
}

func TestEndsWithAny(t *testing.T) {
	if !endsWithAny("foo|", []string{"|", ";"}) {
		t.Error("endsWithAny should return true for suffix match")
	}
	if endsWithAny("foo", []string{"|", ";"}) {
		t.Error("endsWithAny should return false for no match")
	}
}
