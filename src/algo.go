package bayesh

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

const probabilityWeight float64 = (float64(1.0) / float64(3.0))

func addConditionalProbabilities(
	ctx context.Context,
	settings *Settings,
	queries *Queries,
	channel chan error,
	result map[string]float64,
	mu *sync.Mutex,
	cwd *string,
	processedPreviousCmd *string,
) {

	eventCounts, err := queries.ConditionalEventCounts(ctx, cwd, processedPreviousCmd, &settings.MinRequiredEvents)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Error("Failed to compute conditional events: " + err.Error())
		}
		channel <- err
		return
	}

	totalCount := 0
	for _, count := range eventCounts {
		totalCount += count
	}
	mu.Lock()
	defer mu.Unlock()
	// weight * conditional probability
	for cmd, count := range eventCounts {
		result[cmd] += probabilityWeight * (float64(count) / float64(totalCount))
	}
	channel <- nil
}

// ComputeCommandProbabilities computes a weighted average of conditional probabilities.
// I.e. it computes the average of P(cmd|cwd), P(cmd|previousCmd), and P(cmd|cwd, previousCmd).
// This can be thought of as a Bayesian inference (E(P(cmd|a)) where a is the context).
func ComputeCommandProbabilities(ctx context.Context, settings *Settings, queries *Queries, cwd string, processedPreviousCmd string) (map[string]float64, error) {
	nGoRoutines := 3
	// Create a cancellable context to ensure resources are cleaned up if an error occurs
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, nGoRoutines)
	result := make(map[string]float64)
	var mu sync.Mutex

	go addConditionalProbabilities(ctx, settings, queries, errCh, result, &mu, &cwd, nil)
	go addConditionalProbabilities(ctx, settings, queries, errCh, result, &mu, nil, &processedPreviousCmd)
	go addConditionalProbabilities(ctx, settings, queries, errCh, result, &mu, &cwd, &processedPreviousCmd)

	var firstErr error
	for ii := 0; ii < nGoRoutines; ii++ {
		if err := <-errCh; err != nil {
			if firstErr == nil {
				firstErr = err
				cancel() // Cancel other goroutines immediately to save resources
			}
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return result, nil
}
