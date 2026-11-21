package crunner

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestZeroValueCronScheduleDoesNotPanic(t *testing.T) {
	parsed, err := ParseCron("0 9 * * *", false)
	assert.NoError(t, err)
	schedule := cronSchedule{schedule: parsed}

	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	lastRun := now.Add(-time.Hour)
	assert.NotPanics(t, func() {
		_, err := schedule.Next(context.Background(), now, nil)
		assert.NoError(t, err)
		_, err = schedule.Next(context.Background(), now, &lastRun)
		assert.NoError(t, err)
	})
}
