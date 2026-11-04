package bayesh

import (
	"cmp"
	"slices"
)

const weight float64 = (float64(1.0) / float64(3.0))

// RankCommands ranks command templates by the average of P(cmd|cwd), P(cmd|previousCmd)
// and P(cmd|cwd, previousCmd). This can be thought of as a Bayesian inference
// (E(P(cmd|a)) where a is the context). Commands are compared by their templates,
// which are learned from the recorded events.
func RankCommands(events []Row, cwd string, previousCmd string, minRequiredEvents int) []string {
	runs := map[string]int{}
	for _, e := range events {
		runs[e.CurrentCmd] += e.EventCounter
	}
	templater := newTemplater(runs)
	previous := templater.Template(previousCmd)

	byCwd, byPrevious, byBoth := map[string]int{}, map[string]int{}, map[string]int{}
	for _, e := range events {
		cmd := templater.Template(e.CurrentCmd)
		sameCwd := e.Cwd == cwd
		samePrevious := templater.Template(e.PreviousCmd) == previous
		if sameCwd {
			byCwd[cmd] += e.EventCounter
		}
		if samePrevious {
			byPrevious[cmd] += e.EventCounter
		}
		if sameCwd && samePrevious {
			byBoth[cmd] += e.EventCounter
		}
	}

	scores := map[string]float64{}
	for _, counts := range []map[string]int{byCwd, byPrevious, byBoth} {
		addProbabilities(scores, counts, minRequiredEvents)
	}

	cmds := make([]string, 0, len(scores))
	for cmd := range scores {
		cmds = append(cmds, cmd)
	}
	slices.SortFunc(cmds, func(a, b string) int {
		return cmp.Or(cmp.Compare(scores[b], scores[a]), cmp.Compare(a, b))
	})
	return cmds
}

// addProbabilities adds weight * P(cmd) for the commands which occur at least minRequiredEvents times.
func addProbabilities(scores map[string]float64, counts map[string]int, minRequiredEvents int) {
	total := 0
	for cmd, count := range counts {
		if count < minRequiredEvents {
			delete(counts, cmd)
			continue
		}
		total += count
	}
	for cmd, count := range counts {
		scores[cmd] += weight * float64(count) / float64(total)
	}
}
