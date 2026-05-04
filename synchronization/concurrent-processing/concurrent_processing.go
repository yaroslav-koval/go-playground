package concurrent_processing

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func BatchExecute[T any](s []T, batchSize int, f func([]T) error) error {
	for i := 0; i < len(s); i += batchSize {
		end := i + batchSize
		if end > len(s) {
			end = len(s)
		}

		if err := f(s[i:end]); err != nil {
			return err
		}
	}

	return nil
}

// ExecuteConcurrent runs f against each element of s in its own goroutine,
// with at most limit goroutines active at once. Returns the first error any
// f returns; ctx is cancelled on first error so in-flight calls can bail out
// through their own ctx checks. The loop stops enqueuing new work as soon as
// ctx is cancelled, by error or by the parent.
//
// Caveat: on parent cancellation is still runs 1 additional goroutine for a work.
// This logic relies on f(ctx, T) that is cancellation aware.
func ExecuteConcurrent[T any](ctx context.Context, s []T, limit int, f func(context.Context, T) error) error {
	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(limit)

	for _, v := range s {
		if ctx.Err() != nil {
			break
		}

		eg.Go(func() error {
			return f(ctx, v)
		})
	}

	return eg.Wait()
}

// ExecuteConcurrentStrict is the same shape as ExecuteConcurrent has custom semaphore,
// so it doesn't start a new goroutine when the parent context is canceled.
func ExecuteConcurrentStrict[T any](ctx context.Context, s []T, limit int, f func(context.Context, T) error) error {
	eg, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, limit)

	for _, v := range s {
		select {
		case <-ctx.Done():
			return eg.Wait()
		case sem <- struct{}{}:
		}

		eg.Go(func() error {
			defer func() { <-sem }()
			return f(ctx, v)
		})
	}

	return eg.Wait()
}
