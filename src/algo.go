package bayesh

import (
	"context"
	"errors"
	"log/slog"
	"sync"
)

const probabilityWeight float64 = (float64(1.0) / float64(3.0))

type DatabaseQuerer interface {
	ConditionalEventCounts(ctx context.Context, cwd *string, previousCmd *string, minRequiredEvents *int) (map[string]int, error)
}

func addConditionalProbabilities(
	ctx context.Context,
	settings *Settings,
	queries DatabaseQuerer,
	channel chan error,
	result map[string]float64,
	mu *sync.Mutex,
	cwd *string,
	processedPreviousCmd *string,
) {
	slog.Debug("Adding conditional probability", *cwd, *processedPreviousCmd)

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

// Computes a weighted average of conditional probabilities.
// I.e. it computes the average of P(cmd|cwd), P(cmd|previousCmd), and P(cmd|cwd, previousCmd).
// This can be thought of as a Bayesian inference (E(P(cmd|a)) where a is the context).
func ComputeCommandProbabilities(ctx context.Context, settings *Settings, queries DatabaseQuerer, cwd string, processedPreviousCmd string) (map[string]float64, error) {
	var mu sync.Mutex
	var firstErr error

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	inputs := []struct {
		cwd *string
		cmd *string
	}{
		{&cwd, nil},
		{nil, &processedPreviousCmd},
		{&cwd, &processedPreviousCmd},
	}

	errCh := make(chan error, len(inputs))
	result := make(map[string]float64)

	for _, input := range inputs {
		go addConditionalProbabilities(ctx, settings, queries, errCh, result, &mu, input.cwd, input.cmd)
	}

	for range len(inputs) {
		if err := <-errCh; err != nil {
			if firstErr == nil {
				firstErr = err
				cancel() // Cancel other goroutines on first error
			}
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}

	return result, nil
}
