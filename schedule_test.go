package crunner_test

import (
	"context"
	"testing"
	"time"
	s "github.com/jeff-99/crunner"

	"github.com/stretchr/testify/assert"
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
