package crunner_test

import (
	"context"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"

	s "github.com/jeff-99/crunner"
)

func TestCronScheduleEverySecondWithLastRun(t *testing.T) {
	schedule, err := s.WithCronSchedule("* * * * * *", time.UTC, true)
	assert.NoError(t, err)

	now := time.Now().UTC()
	lastRun := now.Add(-time.Second)
	next, err := schedule.Next(context.Background(), now, &lastRun)
	assert.NoError(t, err)

	assert.LessOrEqual(t, next.Sub(now), 1*time.Millisecond)

}

func TestCronScheduleEverySecond(t *testing.T) {
	schedule, err := s.WithCronSchedule("* * * * * *", time.UTC, true)
	assert.NoError(t, err)

	now := time.Now().UTC()
	next, err := schedule.Next(context.Background(), now, nil)
	assert.NoError(t, err)

	assert.LessOrEqual(t, next.Sub(now), 1*time.Second)

}

func mustLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	assert.NoError(t, err)
	return loc
}

func TestCronScheduleNilLocationDefaultsToLocal(t *testing.T) {
	schedule, err := s.WithCronSchedule("0 9 * * *", nil, false)
	assert.NoError(t, err)
	assert.Same(t, time.Local, schedule.Location)

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	lastRun := now.Add(-time.Hour)
	assert.NotPanics(t, func() {
		_, err := schedule.Next(context.Background(), now, nil)
		assert.NoError(t, err)
		_, err = schedule.Next(context.Background(), now, &lastRun)
		assert.NoError(t, err)
	})
}

func TestCronScheduleFirstRunHonoursLocation(t *testing.T) {
	ny := mustLocation(t, "America/New_York")
	schedule, err := s.WithCronSchedule("0 9 * * *", ny, false)
	assert.NoError(t, err)

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	next, err := schedule.Next(context.Background(), now, nil)
	assert.NoError(t, err)

	assert.True(t, time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC).Equal(next), "got %s", next)
	assert.Equal(t, ny.String(), next.Location().String())
}

func TestCronScheduleLaterRunHonoursLocation(t *testing.T) {
	ny := mustLocation(t, "America/New_York")
	schedule, err := s.WithCronSchedule("0 9 * * *", ny, false)
	assert.NoError(t, err)

	now := time.Date(2026, 1, 1, 15, 0, 0, 0, time.UTC)
	lastRun := time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC)
	next, err := schedule.Next(context.Background(), now, &lastRun)
	assert.NoError(t, err)

	assert.True(t, time.Date(2026, 1, 2, 14, 0, 0, 0, time.UTC).Equal(next), "got %s", next)
}

func TestCronScheduleTZPrefixOverridesLocation(t *testing.T) {
	ny := mustLocation(t, "America/New_York")
	schedule, err := s.WithCronSchedule("CRON_TZ=Asia/Tokyo 0 9 * * *", ny, false)
	assert.NoError(t, err)

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	next, err := schedule.Next(context.Background(), now, nil)
	assert.NoError(t, err)

	// 09:00 Tokyo (UTC+9) is 00:00Z, so the next one after 12:00Z on 1 January is 2 January.
	assert.True(t, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC).Equal(next), "got %s", next)
}

func TestCronTaskFirstTriggerWindowHonoursLocation(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		ny := mustLocation(t, "America/New_York")
		schedule, err := s.WithCronSchedule("0 9 * * *", ny, false)
		assert.NoError(t, err)

		start := time.Now() // synctest epoch: 2000-01-01T00:00:00Z
		scheduler := s.New()
		var mu sync.Mutex
		var windowEnd time.Time
		runs := 0

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error {
				mu.Lock()
				defer mu.Unlock()
				runs++
				windowEnd = meta.TriggerWindowEnd
				return nil
			},
			schedule,
			s.WithID("cron-ny"),
		))

		time.Sleep(14*time.Hour + time.Second)
		synctest.Wait()
		scheduler.Stop(context.Background())

		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, 1, runs)
		assert.True(t, start.Add(14*time.Hour).Equal(windowEnd), "window end %s, want 14:00Z (09:00 New York)", windowEnd)
	})
}
