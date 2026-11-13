package bayesh

import (
	"math/rand"
	"slices"
	"strconv"
	"strings"
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

func TestFormatSuggestions(t *testing.T) {
	tests := []struct {
		name     string
		cmds     []string
		sep      string
		expected string
	}{
		{"none", nil, "\x00", ""},
		{"each is terminated", []string{"ls", "cd <ARG>"}, "\n", "ls\ncd \033[94m<ARG>\033[0m\n"},
		{"each is NUL-terminated", []string{"ls", "cd <ARG>"}, "\x00", "ls\x00cd \033[94m<ARG>\033[0m\x00"},
		{"newlines are kept", []string{"for f in *\ndo echo $f\ndone"}, "\x00", "for f in *\ndo echo $f\ndone\x00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result := FormatSuggestions(tt.cmds, tt.sep); result != tt.expected {
				t.Errorf("got %q, want %q", result, tt.expected)
			}
		})
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
		// A value stays literal with at least 6 runs plus 10% of its slot's runs
		{"frequent value stays literal", map[string]int{"make build": 7, "make x": 1, "make y": 1}, "make build", "make build"},  // 7 >= 6 + 10% of 9
		{"less frequent value is folded", map[string]int{"make build": 6, "make x": 1, "make y": 1}, "make build", "make <ARG>"}, // 6 < 6 + 10% of 8
		{"rare value is folded", map[string]int{"make build": 8, "make x": 1, "make y": 1}, "make z", "make <ARG>"},
		{"value with a small share is folded", withValues(map[string]int{"git add a": 9}, "git add f", 30), "git add a", "git add <ARG>"},  // 9 < 6 + 10% of 39
		{"value with a large share stays literal", withValues(map[string]int{"git add a": 10}, "git add f", 30), "git add a", "git add a"}, // 10 >= 6 + 10% of 40
		{"value contained in other words", map[string]int{"ls a": 1, "ls b": 1, "ls c": 1}, "ls s", "ls <ARG>"},
		{"slots are per position", map[string]int{"cp a x": 1, "cp b x": 1, "cp c x": 1}, "cp d x", "cp <ARG> x"},
		{"flags are never folded", map[string]int{"ls -a": 1, "ls -b": 1, "ls -c": 1}, "ls -d", "ls -d"},
		{"command names are never folded", map[string]int{"a": 1, "b": 1, "c": 1}, "d", "d"},
		{"flag value after =", map[string]int{"run --id=1": 1, "run --id=2": 1, "run --id=3": 1}, "run --id=4", "run --id=<ARG>"},
		{"variable after =", map[string]int{"make U=1": 1, "make U=2": 1, "make U=3": 1}, "make U=4 all", "make U=<ARG> all"},
		{"assignment", map[string]int{"P=1 serve": 1, "P=2 serve": 1, "P=3 serve": 1}, "P=4 serve", "P=<ARG> serve"},
		{"pipeline", map[string]int{"cat a": 1, "cat b": 1, "cat c": 1}, "cat d | sort", "cat <ARG> | sort"},
		{"multi-line command", map[string]int{"cat a": 1, "cat b": 1, "cat c": 1}, "for f in x\ndo\n  cat d\ndone", "for f in x\ndo\n  cat <ARG>\ndone"},
		{"multi-line string", nil, "echo 'a\nb'", "echo '<ARG>'"},
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

// withValues adds runs of n distinct commands prefix0, prefix1, ... to history.
func withValues(history map[string]int, prefix string, n int) map[string]int {
	for i := range n {
		history[prefix+strconv.Itoa(i)] = 1
	}
	return history
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
	// cd a needs 6 runs plus 10% of the runs of cd#1, which has 5 runs
	for range 6 {
		if got := templater.add("cd a", 1); got != nil {
			t.Fatalf("add(%q) = %q before cd a is frequent", "cd a", got)
		}
	}
	if got := templater.add("cd a", 1); !slices.Equal(got, []string{"cd"}) { // 8 >= 6 + 10% of 12
		t.Errorf("add(%q) = %q when it becomes frequent, want [cd]", "cd a", got)
	}
	if got := templater.Template("cd a"); got != "cd a" {
		t.Errorf("Template(%q) = %q after add", "cd a", got)
	}
	// Runs of other values lower the share of cd a until it is folded again
	for i := range 8 { // 8 >= 6 + 10% of 20
		if got := templater.add("cd b"+strconv.Itoa(i), 1); got != nil {
			t.Fatalf("add(%q) = %q while cd a is frequent", "cd b"+strconv.Itoa(i), got)
		}
	}
	if got := templater.add("cd b8", 1); !slices.Equal(got, []string{"cd"}) { // 8 < 6 + 10% of 21
		t.Errorf("add(%q) = %q when cd a becomes rare, want [cd]", "cd b8", got)
	}
	if got := templater.Template("cd a"); got != "cd <ARG>" {
		t.Errorf("Template(%q) = %q after other values were added", "cd a", got)
	}
}

// TestAddReportsAllChangedFolding checks on random commands that add reports
// exactly the names of the commands whose templates changed.
func TestAddReportsAllChangedFolding(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	templater := newTemplater()
	var seen []string
	for range 2000 {
		cmd := []string{"cd ", "make ", "git checkout "}[rng.Intn(3)] + strconv.Itoa(int(rng.ExpFloat64()*3))
		before := map[string]string{}
		for _, c := range seen {
			before[c] = templater.Template(c)
		}
		got := templater.add(cmd, 1)
		var want []string
		for _, c := range seen {
			if name, _, _ := strings.Cut(c, " "); templater.Template(c) != before[c] && !slices.Contains(want, name) {
				want = append(want, name)
			}
		}
		if !slices.Contains(seen, cmd) {
			seen = append(seen, cmd)
		}
		slices.Sort(got)
		slices.Sort(want)
		if !slices.Equal(got, want) {
			t.Fatalf("add(%q) = %q, want %q", cmd, got, want)
		}
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
