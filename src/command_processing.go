package bayesh

import (
	"os"
	"strings"

	"github.com/google/shlex"
)

type StatFileSystem interface {
	Stat(name string) (os.FileInfo, error)
}

type Token string

const (
	PATH   Token = "<PATH>"
	STRING Token = "<STRING>"
)

const (
	CYAN  Token = "\033[94m"
	RESET Token = "\033[0m"
)

func AnsiColorTokens(cmds string) string {
	cmds = strings.ReplaceAll(cmds, string(PATH), string(CYAN)+string(PATH)+string(RESET))
	cmds = strings.ReplaceAll(cmds, string(STRING), string(CYAN)+string(STRING)+string(RESET))
	return cmds
}

// Update ProcessCmd to use StatFileSystem
func ProcessCmd(fs StatFileSystem, cmd string) string {
	parts, err := shlex.Split(cmd)
	if err != nil {
		return cmd // fallback: return original if parsing fails
	}
	for i, p := range parts {
		if strings.Count(cmd, p) > 1 {
			continue
		}
		exists := false
		if _, err := fs.Stat(p); err == nil {
			exists = true
		}
		if exists && p != "." && i > 0 && !endsWithAny(parts[i-1], []string{"(", ")", ";", "<", ">", "|", "&"}) {
			cmd = strings.Replace(cmd, p, string(PATH), 1)
		} else if strings.Contains(p, " ") && !exists {
			cmd = strings.Replace(cmd, p, string(STRING), 1)
		}
	}
	return cmd
}

func endsWithAny(s string, suffixes []string) bool {
	for _, suf := range suffixes {
		if strings.HasSuffix(s, suf) {
			return true
		}
	}
	return false
}
