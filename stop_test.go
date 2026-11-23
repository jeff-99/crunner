package crunner_test

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	s "github.com/jeff-99/crunner"
)

// startStuckTask runs a scheduler with a task that ignores cancellation until release is closed.
func startStuckTask(t *testing.T) (s.Scheduler, chan struct{}) {
	t.Helper()

	scheduler := s.New()
	started := make(chan struct{}, 1)
	release := make(chan struct{})

	require.NoError(t, scheduler.Start(t.Context()))
	require.NoError(t, scheduler.RegisterTask(
		s.SimpleTask(func(ctx context.Context) error {
			select {
			case started <- struct{}{}:
			default:
			}
			<-release
			return nil
		}),
		s.WithIntervalSchedule(1*time.Second),
		s.WithID("stuck"),
	))

	tickXSeconds(1 * time.Second)
	<-started

	return scheduler, release
}

func TestStopBoundedByContext(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler, release := startStuckTask(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		begin := time.Now()
		err := scheduler.Stop(ctx)

		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Equal(t, 1*time.Second, time.Since(begin))

		close(release)
		assert.NoError(t, scheduler.Stop(context.Background()))
	})
}

func TestStartWhileDrainingAndIdleAfterwards(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler, release := startStuckTask(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		require.ErrorIs(t, scheduler.Stop(ctx), context.DeadlineExceeded)

		assert.Error(t, scheduler.Start(t.Context()), "Start must fail while a task is still draining")

		close(release)
		synctest.Wait()

		require.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.Stop(context.Background()))
	})
}

func TestSecondStopWaitsForDraining(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler, release := startStuckTask(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		require.ErrorIs(t, scheduler.Stop(ctx), context.DeadlineExceeded)

		result := make(chan error, 1)
		go func() {
			//nolint:staticcheck // a nil context must behave like context.Background()
			result <- scheduler.Stop(nil)
		}()

		synctest.Wait()
		select {
		case err := <-result:
			t.Fatalf("second Stop returned before the task did: %v", err)
		default:
		}

		close(release)
		assert.NoError(t, <-result)
	})
}

func TestSecondStopBoundedByItsOwnContext(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler, release := startStuckTask(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		require.ErrorIs(t, scheduler.Stop(ctx), context.DeadlineExceeded)

		ctx2, cancel2 := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2()
		err := scheduler.Stop(ctx2)
		assert.True(t, errors.Is(err, context.DeadlineExceeded))

		close(release)
		assert.NoError(t, scheduler.Stop(context.Background()))
	})
}
