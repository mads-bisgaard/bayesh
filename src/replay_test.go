package bayesh

import (
	"math"
	"os"
	"slices"
	"strings"
	"testing"
)

const replayHistoryEnvVar = "BAYESH_REPLAY_HISTORY"

// TestReplay measures how well Bayesh predicts a shell history. The history is
// replayed in order, and at up to 500 points the next command is looked up in
// the suggestions made from the commands before it. Run it with e.g.
//
//	BAYESH_REPLAY_HISTORY=~/.zsh_history go test ./src -run TestReplay -v
//
// The file holds one command per line, optionally in zsh's extended history format.
// Since a history has no working directories, all commands share one.
func TestReplay(t *testing.T) {
	path := os.Getenv(replayHistoryEnvVar)
	if path == "" {
		t.Skip("set " + replayHistoryEnvVar + " to a shell history file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var cmds []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, ": ") { // ": <start>:<duration>;<command>"
			_, line, _ = strings.Cut(line, ";")
		}
		if line = strings.TrimSpace(line); line != "" {
			cmds = append(cmds, line)
		}
	}
	if len(cmds) < 2 {
		t.Fatal("the history needs at least two commands")
	}

	defer func(slotValues, literalCount int) {
		minSlotValues, minLiteralCount = slotValues, literalCount
	}(minSlotValues, minLiteralCount)

	t.Logf("%d commands", len(cmds))
	t.Logf("%-24s %7s %7s %7s %7s %10s", "setting", "top-1", "top-5", "MRR", "exact-5", "suggested")
	for _, s := range []struct {
		name                     string
		slotValues, literalCount int
	}{
		{"no <ARG> folding", math.MaxInt, 0},
		{"slot 3, literal 2", 3, 2},
		{"slot 2, literal 2", 2, 2},
		{"slot 5, literal 2", 5, 2},
		{"slot 3, literal 3", 3, 3},
		{"slot 3, literal 5", 3, 5},
		{"slot 3, literal 8", 3, 8},
		{"slot 3, literal 10", 3, 10},
	} {
		minSlotValues, minLiteralCount = s.slotValues, s.literalCount
		r := replay(cmds)
		t.Logf("%-24s %6.1f%% %6.1f%% %7.3f %6.1f%% %9.1f%%", s.name, r.top1, r.top5, r.mrr, r.exact5, r.suggested)
	}
}

type replayResult struct {
	top1, top5, mrr, exact5 float64
	suggested               float64 // how often the next command is suggested at all
}

func replay(cmds []string) replayResult {
	events := map[[2]string]int{} // previous and current command -> runs
	step := max(1, len(cmds)/500)
	var points, top1, top5, exact5, suggested int
	var reciprocalRanks float64
	for i := 1; i < len(cmds); i++ {
		previous, current := cmds[i-1], cmds[i]
		if i%step == 0 {
			rows, runs := []Row{}, map[string]int{}
			for cmds, n := range events {
				rows = append(rows, Row{PreviousCmd: cmds[0], CurrentCmd: cmds[1], EventCounter: n})
				runs[cmds[1]] += n
			}
			ranked := RankCommands(rows, "", previous, 2)
			points++
			if rank := slices.Index(ranked, newTemplater(runs).Template(current)); rank >= 0 {
				suggested++
				reciprocalRanks += 1 / float64(rank+1)
				if rank == 0 {
					top1++
				}
				if rank < 5 {
					top5++
				}
			}
			if slices.Contains(ranked[:min(5, len(ranked))], current) {
				exact5++
			}
		}
		events[[2]string{previous, current}]++
	}
	percent := func(n int) float64 { return 100 * float64(n) / float64(points) }
	return replayResult{
		top1:      percent(top1),
		top5:      percent(top5),
		mrr:       reciprocalRanks / float64(points),
		exact5:    percent(exact5),
		suggested: percent(suggested),
	}
}
