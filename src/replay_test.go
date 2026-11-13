package bayesh

import (
	"context"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"testing"
	"time"
)

const replayHistoryEnvVar = "BAYESH_REPLAY_HISTORY"

// TestReplay measures how well Bayesh predicts a shell history. The history is
// replayed in order, and at up to 500 points the next command is looked up in
// the suggestions made from the commands before it. Run it with e.g.
//
//	BAYESH_REPLAY_HISTORY=~/.zsh_history go test ./src -run TestReplay -v
//
// The file holds one command per line, optionally in zsh's extended history format.
// Since a history has no working directories, all commands share one, and
// since it may have no times, all are recorded at the same time.
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

	defer func(slotValues, literalCount, literalPercent int) {
		minSlotValues, minLiteralCount, minLiteralPercent = slotValues, literalCount, literalPercent
	}(minSlotValues, minLiteralCount, minLiteralPercent)

	t.Logf("%d commands", len(cmds))
	t.Logf("%-28s %7s %7s %7s %7s %10s", "setting", "top-1", "top-5", "MRR", "exact-5", "suggested")
	for _, s := range []struct {
		slotValues, literalCount, literalPercent int
	}{
		{math.MaxInt, 0, 0}, // no <ARG> folding
		{3, 6, 0},
		{3, 8, 0},
		{3, 0, 15},
		{3, 2, 10},
		{3, 4, 10},
		{3, 6, 10},
		{3, 4, 15},
		{3, 6, 15},
		{3, 8, 15},
		{3, 6, 20},
	} {
		name := "no <ARG> folding"
		if s.slotValues != math.MaxInt {
			name = fmt.Sprintf("slot %d, literal %d + %d%%", s.slotValues, s.literalCount, s.literalPercent)
		}
		minSlotValues, minLiteralCount, minLiteralPercent = s.slotValues, s.literalCount, s.literalPercent
		r := replay(t, cmds)
		t.Logf("%-28s %6.1f%% %6.1f%% %7.3f %6.1f%% %9.1f%%", name, r.top1, r.top5, r.mrr, r.exact5, r.suggested)
	}
}

type replayResult struct {
	top1, top5, mrr, exact5 float64
	suggested               float64 // how often the next command is suggested at all
}

// replay records cmds in a new database and ranks the suggestions at up to 500 points.
func replay(t *testing.T, cmds []string) replayResult {
	core, _ := setupCoreWithTempDB(t)
	// The history's times are not used: all events are recorded and ranked at
	// the same time, so they have the same frecency weight.
	now := time.Now()
	core.now = func() time.Time { return now }
	// A single connection, so the pragma applies to all queries. The database is thrown away.
	core.db.SetMaxOpenConns(1)
	if _, err := core.db.Exec("PRAGMA synchronous = OFF"); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	step := max(1, len(cmds)/500)
	var points, top1, top5, exact5, suggested int
	var reciprocalRanks float64
	for i := 1; i < len(cmds); i++ {
		previous, current := cmds[i-1], cmds[i]
		if i%step == 0 {
			ranked, err := core.InferCommands(ctx, "", previous)
			if err != nil {
				t.Fatal(err)
			}
			target, err := core.template(ctx, current)
			if err != nil {
				t.Fatal(err)
			}
			points++
			if rank := slices.Index(ranked, target); rank >= 0 {
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
		if err := core.RecordEvent(ctx, "", previous, current); err != nil {
			t.Fatal(err)
		}
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
