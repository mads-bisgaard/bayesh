package bayesh

import (
	"fmt"
	"slices"
	"testing"
)

func TestRankCommands(t *testing.T) {
	events := []Row{
		{Cwd: "/a", PreviousCmd: "git status", CurrentCmd: "git add .", EventCounter: 3},
		{Cwd: "/a", PreviousCmd: "git status", CurrentCmd: "git diff", EventCounter: 1},
		{Cwd: "/b", PreviousCmd: "ls", CurrentCmd: "git diff", EventCounter: 4},
	}
	// P(add|/a)=3/4, P(add|status)=3/4, P(add|/a,status)=3/4 => 3/4
	// P(diff|/a)=1/4, P(diff|status)=1/4, P(diff|/a,status)=1/4 => 1/4
	got := RankCommands(events, "/a", "git status", 1)
	if want := []string{"git add .", "git diff"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
	// In /b after ls only "git diff" has been seen
	got = RankCommands(events, "/b", "ls", 1)
	if want := []string{"git diff"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRankCommandsMinRequiredEvents(t *testing.T) {
	events := []Row{
		{Cwd: "/a", PreviousCmd: "", CurrentCmd: "often", EventCounter: 2},
		{Cwd: "/a", PreviousCmd: "", CurrentCmd: "once", EventCounter: 1},
	}
	got := RankCommands(events, "/a", "", 2)
	if want := []string{"often"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRankCommandsTemplates(t *testing.T) {
	events := []Row{
		{Cwd: "/a", PreviousCmd: "git checkout x", CurrentCmd: "vim a.go", EventCounter: 1},
		{Cwd: "/a", PreviousCmd: "git checkout y", CurrentCmd: "vim b.go", EventCounter: 1},
		{Cwd: "/a", PreviousCmd: "git checkout z", CurrentCmd: "vim c.go", EventCounter: 1},
		{Cwd: "/a", PreviousCmd: "vim c.go", CurrentCmd: "git checkout x", EventCounter: 1},
		{Cwd: "/a", PreviousCmd: "vim a.go", CurrentCmd: "git checkout y", EventCounter: 1},
		{Cwd: "/a", PreviousCmd: "vim b.go", CurrentCmd: "git checkout z", EventCounter: 1},
	}
	// The varying arguments are folded, so the counts of each template add up,
	// and a never seen previous command still matches its template.
	got := RankCommands(events, "/elsewhere", "git checkout new", 2)
	if want := []string{"vim <ARG>"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRankCommandsEmpty(t *testing.T) {
	if got := RankCommands(nil, "/a", "ls", 0); len(got) != 0 {
		t.Errorf("expected no commands, got %q", got)
	}
}

func BenchmarkRankCommands(b *testing.B) {
	// 20000 events of 10000 distinct commands, of which a quarter needs the shell parser
	cmd := func(i int) string {
		switch i % 4 {
		case 0:
			return fmt.Sprintf("git checkout branch-%d", i%2500)
		case 1:
			return fmt.Sprintf("vim src/file%d.go", i%2500)
		case 2:
			return fmt.Sprintf("make target-%d", i%2500)
		}
		return fmt.Sprintf(`git commit -m "message %d" && git push`, i%2500)
	}
	var events []Row
	for i := range 20000 {
		events = append(events, Row{
			Cwd:          fmt.Sprintf("/project/%d", i%50),
			PreviousCmd:  cmd(i + 1),
			CurrentCmd:   cmd(i),
			EventCounter: 1,
		})
	}
	for b.Loop() {
		RankCommands(events, "/project/1", "git checkout branch-1", 2)
	}
}
