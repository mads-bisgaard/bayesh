package bayesh

import (
	"slices"
	"testing"

	"mvdan.cc/sh/v3/syntax"
)

func TestAnsiColorTokens(t *testing.T) {
	input := "echo <ARG> <ARG>"
	expected := "echo \033[94m<ARG>\033[0m \033[94m<ARG>\033[0m"
	if result := AnsiColorTokens(input); result != expected {
		t.Errorf("AnsiColorTokens failed: got %q, want %q", result, expected)
	}
}

func TestTemplate(t *testing.T) {
	tests := []struct {
		name    string
		history map[string]int // command -> number of runs
		cmd     string
		want    string
	}{
		{"plain command", nil, "git status", "git status"},
		{"quoted string with spaces", nil, `git commit -m "fix the bug"`, `git commit -m "<ARG>"`},
		{"partly quoted string", nil, `run --msg="a b"`, "run <ARG>"},
		{"quoted string without spaces", nil, `echo "hello"`, `echo "hello"`},
		{"string in command substitution", nil, `log=$(git commit -m 'a message' ; git push)`, "log=$(git commit -m '<ARG>' ; git push)"},
		{"string containing a command", nil, `echo "$(date) is now"`, `echo "<ARG>"`},
		{"unparsable command", nil, "cat 'a b' > (x", "cat 'a b' > (x"},
		{"slot with few values", map[string]int{"cd a": 1, "cd b": 1}, "cd c", "cd c"},
		{"variable slot", map[string]int{"cd a": 1, "cd b": 1, "cd c": 1}, "cd d", "cd <ARG>"},
		{"seen value in variable slot", map[string]int{"cd a": 1, "cd b": 1, "cd c": 1}, "cd a", "cd <ARG>"},
		{"frequent value stays literal", map[string]int{"make build": 6, "make x": 1, "make y": 1}, "make build", "make build"},
		{"less frequent value is folded", map[string]int{"make build": 5, "make x": 1, "make y": 1}, "make build", "make <ARG>"},
		{"rare value is folded", map[string]int{"make build": 6, "make x": 1, "make y": 1}, "make z", "make <ARG>"},
		{"value contained in other words", map[string]int{"ls a": 1, "ls b": 1, "ls c": 1}, "ls s", "ls <ARG>"},
		{"slots are per position", map[string]int{"cp a x": 1, "cp b x": 1, "cp c x": 1}, "cp d x", "cp <ARG> x"},
		{"flags are never folded", map[string]int{"ls -a": 1, "ls -b": 1, "ls -c": 1}, "ls -d", "ls -d"},
		{"command names are never folded", map[string]int{"a": 1, "b": 1, "c": 1}, "d", "d"},
		{"flag value after =", map[string]int{"run --id=1": 1, "run --id=2": 1, "run --id=3": 1}, "run --id=4", "run --id=<ARG>"},
		{"variable after =", map[string]int{"make U=1": 1, "make U=2": 1, "make U=3": 1}, "make U=4 all", "make U=<ARG> all"},
		{"assignment", map[string]int{"P=1 serve": 1, "P=2 serve": 1, "P=3 serve": 1}, "P=4 serve", "P=<ARG> serve"},
		{"pipeline", map[string]int{"cat a": 1, "cat b": 1, "cat c": 1}, "cat d | sort", "cat <ARG> | sort"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templater := newTemplater()
			for cmd, n := range tt.history {
				templater.add(cmd, n)
			}
			if got := templater.Template(tt.cmd); got != tt.want {
				t.Errorf("Template(%q) = %q, want %q", tt.cmd, got, tt.want)
			}
		})
	}
}

func TestAddReportsChangedFolding(t *testing.T) {
	templater := newTemplater()
	steps := []struct {
		cmd  string
		want []string
	}{
		{"cd a", nil},
		{"cd b", nil},
		{"cd c", []string{"cd"}}, // cd#1 becomes variable
		{"cd d", nil},
		{"git status && cd e", nil}, // git#1 has one value, cd#1 stays variable
		{"ls", nil},
	}
	for _, s := range steps {
		if got := templater.add(s.cmd, 1); !slices.Equal(got, s.want) {
			t.Errorf("add(%q) = %q, want %q", s.cmd, got, s.want)
		}
	}
	for range minLiteralCount - 2 {
		if got := templater.add("cd a", 1); got != nil {
			t.Fatalf("add(%q) = %q before cd a is frequent", "cd a", got)
		}
	}
	if got := templater.add("cd a", 1); !slices.Equal(got, []string{"cd"}) {
		t.Errorf("add(%q) = %q when it becomes frequent, want [cd]", "cd a", got)
	}
	if got := templater.Template("cd a"); got != "cd a" {
		t.Errorf("Template(%q) = %q after add", "cd a", got)
	}
}

func TestPlainArgsMatchShellArgs(t *testing.T) {
	parser := syntax.NewParser()
	for _, cmd := range []string{"git status", "cd ~/a/b", "ls -la  src", "make build-all", "ssh me@host:22", "cp a.go b,c x+y 50%", "vim æøå.txt"} {
		if !isPlain(cmd) {
			t.Errorf("%q should be plain", cmd)
		}
		if plain, shell := plainArgs(cmd), shellArgs(parser, cmd); !slices.Equal(plain, shell) {
			t.Errorf("%q: plainArgs = %+v, shellArgs = %+v", cmd, plain, shell)
		}
	}
	for _, cmd := range []string{"a | b", "a; b", "a=1 b", "a --b=1", `a "b"`, "a $B", "a *.go", "a\tb", "a # b", "a!"} {
		if isPlain(cmd) {
			t.Errorf("%q should not be plain", cmd)
		}
	}
}
