package bayesh

import (
	"slices"
	"strconv"
	"strings"
	"unicode"

	"mvdan.cc/sh/v3/syntax"
)

type Token string

const (
	ARG    Token = "<ARG>"
	STRING Token = "<STRING>"
	PATH   Token = "<PATH>" // only found in databases recorded by older versions
)

const (
	CYAN  Token = "\033[94m"
	RESET Token = "\033[0m"
)

// An argument slot is folded into <ARG> once it has seen minSlotValues
// distinct values. Values run at least minLiteralCount times stay literal.
// They are variables so the replay test (replay_test.go) can compare settings;
// on a real history top-5 and MRR stopped improving at minLiteralCount = 8.
var (
	minSlotValues   = 3
	minLiteralCount = 8
)

func AnsiColorTokens(cmds string) string {
	for _, token := range []Token{ARG, STRING, PATH} {
		cmds = strings.ReplaceAll(cmds, string(token), string(CYAN+token+RESET))
	}
	return cmds
}

// arg is an argument in a command which may be replaced by a token.
type arg struct {
	start, end int    // byte offsets into the command
	slot       string // where the argument occurs, e.g. "git#2"; empty for quoted strings
	value      string
}

// templater replaces the arguments of commands which vary between runs with tokens.
type templater struct {
	parser    *syntax.Parser
	args      map[string][]arg          // command -> its arguments
	values    map[string]map[string]int // slot -> value -> number of runs
	templates map[string]string         // command -> its template
}

// newTemplater learns which argument slots vary from the number of runs of each command.
func newTemplater(runs map[string]int) *templater {
	t := &templater{
		parser:    syntax.NewParser(),
		args:      map[string][]arg{},
		values:    map[string]map[string]int{},
		templates: map[string]string{},
	}
	for cmd, n := range runs {
		for _, a := range t.parse(cmd) {
			if a.slot == "" {
				continue
			}
			if t.values[a.slot] == nil {
				t.values[a.slot] = map[string]int{}
			}
			t.values[a.slot][a.value] += n
		}
	}
	return t
}

func (t *templater) Template(cmd string) string {
	if template, ok := t.templates[cmd]; ok {
		return template
	}
	var b strings.Builder
	last := 0
	for _, a := range t.parse(cmd) {
		if a.start < last || !t.fold(a) {
			continue
		}
		b.WriteString(cmd[last:a.start])
		if a.slot == "" {
			b.WriteString(string(STRING))
		} else {
			b.WriteString(string(ARG))
		}
		last = a.end
	}
	b.WriteString(cmd[last:])
	t.templates[cmd] = b.String()
	return t.templates[cmd]
}

func (t *templater) fold(a arg) bool {
	if a.slot == "" {
		return true
	}
	values := t.values[a.slot]
	return len(values) >= minSlotValues && values[a.value] < minLiteralCount
}

func (t *templater) parse(cmd string) []arg {
	args, ok := t.args[cmd]
	if !ok {
		if isPlain(cmd) {
			args = plainArgs(cmd)
		} else {
			args = shellArgs(t.parser, cmd)
		}
		t.args[cmd] = args
	}
	return args
}

// isPlain reports whether cmd is a single command of words the shell takes literally.
func isPlain(cmd string) bool {
	return !strings.ContainsFunc(cmd, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune(" -_./:@%+,~", r)
	})
}

// plainArgs is a fast version of shellArgs for plain commands.
func plainArgs(cmd string) []arg {
	var args []arg
	name, offset := "", 0
	for i, field := range strings.Fields(cmd) {
		start := offset + strings.Index(cmd[offset:], field)
		offset = start + len(field)
		if i == 0 {
			name = field
		} else if !strings.HasPrefix(field, "-") {
			args = append(args, arg{start: start, end: offset, slot: name + "#" + strconv.Itoa(i), value: field})
		}
	}
	return args
}

// shellArgs returns the arguments of all commands in cmd, ordered by position.
// Command names and flags are not arguments. An unparsable cmd has none.
func shellArgs(parser *syntax.Parser, cmd string) []arg {
	file, err := parser.Parse(strings.NewReader(cmd), "")
	if err != nil {
		return nil
	}
	var args []arg
	add := func(w *syntax.Word, skip int, slot string) {
		if a, ok := newArg(cmd, w, skip, slot); ok {
			args = append(args, a)
		}
	}
	syntax.Walk(file, func(node syntax.Node) bool {
		call, ok := node.(*syntax.CallExpr)
		if !ok {
			return true
		}
		name := ""
		if len(call.Args) > 0 {
			name = source(cmd, call.Args[0])
		}
		for _, assign := range call.Assigns {
			if assign.Name != nil && assign.Value != nil {
				add(assign.Value, 0, name+" "+assign.Name.Value+"=")
			}
		}
		for i := 1; i < len(call.Args); i++ {
			w := call.Args[i]
			if key, _, ok := strings.Cut(w.Lit(), "="); ok && key != "" {
				add(w, len(key)+1, name+" "+key+"=") // e.g. --name=value
			} else {
				add(w, 0, name+"#"+strconv.Itoa(i))
			}
		}
		return true
	})
	slices.SortFunc(args, func(a, b arg) int { return a.start - b.start })
	return args
}

// newArg returns w, without its first skip bytes, as an argument occurring in slot.
func newArg(cmd string, w *syntax.Word, skip int, slot string) (arg, bool) {
	start, end := int(w.Pos().Offset())+skip, int(w.End().Offset())
	text := cmd[start:end]
	switch {
	case isQuoted(w) && strings.ContainsAny(text, " \t\n"):
		if q := text[0]; (q == '"' || q == '\'') && text[len(text)-1] == q {
			start, end = start+1, end-1 // keep the quotes, so the cursor lands between them
		}
		return arg{start: start, end: end}, true
	case w.Lit() == "" || text == "" || (skip == 0 && strings.HasPrefix(text, "-")):
		return arg{}, false
	}
	return arg{start: start, end: end, slot: slot, value: text}, true
}

func isQuoted(w *syntax.Word) bool {
	return slices.ContainsFunc(w.Parts, func(p syntax.WordPart) bool {
		switch p.(type) {
		case *syntax.SglQuoted, *syntax.DblQuoted:
			return true
		}
		return false
	})
}

func source(cmd string, node syntax.Node) string {
	return cmd[node.Pos().Offset():node.End().Offset()]
}
