package bayesh

import (
	"slices"
	"testing"
)

func TestRankCommands(t *testing.T) {
	byCwd := map[string]int{"git add .": 3, "git diff": 1}
	byPrevious := map[string]int{"git add .": 3, "git diff": 1, "ls": 4}
	byBoth := map[string]int{"git add .": 3, "git diff": 1}
	// add: (3/4 + 3/8 + 3/4)/3, ls: (4/8)/3, diff: (1/4 + 1/8 + 1/4)/3
	got := RankCommands(byCwd, byPrevious, byBoth)
	if want := []string{"git add .", "git diff", "ls"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRankCommandsTies(t *testing.T) {
	got := RankCommands(map[string]int{"b": 1, "c": 1, "a": 1}, nil, nil)
	if want := []string{"a", "b", "c"}; !slices.Equal(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRankCommandsEmpty(t *testing.T) {
	if got := RankCommands(nil, map[string]int{}, nil); len(got) != 0 {
		t.Errorf("expected no commands, got %q", got)
	}
}
