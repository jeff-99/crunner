package crunner

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultOptionsHaveNoID(t *testing.T) {
	assert.Empty(t, WithDefaultOptions().ID)
}

func storedKeys(s *internalScheduler) []string {
	var keys []string
	s.tasks.Range(func(k, _ any) bool {
		keys = append(keys, k.(string))
		return true
	})
	return keys
}

func TestGeneratedIDIsStable(t *testing.T) {
	for name, opts := range map[string][]SchedulingOption{
		"no ID option": nil,
		"empty ID":     {WithID("")},
	} {
		t.Run(name, func(t *testing.T) {
			s := New().(*internalScheduler)
			assert.NoError(t, s.RegisterTask(func(context.Context, TriggerMetadata) error { return nil }, WithIntervalSchedule(time.Second), opts...))

			keys := storedKeys(s)
			if assert.Len(t, keys, 1) {
				assert.True(t, strings.HasPrefix(keys[0], "task-"), "generated key %q", keys[0])
				v, _ := s.tasks.Load(keys[0])
				su := v.(*schedulingUnit)
				assert.Equal(t, keys[0], su.ID())
				assert.Equal(t, su.ID(), su.ID(), "ID must not change between calls")

				assert.NoError(t, s.RemoveTask(keys[0]))
				assert.Empty(t, s.ScheduledTasks())
			}
		})
	}
}
