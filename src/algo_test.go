package bayesh

import (
	"context"
	"math"
	"sync"
	"testing"
)

const testErrorMargin = 1e-9

type testDatabaseQuerer struct {
	eventCounts map[string]int
}

func (t *testDatabaseQuerer) ConditionalEventCounts(ctx context.Context, cwd *string, previousCmd *string, minRequiredEvents *int) (map[string]int, error) {
	return t.eventCounts, nil
}

func TestAddConditionalProbabilities(t *testing.T) {
	settings := &Settings{
		MinRequiredEvents: 1,
	}
	queries := &testDatabaseQuerer{
		eventCounts: map[string]int{
			"cmd1": 3,
			"cmd2": 1,
		},
	}
	result := make(map[string]float64)
	mu := sync.Mutex{}
	errCh := make(chan error, 1)

	addConditionalProbabilities(context.Background(), settings, queries, errCh, result, &mu, nil, nil)

	if err := <-errCh; err != nil {
		t.Fatalf("addConditionalProbabilities returned error: %v", err)
	}

	expectedCmd1Prob := probabilityWeight * (3.0 / 4.0)
	expectedCmd2Prob := probabilityWeight * (1.0 / 4.0)

	if prob, ok := result["cmd1"]; !ok || math.Abs(prob-expectedCmd1Prob) > testErrorMargin {
		t.Errorf("Expected cmd1 probability %v, got %v", expectedCmd1Prob, prob)
	}
	if prob, ok := result["cmd2"]; !ok || math.Abs(prob-expectedCmd2Prob) > testErrorMargin {
		t.Errorf("Expected cmd2 probability %v, got %v", expectedCmd2Prob, prob)
	}
}
