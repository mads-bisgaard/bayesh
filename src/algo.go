package bayesh

import (
	"cmp"
	"slices"
)

// RankCommands ranks command templates by the average of their probabilities in
// each of counts, which hold the frecency of each template given the cwd, the
// previous command and both. This can be thought of as a Bayesian
// inference (E(P(cmd|a)) where a is the context). Ties are ordered alphabetically.
func RankCommands(counts ...map[string]float64) []string {
	weight := 1 / float64(len(counts))
	scores := map[string]float64{}
	for _, c := range counts {
		total := 0.0
		for _, n := range c {
			total += n
		}
		for cmd, n := range c {
			scores[cmd] += weight * n / total
		}
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
