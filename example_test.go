package crunner_test

import (
	"context"
	"fmt"
	"time"

	s "github.com/jeff-99/crunner"
)

// ScheduledTasks and Task return read-only snapshots, including the IDs of tasks registered without
// an explicit ID.
func ExampleScheduler_ScheduledTasks() {
	scheduler := s.New()
	task := func(ctx context.Context, meta s.TriggerMetadata) error { return nil }

	_ = scheduler.RegisterTask(task, s.WithIntervalSchedule(time.Minute), s.WithID("report"))
	_ = scheduler.RegisterTask(task, s.WithIntervalSchedule(time.Hour), s.WithID("cleanup"))

	for _, info := range scheduler.ScheduledTasks() {
		fmt.Println(info.ID, info.Count, info.Running)
	}

	info, found := scheduler.Task("report")
	fmt.Println(info.ID, found)

	// Output:
	// cleanup 0 false
	// report 0 false
	// report true
}
