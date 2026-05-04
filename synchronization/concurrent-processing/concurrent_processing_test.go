package concurrent_processing

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func TestExecuteConcurrent(t *testing.T) {
	t.Parallel()

	t.Run("happy path - all items processed exactly once", func(t *testing.T) {
		t.Parallel()

		const n = 100
		items := makeIntRange(n)

		var mu sync.Mutex
		seen := make([]int, 0, n)

		err := ExecuteConcurrent(context.Background(), items, 10, func(_ context.Context, v int) error {
			mu.Lock()
			seen = append(seen, v)
			mu.Unlock()
			return nil
		})

		require.NoError(t, err)
		assert.ElementsMatch(t, items, seen)
	})

	t.Run("concurrency limit is respected", func(t *testing.T) {
		t.Parallel()

		const limit = 8
		var inFlight, peak atomic.Int32

		err := ExecuteConcurrent(context.Background(), makeIntRange(64), limit, func(_ context.Context, _ int) error {
			n := inFlight.Add(1)
			updateMax(&peak, n)
			time.Sleep(2 * time.Millisecond)
			inFlight.Add(-1)
			return nil
		})

		require.NoError(t, err)
		assert.LessOrEqual(t, peak.Load(), int32(limit), "max in-flight must not exceed limit")
		assert.Equal(t, int32(limit), peak.Load(), "limit should actually be hit when items >> limit")
	})

	t.Run("limit one is sequential", func(t *testing.T) {
		t.Parallel()

		var inFlight, peak atomic.Int32

		err := ExecuteConcurrent(context.Background(), makeIntRange(20), 1, func(_ context.Context, _ int) error {
			n := inFlight.Add(1)
			updateMax(&peak, n)
			time.Sleep(time.Millisecond)
			inFlight.Add(-1)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, int32(1), peak.Load())
	})

	t.Run("limit greater than n - all items still run", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32

		err := ExecuteConcurrent(context.Background(), makeIntRange(5), 100, func(_ context.Context, _ int) error {
			calls.Add(1)
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, int32(5), calls.Load())
	})

	t.Run("empty input returns nil and never invokes f", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32

		err := ExecuteConcurrent(context.Background(), []int{}, 10, func(_ context.Context, _ int) error {
			calls.Add(1)
			return nil
		})

		require.NoError(t, err)
		assert.Zero(t, calls.Load())
	})

	t.Run("nil input returns nil and never invokes f", func(t *testing.T) {
		t.Parallel()

		var calls atomic.Int32

		err := ExecuteConcurrent(context.Background(), []int(nil), 10, func(_ context.Context, _ int) error {
			calls.Add(1)
			return nil
		})

		require.NoError(t, err)
		assert.Zero(t, calls.Load())
	})

	t.Run("already-cancelled ctx returns ctx error and never invokes f", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		var calls atomic.Int32

		err := ExecuteConcurrent(ctx, makeIntRange(50), 10, func(_ context.Context, _ int) error {
			calls.Add(1)
			return nil
		})

		assert.ErrorIs(t, err, context.Canceled)
		assert.Zero(t, calls.Load())
	})

	t.Run("first error is returned", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("the failing one")

		err := ExecuteConcurrent(context.Background(), makeIntRange(50), 10, func(_ context.Context, v int) error {
			if v == 3 {
				return wantErr
			}
			// stall the others so v=3 is the earliest finisher
			time.Sleep(50 * time.Millisecond)
			return nil
		})

		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("first error cancels in-flight siblings ctx", func(t *testing.T) {
		t.Parallel()

		wantErr := errors.New("boom")
		var sawCancel atomic.Bool

		err := ExecuteConcurrent(context.Background(), makeIntRange(40), 10, func(ctx context.Context, v int) error {
			if v == 1 {
				return wantErr
			}
			select {
			case <-ctx.Done():
				sawCancel.Store(true)
				return nil
			case <-time.After(500 * time.Millisecond):
				return nil
			}
		})

		assert.ErrorIs(t, err, wantErr)
		assert.True(t, sawCancel.Load(), "at least one sibling should have observed ctx.Done after the failing item")
	})

	t.Run("parent ctx cancelled mid-run does not spawn extra goroutines", func(t *testing.T) {
		t.Parallel()

		const limit = 5
		const n = 1000

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		release := make(chan struct{})
		var calls atomic.Int32

		done := make(chan struct{})
		go func() {
			_ = ExecuteConcurrent(ctx, makeIntRange(n), limit, func(_ context.Context, _ int) error {
				calls.Add(1)
				<-release
				return nil
			})
			close(done)
		}()

		// Wait until the first `limit` items have started running and the
		// producer is about to park on the full semaphore for iter limit+1.
		require.Eventually(t, func() bool {
			return calls.Load() == int32(limit)
		}, time.Second, time.Millisecond)

		// Give the producer goroutine time to enter the select and park
		// on `sem <- struct{}{}` before we cancel the ctx. Without this,
		// the producer might still be pre-select; cancelling then is harmless,
		// but the test is clearer when we cancel mid-park.
		time.Sleep(20 * time.Millisecond)

		cancel()
		close(release)

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("ExecuteConcurrent did not return after cancellation")
		}

		// The strict guarantee: only the in-flight `limit` items were
		// invoked. Iter limit+1 (parked on sem when the cancel landed)
		// must not have been scheduled.
		assert.Equal(t, int32(limit), calls.Load())
	})
}

// makeIntRange returns the slice [0, n).
func makeIntRange(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// updateMax atomically lifts m up to v if v > m.
func updateMax(m *atomic.Int32, v int32) {
	for {
		old := m.Load()
		if v <= old || m.CompareAndSwap(old, v) {
			return
		}
	}
}

// esDurations approximates the measured per-record duration distribution of
// the acquisitions signal resolver (heavy right-skew, p50 ~461ms, max ~2312ms),
// scaled 1000x to microseconds so the benchmark runs fast while keeping the
// shape. Three groups of 10, each containing one tail outlier - so every
// batch of 10 in BatchExecute hits a slow item and stalls its peers.
// Numbers are based on real measurements of ES calls.
var esDurations = []time.Duration{
	100 * time.Microsecond, 200 * time.Microsecond, 300 * time.Microsecond, 400 * time.Microsecond, 500 * time.Microsecond,
	600 * time.Microsecond, 700 * time.Microsecond, 800 * time.Microsecond, 900 * time.Microsecond, 2000 * time.Microsecond,
	150 * time.Microsecond, 250 * time.Microsecond, 400 * time.Microsecond, 500 * time.Microsecond, 600 * time.Microsecond,
	700 * time.Microsecond, 800 * time.Microsecond, 900 * time.Microsecond, 1100 * time.Microsecond, 1900 * time.Microsecond,
	180 * time.Microsecond, 300 * time.Microsecond, 450 * time.Microsecond, 550 * time.Microsecond, 650 * time.Microsecond,
	750 * time.Microsecond, 850 * time.Microsecond, 1000 * time.Microsecond, 1300 * time.Microsecond, 2300 * time.Microsecond,
}

// BenchmarkWallTime_ESDistribution compares per-call wall time of the two
// concurrency strategies. BatchExecute mirrors the prior "errgroup inside
// fixed-size batch" pattern from the signal resolvers - every batch waits
// for its slowest item before the next batch starts. ExecuteConcurrent uses
// a ctx-aware semaphore with no batch barrier - freed slots are reused
// immediately.
func BenchmarkWallTime_ESDistribution(b *testing.B) {
	b.Run("BatchExecute", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = BatchExecute(esDurations, 10, func(batch []time.Duration) error {
				eg, _ := errgroup.WithContext(context.Background())
				for _, d := range batch {
					eg.Go(func() error {
						time.Sleep(d)
						return nil
					})
				}
				return eg.Wait()
			})
		}
	})

	b.Run("ExecuteConcurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ExecuteConcurrent(context.Background(), esDurations, 10, func(_ context.Context, d time.Duration) error {
				time.Sleep(d)
				return nil
			})
		}
	})
}
