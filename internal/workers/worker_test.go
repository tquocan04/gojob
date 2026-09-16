package workers

import (
	"context"
	"testing"
	"time"
)

func TestRetryDelay(t *testing.T) {
	tests := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{name: "non-positive attempt falls back to 1s", attempt: 0, want: 1 * time.Second},
		{name: "negative attempt falls back to 1s", attempt: -3, want: 1 * time.Second},
		{name: "first attempt", attempt: 1, want: 1 * time.Second},
		{name: "second attempt", attempt: 2, want: 2 * time.Second},
		{name: "third attempt", attempt: 3, want: 4 * time.Second},
		{name: "fifth attempt", attempt: 5, want: 16 * time.Second}, // 2^4
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := retryDelay(tt.attempt)
			if got != tt.want {
				t.Fatalf("retryDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestRetryDelayGrowsExponentially(t *testing.T) {
	prev := retryDelay(1)
	for attempt := 2; attempt <= 5; attempt++ {
		got := retryDelay(attempt)
		if got != prev*2 {
			t.Fatalf("retryDelay(%d) = %v, want %v (2x of previous)", attempt, got, prev*2)
		}
		prev = got
	}
}

func TestSleepReturnsEarlyOnCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	start := time.Now()
	err := sleep(ctx, time.Hour)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("sleep with cancelled context should return an error")
	}
	if elapsed > time.Second {
		t.Fatalf("sleep did not return early on cancellation, took %v", elapsed)
	}
}

func TestSleepWaitsForDuration(t *testing.T) {
	start := time.Now()
	err := sleep(context.Background(), 20*time.Millisecond)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("sleep returned unexpected error: %v", err)
	}
	if elapsed < 20*time.Millisecond {
		t.Fatalf("sleep returned before duration elapsed, took %v", elapsed)
	}
}
