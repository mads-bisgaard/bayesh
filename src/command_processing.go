package src

import (
	"os"
	"strings"

	"github.com/google/shlex"
)

type Token string

const (
	PATH   Token = "<PATH>"
	STRING Token = "<STRING>"
)

func AnsiColorTokens(cmds string) string {
	cmds = strings.ReplaceAll(cmds, string(PATH), "\033[94m"+string(PATH)+"\033[0m")
	cmds = strings.ReplaceAll(cmds, string(STRING), "\033[94m"+string(STRING)+"\033[0m")
	return cmds
}

func ProcessCmd(cmd string) string {
	parts, err := shlex.Split(cmd)
	if err != nil {
		return cmd // fallback: return original if parsing fails
	}
	for i, p := range parts {
		if strings.Count(cmd, p) > 1 {
			continue
		}
		exists := false
		if _, err := os.Stat(p); err == nil {
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
