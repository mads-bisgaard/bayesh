package bayesh

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// benchHistory returns n generated commands. About half of them are distinct,
// and one in seven needs the shell parser.
func benchHistory(n int) []string {
	rng := rand.New(rand.NewSource(1))
	values := max(1, n/10)
	cmds := make([]string, n)
	for i := range cmds {
		v := rng.Intn(values)
		switch i % 7 {
		case 0:
			cmds[i] = fmt.Sprintf("git checkout branch-%d", v)
		case 1:
			cmds[i] = fmt.Sprintf("vim src/file%d.go", v)
		case 2:
			cmds[i] = fmt.Sprintf("cd dir%d/sub", v)
		case 3:
			cmds[i] = fmt.Sprintf("make target-%d", v)
		case 4:
			cmds[i] = fmt.Sprintf("ssh host%d", v)
		case 5:
			cmds[i] = fmt.Sprintf("kill %d", v)
		default:
			cmds[i] = fmt.Sprintf(`git commit -m "message %d" && git push`, v)
		}
	}
	return cmds
}

// newBenchCore returns a core whose database holds the events of cmds, spread
// over 50 directories. They are recorded in one transaction to save time.
func newBenchCore(b *testing.B, cmds []string) *Core {
	b.Helper()
	core, _ := setupCoreWithTempDB(b)
	ctx := context.Background()
	tx, err := core.db.BeginTx(ctx, nil)
	if err != nil {
		b.Fatal(err)
	}
	queries := New(core.db).WithTx(tx)
	for i := 1; i < len(cmds); i++ {
		if err := recordEvent(ctx, queries, time.Now(), fmt.Sprintf("/project/%d", i%50), cmds[i-1], cmds[i]); err != nil {
			b.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		b.Fatal(err)
	}
	b.Cleanup(func() { core.Close() })
	return core
}

// BenchmarkInferCommands measures opening Bayesh end to end on a real SQLite
// database. Its time should not grow with the number of events.
func BenchmarkInferCommands(b *testing.B) {
	for _, n := range []int{5_000, 20_000, 100_000} {
		b.Run(fmt.Sprintf("events=%d", n), func(b *testing.B) {
			core := newBenchCore(b, benchHistory(n))
			ctx := context.Background()
			for b.Loop() {
				if _, err := core.InferCommands(ctx, "/project/1", "git checkout branch-new"); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkRecordEvent measures recording a command, which runs in the background
// after every command, in a database which already holds 20 000 events.
func BenchmarkRecordEvent(b *testing.B) {
	core := newBenchCore(b, benchHistory(20_000))
	ctx := context.Background()
	next := benchHistory(100_000)
	i := 0
	for b.Loop() {
		if err := core.RecordEvent(ctx, "/project/1", next[i], next[i+1]); err != nil {
			b.Fatal(err)
		}
		i++
	}
}
