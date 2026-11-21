package crunner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSchedulerDefaults(t *testing.T) {
	s := New().(*internalScheduler)
	assert.Equal(t, 30*24*time.Hour, s.taskTimeout)
	assert.Equal(t, time.Minute, s.scheduleTimeout)
	assert.Equal(t, 100, s.queueSize)
	assert.Equal(t, 100, cap(s.newTaskCh))
	assert.Equal(t, 100, cap(s.tasksIn))
	assert.Equal(t, 100, cap(s.doneCh))
}

func TestSchedulerOptionsAndQueueClamp(t *testing.T) {
	s := New(WithTaskTimeout(5*time.Second), WithScheduleTimeout(time.Second), WithQueueSize(7)).(*internalScheduler)
	assert.Equal(t, 5*time.Second, s.taskTimeout)
	assert.Equal(t, time.Second, s.scheduleTimeout)
	assert.Equal(t, 7, cap(s.tasksIn))

	for _, n := range []int{0, -5} {
		s := New(WithQueueSize(n)).(*internalScheduler)
		assert.Equal(t, 1, s.queueSize)
		assert.Equal(t, 1, cap(s.newTaskCh))
		assert.Equal(t, 1, cap(s.tasksIn))
		assert.Equal(t, 1, cap(s.doneCh))
	}
}
