package bayesh

import (
	"context"
	"errors"
	"math"
	"testing"
)

const testErrorMargin = 1e-9

type testDatabaseQuerer struct {
	eventCounts map[string]int
	err         error
}

func (t *testDatabaseQuerer) ConditionalEventCounts(ctx context.Context, cwd *string, previousCmd *string, minRequiredEvents *int) (map[string]int, error) {
	return t.eventCounts, t.err
}

func TestComputeCommandProbabilities(t *testing.T) {
	settings := &Settings{
		MinRequiredEvents: 1,
	}
	queries := &testDatabaseQuerer{
		eventCounts: map[string]int{
			"cmd1": 3,
			"cmd2": 1,
		},
		err: nil,
	}
	result := make(map[string]float64)

	result, err := ComputeCommandProbabilities(context.Background(), settings, queries, "/home/user/project", "git status")

	if err != nil {
		t.Fatalf("ComputeCommandProbabilities returned error: %v", err)
	}

	expectedCmd1Prob := (3.0 / 4.0)
	expectedCmd2Prob := (1.0 / 4.0)

	if prob, ok := result["cmd1"]; !ok || math.Abs(prob-expectedCmd1Prob) > testErrorMargin {
		t.Errorf("Expected cmd1 probability %v, got %v", expectedCmd1Prob, prob)
	}
	if prob, ok := result["cmd2"]; !ok || math.Abs(prob-expectedCmd2Prob) > testErrorMargin {
		t.Errorf("Expected cmd2 probability %v, got %v", expectedCmd2Prob, prob)
	}
}

func TestComputeCommandProbabilitiesError(t *testing.T) {
	settings := &Settings{
		MinRequiredEvents: 1,
	}
	queries := &testDatabaseQuerer{
		eventCounts: nil,
		err:         errors.New("database error"),
	}
	_, err := ComputeCommandProbabilities(context.Background(), settings, queries, "/home/user/project", "git status")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

}
