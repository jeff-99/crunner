package crunner_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"

	s "github.com/jeff-99/crunner"
)

func tickXSeconds(d time.Duration) {
	seconds := d / time.Second
	for i := 0; i < int(seconds); i++ {
		time.Sleep(1 * time.Second)
	}

	time.Sleep(1 * time.Nanosecond)

	// Yield to allow the scheduler to run and trigger the last tasksk
	runtime.Gosched()
}

func TestSchedulingSimpleTask(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()

		count := atomic.Int32{}

		task := func(ctx context.Context) error {
			count.Add(1)
			return nil
		}
		scheduler.Start(t.Context())

		scheduler.RegisterTask(
			s.SimpleTask(task),
			s.WithIntervalSchedule(1*time.Second),
			s.WithID("test-task"),
		)

		tickXSeconds(5 * time.Second)

		scheduler.Stop()

		assert.Equal(t, int32(5), count.Load())

	})

}

func TestSchedulingTaskEveryXSeconds(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		interval          time.Duration
		numberOfIntervals float32
		expectedRuns      int64
	}{
		"every 1 second": {
			interval:          1 * time.Second,
			numberOfIntervals: 5,
			expectedRuns:      5, // there is a time tick every second so with 5 seconds we will have 6 ticks so task triggers 5 times
		},
		"every 2 seconds": {
			interval:          2 * time.Second,
			numberOfIntervals: 5,
			expectedRuns:      5, // there is a time tick every second so with 5*2 seconds we will only have 6 ticks so task triggers 5 times
		},
		"every 1.5 seconds": {
			interval:          1500 * time.Millisecond,
			numberOfIntervals: 5,
			expectedRuns:      4, // there is a time tick every second so with 5*1.5 seconds we will only have 7 ticks so task triggers 4 times
		},
		"every 2 seconds for 5 days": {
			interval:          2 * time.Second,
			numberOfIntervals: float32((1 * time.Hour * 24 * 5) / (2 * time.Second)),
			expectedRuns:      int64((1 * time.Hour * 24 * 5) / (2 * time.Second)),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				scheduler := s.New()

				count := atomic.Int32{}

				task := func(ctx context.Context) error {
					count.Add(1)
					return nil
				}
				scheduler.Start(t.Context())

				interval := test.interval
				numberOfIntervals := test.numberOfIntervals

				scheduler.RegisterTask(
					s.SimpleTask(task),
					s.WithIntervalSchedule(interval),
					s.WithID("test-task"),
				)

				// dummy task that will run every second, which is not measured
				scheduler.RegisterTask(
					s.SimpleTask(func(ctx context.Context) error { return nil }),
					s.WithIntervalSchedule(1*time.Second),
					s.WithID("test-task-dummy"),
				)

				tickXSeconds(time.Duration(numberOfIntervals) * interval)

				scheduler.Stop()

				expectedCount := int32(test.expectedRuns)
				assert.Equal(t, expectedCount, count.Load())

			})
		})
	}

}

func TestSchedulingTaskAfterXIntervalsOfRunning(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()

		intervalScheduleCount := atomic.Int32{}

		intervalTask := func(ctx context.Context) error {
			intervalScheduleCount.Add(1)
			return nil
		}

		cronScheduleCount := atomic.Int32{}
		cronTask := func(ctx context.Context) error {
			cronScheduleCount.Add(1)
			return nil
		}

		scheduler.Start(t.Context())

		// first wait for 5 seconds to let the scheduler run
		// and then register the task
		tickXSeconds(5 * time.Second)

		scheduler.RegisterTask(
			s.SimpleTask(intervalTask),
			s.WithIntervalSchedule(1*time.Second),
			s.WithID("interval-task"),
		)

		schedule, _ := s.WithCronSchedule("* * * * * *", time.Now().Location(), true)
		scheduler.RegisterTask(
			s.SimpleTask(cronTask),
			schedule,
			s.WithID("cron-task"),
		)

		tickXSeconds(2 * time.Second)

		assert.Equal(t, int32(2), intervalScheduleCount.Load())
		assert.Equal(t, int32(2), cronScheduleCount.Load())

		tickXSeconds(5 * time.Second)

		assert.Equal(t, int32(7), intervalScheduleCount.Load())
		assert.Equal(t, int32(7), cronScheduleCount.Load())

		scheduler.Stop()

	})

}

func TestRemovingTaskStopsExecution(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()

		count := atomic.Int32{}

		task := func(ctx context.Context) error {
			count.Add(1)
			return nil
		}
		scheduler.Start(t.Context())

		scheduler.RegisterTask(
			s.SimpleTask(task),
			s.WithIntervalSchedule(1*time.Second),
			s.WithID("test-task"),
		)

		tickXSeconds(2 * time.Second)
		assert.Equal(t, int32(2), count.Load())

		scheduler.RemoveTask("test-task")

		tickXSeconds(2 * time.Second)

		scheduler.Stop()

		assert.Equal(t, int32(2), count.Load())

	})

}

func TestStoppingAndStartingScheduler(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()

		count := atomic.Int32{}

		task := func(ctx context.Context) error {
			count.Add(1)
			// t.Log(time.Now())
			return nil
		}
		scheduler.Start(t.Context())

		scheduler.RegisterTask(
			s.SimpleTask(task),
			s.WithIntervalSchedule(1*time.Second),
			s.WithID("test-task"),
		)

		tickXSeconds(2 * time.Second)
		assert.Equal(t, int32(2), count.Load())

		scheduler.Stop()

		tickXSeconds(2 * time.Second)

		scheduler.Start(t.Context())

		tickXSeconds(2 * time.Second)

		// Windows are contiguous: the two windows missed while stopped ([2s,3s) and [3s,4s)) are within
		// MaxDrift, so both run on restart, followed by [4s,5s) and [5s,6s): 2 + 4 = 6 runs.
		assert.Equal(t, int32(6), count.Load())

		scheduler.Stop()

	})

}

func TestSchedulerSuccessCallback(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()

		count := atomic.Int32{}

		task := func(ctx context.Context) error {
			count.Add(1)
			if count.Load()%2 == 0 {
				return fmt.Errorf("execution count divisible by 2")
			}
			return nil
		}

		succesfulExecution := atomic.Int32{}

		scheduler.Start(t.Context())

		scheduler.RegisterTask(
			s.SimpleTask(task),
			s.WithIntervalSchedule(1*time.Second),
			s.WithID("test-task"),
			s.WithSuccessCallback(func(ctx context.Context) {
				succesfulExecution.Add(1)
			}),
		)

		tickXSeconds(2 * time.Second)
		assert.Equal(t, int32(2), count.Load())
		assert.Equal(t, int32(1), succesfulExecution.Load())

		scheduler.Stop()

	})

}

func TestSchedulerErrorCallback(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()

		count := atomic.Int32{}

		task := func(ctx context.Context) error {
			count.Add(1)
			if count.Load()%2 == 0 {
				return fmt.Errorf("execution count divisible by 2")
			}

			return nil
		}

		erroredExecutions := atomic.Int32{}

		scheduler.Start(t.Context())

		scheduler.RegisterTask(
			s.SimpleTask(task),
			s.WithIntervalSchedule(1*time.Second),
			s.WithID("test-task"),
			s.WithErrorCallback(func(ctx context.Context, err error) {
				erroredExecutions.Add(1)
			}),
		)

		tickXSeconds(2 * time.Second)
		assert.Equal(t, int32(2), count.Load())
		assert.Equal(t, int32(1), erroredExecutions.Load())

		scheduler.Stop()

	})

}

func TestTriggerMetadataInTask(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()

		schedulingStart := time.Now()

		var mu sync.Mutex
		var last s.TriggerMetadata

		task := func(ctx context.Context, meta s.TriggerMetadata) error {
			mu.Lock()
			defer mu.Unlock()
			last = meta
			return nil
		}
		snapshot := func() (expected, actual, triggerWindowStart, triggerWindowEnd time.Time) {
			mu.Lock()
			defer mu.Unlock()
			return last.ExpectedTriggerTime, last.ActualTriggerTime, last.TriggerWindowStart, last.TriggerWindowEnd
		}

		scheduler.Start(t.Context())

		scheduler.RegisterTask(
			task,
			s.WithIntervalSchedule(2*time.Second),
			s.WithID("test-task"),
		)

		// trigger the first execution
		tickXSeconds(1 * time.Second)

		expected, actual, triggerWindowStart, triggerWindowEnd := snapshot()
		assert.Equal(t, time.Time{}, expected)
		assert.Equal(t, time.Time{}, actual)
		assert.Equal(t, time.Time{}, triggerWindowStart)
		assert.Equal(t, time.Time{}, triggerWindowEnd)

		// trigger the second execution
		tickXSeconds(1 * time.Second)

		expected, actual, triggerWindowStart, triggerWindowEnd = snapshot()

		assert.Equal(t, schedulingStart.Add(2*time.Second), expected)
		assert.Equal(t, schedulingStart.Add(2*time.Second), actual)
		assert.Equal(t, schedulingStart, triggerWindowStart)
		assert.Equal(t, schedulingStart.Add(2*time.Second), triggerWindowEnd)

		// schedule to next trigger
		tickXSeconds(2 * time.Second)

		expected, actual, triggerWindowStart, triggerWindowEnd = snapshot()

		assert.Equal(t, schedulingStart.Add(4*time.Second), expected)
		assert.Equal(t, schedulingStart.Add(4*time.Second), actual)
		assert.Equal(t, schedulingStart.Add(2*time.Second), triggerWindowStart)
		assert.Equal(t, schedulingStart.Add(4*time.Second), triggerWindowEnd)

		scheduler.Stop()

	})
}

func TestTaskWithBackfill(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()

		var counter atomic.Int32

		task := func(ctx context.Context, meta s.TriggerMetadata) error {
			counter.Add(1)
			return nil
		}

		scheduler.Start(t.Context())

		scheduler.RegisterTask(
			task,
			s.WithIntervalSchedule(10*time.Second),
			s.WithID("test-task"),
			s.WithNotBefore(time.Now().Add(-100*time.Second)),
			s.WithBackfill(),
		)

		tickXSeconds(1 * time.Second)

		assert.Equal(t, 10, int(counter.Load()))

		tickXSeconds(8 * time.Second)

		assert.Equal(t, 10, int(counter.Load()))

		scheduler.Stop()

	})
}

type triggerWindow struct {
	start, end time.Time
}

func TestTaskTriggerWindows(t *testing.T) {
	tests := map[string]struct {
		interval          time.Duration
		numberOfIntervals float32
	}{
		"every 1 second": {
			interval:          1 * time.Second,
			numberOfIntervals: 5,
		},
		"every 2 seconds": {
			interval:          2 * time.Second,
			numberOfIntervals: 5,
		},
		"every 1.5 seconds": {
			interval:          1500 * time.Millisecond,
			numberOfIntervals: 5,
		},
		"every 2 seconds for 5 days": {
			interval:          2 * time.Second,
			numberOfIntervals: float32((1 * time.Hour * 24 * 5) / (2 * time.Second)),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				interval := test.interval
				numberOfIntervals := test.numberOfIntervals
				epoch := time.Now()

				scheduler := s.New()

				var windows []triggerWindow
				task := func(ctx context.Context, meta s.TriggerMetadata) error {
					windows = append(windows, triggerWindow{
						start: meta.TriggerWindowStart,
						end:   meta.TriggerWindowEnd,
					})
					return nil
				}

				scheduler.Start(t.Context())

				scheduler.RegisterTask(
					task,
					s.WithIntervalSchedule(interval),
					s.WithID("test-task"),
				)

				tickXSeconds(time.Duration(numberOfIntervals) * interval)

				// Stop waits for running tasks, so reading windows afterwards is race-free.
				scheduler.Stop()

				if interval%time.Second == 0 {
					// interval task always start immediately, so we should have one more window than the number of intervals
					assert.Equal(t, int(numberOfIntervals), len(windows))
				} else {
					// if the interval is not a multiple of a second then the last run will not have triggered yet, so we should have one less window
					assert.Equal(t, int(numberOfIntervals)-1, len(windows))
				}

				for i, w := range windows {
					if i == 0 {
						// Interval task always start immediately, given no time has passed since the start the first window should be the same as the epoch
						assert.Equal(t, epoch, w.start)
						assert.Equal(t, epoch.Add(time.Duration(1)*interval), w.end)
					} else {
						// for each subsequent window the start time should be the end time of the previous window and the end time should be the start time + the interval
						assert.Equal(t, epoch.Add(time.Duration(i)*interval), w.start)
						assert.Equal(t, epoch.Add(time.Duration(i+1)*interval), w.end)
					}
				}
			})
		})
	}
}

func TestTaskTriggerWindowsWithBackfill(t *testing.T) {
	tests := map[string]struct {
		interval          time.Duration
		numberOfIntervals float32
		startTimeDelta    time.Duration
	}{
		"every 1 second": {
			interval:          1 * time.Second,
			numberOfIntervals: 5,
			startTimeDelta:    100 * time.Second,
		},
		"every 2 seconds": {
			interval:          2 * time.Second,
			numberOfIntervals: 5,
			startTimeDelta:    100 * time.Second,
		},
		"every 1.5 seconds": {
			interval:          1500 * time.Millisecond,
			numberOfIntervals: 5,
			startTimeDelta:    100 * time.Second,
		},
		"every 2 seconds for 5 days": {
			interval:          2 * time.Second,
			numberOfIntervals: float32((1 * time.Hour * 24 * 5) / (2 * time.Second)),
			startTimeDelta:    100 * time.Second,
		},
		"backfill shouldn't trigger": {
			interval:          1 * time.Hour,
			numberOfIntervals: 2,
			startTimeDelta:    (1 * time.Hour) / 2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				interval := test.interval
				numberOfIntervals := test.numberOfIntervals
				startTimeDelta := test.startTimeDelta
				epoch := time.Now()

				scheduler := s.New()

				var windows []triggerWindow
				task := func(ctx context.Context, meta s.TriggerMetadata) error {
					windows = append(windows, triggerWindow{
						start: meta.TriggerWindowStart,
						end:   meta.TriggerWindowEnd,
					})
					return nil
				}

				scheduler.Start(t.Context())

				scheduler.RegisterTask(
					task,
					s.WithIntervalSchedule(interval),
					s.WithID("test-task"),
					s.WithNotBefore(epoch.Add(-1*startTimeDelta)),
					s.WithBackfill(),
				)

				tickXSeconds(time.Duration(numberOfIntervals) * interval)

				// Stop waits for running tasks, so reading windows afterwards is race-free.
				scheduler.Stop()

				// with backfill enabled and a start time before the epoch, so we can calculate the number of backfill jobs and add them to the number of intervals
				// backfills := int(startTimeDelta / interval)
				// assert.Equal(t, int(numberOfIntervals)+backfills, len(windows))
				totalSimulatedTime := time.Duration(numberOfIntervals)*interval + startTimeDelta
				expectedWindows := int(totalSimulatedTime / interval)

				assert.Equal(t, expectedWindows, len(windows))

				windowEpoch := epoch.Add(-1 * startTimeDelta)
				for i, w := range windows {
					assert.Equal(t, windowEpoch.Add(time.Duration(i)*interval), w.start)
					assert.Equal(t, windowEpoch.Add(time.Duration(i+1)*interval), w.end)
				}
			})
		})
	}

}

// Mock schedule that always errors on Next()
type errorSchedule struct{}

func (e errorSchedule) Next(ctx context.Context, schedulingStart time.Time, lastRun *time.Time) (time.Time, error) {
	return time.Time{}, fmt.Errorf("mock schedule error")
}

func TestScheduleNextError(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		runCount := atomic.Int32{}

		task := func(ctx context.Context, meta s.TriggerMetadata) error {
			runCount.Add(1)
			return nil
		}

		scheduler.Start(t.Context())

		// Register task with the erroring schedule
		err := scheduler.RegisterTask(
			task,
			errorSchedule{},
			s.WithID("error-schedule-test"),
		)
		assert.NoError(t, err, "RegisterTask should succeed")

		// Wait a short time to allow the main loop to attempt scheduling
		time.Sleep(100 * time.Millisecond)
		runtime.Gosched() // Yield to increase chance of main loop running
		time.Sleep(100 * time.Millisecond)

		scheduler.Stop()

		// Assert that the task never ran because scheduling failed
		assert.Equal(t, int32(0), runCount.Load(), "Task should not run if schedule.Next errors")
	})
}

func TestRegisterTaskOverwrite(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		count1 := atomic.Int32{}
		count2 := atomic.Int32{}
		taskID := "overwrite-test-task"

		// Task 1 increments count1
		task1 := func(ctx context.Context, meta s.TriggerMetadata) error {
			count1.Add(1)
			return nil
		}

		// Task 2 increments count2
		task2 := func(ctx context.Context, meta s.TriggerMetadata) error {
			count2.Add(1)
			return nil
		}

		scheduler.Start(t.Context())

		// Register Task 1
		err := scheduler.RegisterTask(
			task1,
			s.WithIntervalSchedule(1*time.Second),
			s.WithID(taskID),
		)
		assert.NoError(t, err)

		// Let Task 1 run once
		tickXSeconds(1 * time.Second)
		assert.Equal(t, int32(1), count1.Load(), "Task 1 should run once before overwrite")
		assert.Equal(t, int32(0), count2.Load(), "Task 2 should not have run yet")

		// Register Task 2 with the same ID, overwriting Task 1
		err = scheduler.RegisterTask(
			task2,                                 // Different task function
			s.WithIntervalSchedule(1*time.Second), // Same schedule
			s.WithID(taskID),                      // Same ID
		)
		assert.NoError(t, err)

		// Allow immediate execution of task2 due to overwrite scheduling near interval boundary
		runtime.Gosched()
		time.Sleep(50 * time.Millisecond) // Short sleep to check whether an immediate run occurs
		assert.Equal(t, int32(1), count1.Load(), "Task 1 count should remain 1 after overwrite")
		assert.Equal(t, int32(0), count2.Load(), "Task 2 should not run immediately after overwrite")

		// Wait for the next full interval
		tickXSeconds(1 * time.Second)

		// Assert that Task 2 ran again at the end of the interval
		assert.Equal(t, int32(1), count1.Load(), "Task 1 count should still be 1")
		assert.Equal(t, int32(1), count2.Load(), "Task 2 should run after 1 second")

		// Wait for another interval
		tickXSeconds(1 * time.Second)

		// Assert that only Task 2 ran again
		assert.Equal(t, int32(1), count1.Load(), "Task 1 count should still remain 1")
		assert.Equal(t, int32(2), count2.Load(), "Task 2 should run a second time")

		scheduler.Stop()
	})
}

func TestRemoveTaskRepeatedly(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		taskID := "remove-repeatedly-test"
		count := atomic.Int32{}

		task := func(ctx context.Context, meta s.TriggerMetadata) error {
			count.Add(1)
			return nil
		}

		scheduler.Start(t.Context())

		// Register the task
		err := scheduler.RegisterTask(
			task,
			s.WithIntervalSchedule(1*time.Second),
			s.WithID(taskID),
		)
		assert.NoError(t, err, "RegisterTask should succeed")

		// Let it run once to be sure it's active
		tickXSeconds(1 * time.Second)
		assert.Equal(t, int32(1), count.Load(), "Task should run once before removal")

		// Remove the task (first time)
		err = scheduler.RemoveTask(taskID)
		assert.NoError(t, err, "First RemoveTask should succeed")

		// Wait to ensure it doesn't run again
		tickXSeconds(1 * time.Second)
		assert.Equal(t, int32(1), count.Load(), "Task should not run after removal")

		// Remove the task again (second time)
		err = scheduler.RemoveTask(taskID)
		assert.Error(t, err, "Second RemoveTask for the same ID should fail")
		assert.Contains(t, err.Error(), fmt.Sprintf("task with id %s not found", taskID), "Error message should indicate task not found")

		scheduler.Stop()
	})
}

func TestRemoveTaskCancelCallback(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		taskID := "cancel-callback-test"
		taskRunCount := atomic.Int32{}
		cancelCallbackCount := atomic.Int32{}

		task := func(ctx context.Context, meta s.TriggerMetadata) error {
			taskRunCount.Add(1)
			return nil
		}

		cancelCallback := func(ctx context.Context) {
			cancelCallbackCount.Add(1)
		}

		scheduler.Start(t.Context())

		// Register the task with a cancel callback
		err := scheduler.RegisterTask(
			task,
			s.WithIntervalSchedule(1*time.Second),
			s.WithID(taskID),
			s.WithCancelCallback(cancelCallback),
		)
		assert.NoError(t, err, "RegisterTask should succeed")

		// Let it run once
		tickXSeconds(1 * time.Second)
		assert.Equal(t, int32(1), taskRunCount.Load(), "Task should run once before removal")
		assert.Equal(t, int32(0), cancelCallbackCount.Load(), "Cancel callback should not have run yet")

		// Remove the task
		err = scheduler.RemoveTask(taskID)
		assert.NoError(t, err, "RemoveTask should succeed")

		// Allow time for the cancel callback goroutine to execute
		runtime.Gosched()
		time.Sleep(50 * time.Millisecond)

		// Assert cancel callback was called
		assert.Equal(t, int32(1), cancelCallbackCount.Load(), "Cancel callback should run once after removal")

		// Wait to ensure task doesn't run again
		tickXSeconds(1 * time.Second)
		assert.Equal(t, int32(1), taskRunCount.Load(), "Task should not run after removal")

		scheduler.Stop()
	})
}

func TestOverwriteTaskCancelCallback(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		taskID := "overwrite-cancel-test"
		task1RunCount := atomic.Int32{}
		task2RunCount := atomic.Int32{}
		cancel1CallbackCount := atomic.Int32{}

		// Task 1
		task1 := func(ctx context.Context, meta s.TriggerMetadata) error {
			task1RunCount.Add(1)
			return nil
		}

		// Cancel callback for Task 1
		cancelCallback1 := func(ctx context.Context) {
			cancel1CallbackCount.Add(1)
		}

		// Task 2 (no cancel callback needed for this test)
		task2 := func(ctx context.Context, meta s.TriggerMetadata) error {
			task2RunCount.Add(1)
			return nil
		}

		scheduler.Start(t.Context())

		// Register Task 1 with its cancel callback
		err := scheduler.RegisterTask(
			task1,
			s.WithIntervalSchedule(1*time.Second),
			s.WithID(taskID),
			s.WithCancelCallback(cancelCallback1), // Add cancel callback here
		)
		assert.NoError(t, err, "RegisterTask 1 should succeed")

		// Let Task 1 run once
		tickXSeconds(1 * time.Second)
		assert.Equal(t, int32(1), task1RunCount.Load(), "Task 1 should run once before overwrite")
		assert.Equal(t, int32(0), cancel1CallbackCount.Load(), "Cancel callback 1 should not have run yet")
		assert.Equal(t, int32(0), task2RunCount.Load(), "Task 2 should not have run yet")

		// Register Task 2 with the same ID, overwriting Task 1
		err = scheduler.RegisterTask(
			task2,
			s.WithIntervalSchedule(1*time.Second), // Same schedule
			s.WithID(taskID),                      // Same ID
			// No cancel callback needed for task2 in this test
		)
		assert.NoError(t, err, "RegisterTask 2 (overwrite) should succeed")

		// Allow time for the overwrite logic (including cancel callback) to execute
		runtime.Gosched()
		time.Sleep(50 * time.Millisecond)

		// Assert cancel callback for Task 1 was called during overwrite
		assert.Equal(t, int32(1), cancel1CallbackCount.Load(), "Cancel callback 1 should run once after overwrite")
		assert.Equal(t, int32(1), task1RunCount.Load(), "Task 1 count should remain 1")
		assert.Equal(t, task2RunCount.Load(), int32(0), "Task 2 should run at least once after overwrite")

		// Wait for another interval to ensure only task 2 runs
		tickXSeconds(1 * time.Second)
		assert.Equal(t, int32(1), task1RunCount.Load(), "Task 1 count should still be 1")
		assert.Equal(t, task2RunCount.Load(), int32(1), "Task 2 should run again") // Should be at least 2 now
		assert.Equal(t, int32(1), cancel1CallbackCount.Load(), "Cancel callback 1 count should remain 1")

		scheduler.Stop()
	})
}

// lateOnceSchedule returns a trigger time in the past on its first call, so the
// first run is always late, and a far-future time afterwards.
type lateOnceSchedule struct {
	calls atomic.Int32
}

func (l *lateOnceSchedule) Next(ctx context.Context, schedulingStart time.Time, lastRun *time.Time) (time.Time, error) {
	if l.calls.Add(1) == 1 {
		return time.Now().Add(-time.Hour), nil
	}
	return time.Now().Add(24 * time.Hour), nil
}

func TestMassDriftSkipDoesNotStallScheduler(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		const lateTasks = 300

		var cancelled atomic.Int32
		var canaryRuns atomic.Int32
		var lateRuns atomic.Int32
		schedules := make([]*lateOnceSchedule, lateTasks)

		for i := range lateTasks {
			schedules[i] = &lateOnceSchedule{}
			err := scheduler.RegisterTask(
				func(ctx context.Context, meta s.TriggerMetadata) error { lateRuns.Add(1); return nil },
				schedules[i],
				s.WithID(fmt.Sprintf("late-%d", i)),
				s.WithMaxDrift(time.Second),
				s.WithCancelCallback(func(ctx context.Context) { cancelled.Add(1) }),
			)
			assert.NoError(t, err)
		}
		err := scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { canaryRuns.Add(1); return nil },
			s.WithIntervalSchedule(time.Second),
			s.WithID("canary"),
		)
		assert.NoError(t, err)

		scheduler.Start(t.Context())
		tickXSeconds(5 * time.Second)

		assert.Equal(t, int32(lateTasks), cancelled.Load(), "every late task should be cancelled")
		assert.Equal(t, int32(0), lateRuns.Load(), "cancelled runs are not executed")
		for i, sch := range schedules {
			assert.Equal(t, int32(2), sch.calls.Load(), "late task %d should be rescheduled (merged window) once", i)
		}
		assert.Equal(t, int32(5), canaryRuns.Load(), "canary task should keep running")

		scheduler.Stop()
	})
}

func TestBlockingDriftSkipCallbackDoesNotBlockOtherTasks(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		var canaryRuns atomic.Int32
		sch := &lateOnceSchedule{}
		release := make(chan struct{})
		callbackReturned := make(chan struct{})

		scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { return nil },
			sch,
			s.WithID("late"),
			s.WithMaxDrift(time.Second),
			s.WithCancelCallback(func(ctx context.Context) {
				<-release
				close(callbackReturned)
			}),
		)
		scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { canaryRuns.Add(1); return nil },
			s.WithIntervalSchedule(time.Second),
			s.WithID("canary"),
		)

		scheduler.Start(t.Context())
		tickXSeconds(3 * time.Second)

		assert.Equal(t, int32(3), canaryRuns.Load(), "other tasks keep running while the callback blocks")
		assert.Equal(t, int32(2), sch.calls.Load(), "skipped task is rescheduled while the callback blocks")

		stopped := make(chan struct{})
		go func() {
			scheduler.Stop()
			close(stopped)
		}()
		synctest.Wait()
		select {
		case <-stopped:
			t.Fatal("Stop returned while the skip callback was still running")
		default:
		}

		close(release)
		<-callbackReturned
		<-stopped
	})
}

func TestStressConcurrentRegisterReplaceRemoveStop(t *testing.T) {
	scheduler := s.New(s.WithSummaryInterval(time.Millisecond))
	assert.NoError(t, scheduler.Start(t.Context()))

	var stableRuns atomic.Int64
	assert.NoError(t, scheduler.RegisterTask(
		func(ctx context.Context, meta s.TriggerMetadata) error {
			stableRuns.Add(1)
			return nil
		},
		s.WithIntervalSchedule(time.Millisecond),
		s.WithID("stable"),
	))

	var wg sync.WaitGroup
	for g := range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			deadline := time.Now().Add(300 * time.Millisecond)
			for i := 0; time.Now().Before(deadline); i++ {
				id := fmt.Sprintf("stress-%d", (g+i)%6)
				_ = scheduler.RegisterTask(
					func(ctx context.Context, meta s.TriggerMetadata) error { return nil },
					s.WithIntervalSchedule(time.Millisecond),
					s.WithID(id),
				)
				if i%3 == 0 {
					_ = scheduler.RemoveTask(id)
				}
			}
		}()
	}
	wg.Wait()

	assert.NoError(t, scheduler.Stop())
	assert.Positive(t, stableRuns.Load(), "an untouched task should keep running during the churn")
}

// fixedSchedule always triggers at a fixed offset from the moment Next is called.
type fixedSchedule struct{ after time.Duration }

func (f fixedSchedule) Next(ctx context.Context, schedulingStart time.Time, lastRun *time.Time) (time.Time, error) {
	return time.Now().Add(f.after), nil
}

func TestSupersededTaskPendingTimerDoesNotRun(t *testing.T) {
	for name, supersede := range map[string]func(scheduler s.Scheduler) error{
		"replaced": func(scheduler s.Scheduler) error {
			return scheduler.RegisterTask(
				func(ctx context.Context, meta s.TriggerMetadata) error { return nil },
				fixedSchedule{after: time.Hour},
				s.WithID("task"),
			)
		},
		"removed": func(scheduler s.Scheduler) error { return scheduler.RemoveTask("task") },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				scheduler := s.New()
				var oldRuns atomic.Int32

				assert.NoError(t, scheduler.Start(t.Context()))
				assert.NoError(t, scheduler.RegisterTask(
					func(ctx context.Context, meta s.TriggerMetadata) error { oldRuns.Add(1); return nil },
					fixedSchedule{after: 5 * time.Second},
					s.WithID("task"),
				))

				tickXSeconds(1 * time.Second)
				assert.NoError(t, supersede(scheduler))

				tickXSeconds(10 * time.Second)
				assert.Equal(t, int32(0), oldRuns.Load(), "old task must not run after its pending timer fires")

				scheduler.Stop()
			})
		})
	}
}

func TestConcurrentRegistrationsSameIDCancelDisplacedExactlyOnce(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		assert.NoError(t, scheduler.Start(t.Context()))

		const n = 50
		var runs atomic.Int32
		cancels := make([]atomic.Int32, n)

		var wg sync.WaitGroup
		start := make(chan struct{})
		for i := range n {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				assert.NoError(t, scheduler.RegisterTask(
					func(ctx context.Context, meta s.TriggerMetadata) error { runs.Add(1); return nil },
					s.WithIntervalSchedule(time.Second),
					s.WithID("same"),
					s.WithCancelCallback(func(ctx context.Context) { cancels[i].Add(1) }),
				))
			}()
		}
		close(start)
		wg.Wait()

		tickXSeconds(1 * time.Second)

		total := int32(0)
		for i := range cancels {
			c := cancels[i].Load()
			assert.LessOrEqual(t, c, int32(1), "cancel callback %d must fire at most once", i)
			total += c
		}
		assert.Equal(t, int32(n-1), total, "every displaced task is cancelled exactly once")
		assert.Equal(t, int32(1), runs.Load(), "exactly one registration remains and runs")
		assert.Len(t, scheduler.ScheduledTasks(), 1)

		scheduler.Stop()
	})
}

func TestConcurrentRemoveSameIDSucceedsOnce(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		assert.NoError(t, scheduler.Start(t.Context()))

		var cancelled atomic.Int32
		assert.NoError(t, scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { return nil },
			s.WithIntervalSchedule(time.Second),
			s.WithID("victim"),
			s.WithCancelCallback(func(ctx context.Context) { cancelled.Add(1) }),
		))

		const n = 50
		var successes atomic.Int32
		var wg sync.WaitGroup
		start := make(chan struct{})
		for range n {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				if scheduler.RemoveTask("victim") == nil {
					successes.Add(1)
				}
			}()
		}
		close(start)
		wg.Wait()
		synctest.Wait()

		assert.Equal(t, int32(1), successes.Load(), "exactly one RemoveTask succeeds")
		assert.Equal(t, int32(1), cancelled.Load(), "cancel callback fires exactly once")

		scheduler.Stop()
	})
}

func TestPanicErrorUnwrap(t *testing.T) {
	t.Parallel()
	var err error = fmt.Errorf("wrapped: %w", &s.PanicError{Value: io.EOF})

	var pe *s.PanicError
	assert.True(t, errors.As(err, &pe))
	assert.True(t, errors.Is(err, io.EOF))

	assert.NoError(t, (&s.PanicError{Value: "boom"}).Unwrap())
	assert.Contains(t, (&s.PanicError{Value: "boom"}).Error(), "boom")
}

func TestPanickingTaskIsReportedAsError(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		var mu sync.Mutex
		var gotErr error
		var successes atomic.Int32

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { panic("boom") },
			s.WithIntervalSchedule(time.Second),
			s.WithID("panicker"),
			s.WithErrorCallback(func(ctx context.Context, err error) {
				mu.Lock()
				defer mu.Unlock()
				gotErr = err
			}),
			s.WithSuccessCallback(func(ctx context.Context) { successes.Add(1) }),
		))

		tickXSeconds(1 * time.Second)
		scheduler.Stop()

		mu.Lock()
		defer mu.Unlock()
		var pe *s.PanicError
		if assert.True(t, errors.As(gotErr, &pe), "error callback should receive a *PanicError") {
			assert.Equal(t, "boom", pe.Value)
			assert.NotEmpty(t, pe.Stack)
		}
		assert.Equal(t, int32(0), successes.Load())
	})
}

func TestPanickingTaskKeepsRunningAndDoesNotAffectOthers(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		var panics, healthy, errorCallbacks atomic.Int32

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { panics.Add(1); panic("always") },
			s.WithIntervalSchedule(time.Second),
			s.WithID("panicker"),
			s.WithErrorCallback(func(ctx context.Context, err error) { errorCallbacks.Add(1) }),
		))
		assert.NoError(t, scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { healthy.Add(1); return nil },
			s.WithIntervalSchedule(time.Second),
			s.WithID("healthy"),
		))

		tickXSeconds(5 * time.Second)
		scheduler.Stop()

		assert.Equal(t, int32(5), panics.Load(), "panicking task is rescheduled on every window")
		assert.Equal(t, int32(5), errorCallbacks.Load())
		assert.Equal(t, int32(5), healthy.Load(), "healthy task keeps its schedule")
	})
}

func TestPanicWithErrorValueUnwraps(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		var mu sync.Mutex
		var gotErr error

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { panic(io.EOF) },
			s.WithIntervalSchedule(time.Second),
			s.WithID("eof"),
			s.WithErrorCallback(func(ctx context.Context, err error) {
				mu.Lock()
				defer mu.Unlock()
				gotErr = err
			}),
		))

		tickXSeconds(1 * time.Second)
		scheduler.Stop()

		mu.Lock()
		defer mu.Unlock()
		assert.True(t, errors.Is(gotErr, io.EOF))
	})
}

func TestRegisterTaskValidation(t *testing.T) {
	t.Parallel()
	okTask := func(ctx context.Context, meta s.TriggerMetadata) error { return nil }
	now := time.Now()
	var nilSuccess s.SuccessCallback
	var nilError s.ErrorCallback
	var nilCancel s.CancelCallback

	tests := map[string]struct {
		task     s.Task
		schedule s.Schedule
		opts     []s.SchedulingOption
		wantErrs []error
	}{
		"nil task":                    {nil, s.WithIntervalSchedule(time.Second), nil, []error{s.ErrInvalidTask}},
		"nil schedule":                {okTask, nil, nil, []error{s.ErrInvalidSchedule}},
		"typed nil":                   {okTask, (*lateOnceSchedule)(nil), nil, []error{s.ErrInvalidSchedule}},
		"zero interval":               {okTask, s.WithIntervalSchedule(0), nil, []error{s.ErrInvalidSchedule}},
		"negative interval":           {okTask, s.WithIntervalSchedule(-time.Second), nil, []error{s.ErrInvalidSchedule}},
		"negative drift":              {okTask, s.WithIntervalSchedule(time.Second), []s.SchedulingOption{s.WithMaxDrift(-time.Second)}, []error{s.ErrInvalidOption}},
		"not after equals not before": {okTask, s.WithIntervalSchedule(time.Second), []s.SchedulingOption{s.WithNotBefore(now), s.WithNotAfter(now)}, []error{s.ErrInvalidOption}},
		"not after before not before": {okTask, s.WithIntervalSchedule(time.Second), []s.SchedulingOption{s.WithNotBefore(now), s.WithNotAfter(now.Add(-time.Hour))}, []error{s.ErrInvalidOption}},
		"backfill without not before": {okTask, s.WithIntervalSchedule(time.Second), []s.SchedulingOption{s.WithBackfill()}, []error{s.ErrInvalidOption}},
		"nil success callback":        {okTask, s.WithIntervalSchedule(time.Second), []s.SchedulingOption{s.WithSuccessCallback(nilSuccess)}, []error{s.ErrInvalidOption}},
		"nil error callback":          {okTask, s.WithIntervalSchedule(time.Second), []s.SchedulingOption{s.WithErrorCallback(nilError)}, []error{s.ErrInvalidOption}},
		"nil cancel callback":         {okTask, s.WithIntervalSchedule(time.Second), []s.SchedulingOption{s.WithCancelCallback(nilCancel)}, []error{s.ErrInvalidOption}},
		"multiple problems":           {nil, s.WithIntervalSchedule(time.Second), []s.SchedulingOption{s.WithMaxDrift(-time.Second)}, []error{s.ErrInvalidTask, s.ErrInvalidOption}},
		"all three kinds":             {nil, nil, []s.SchedulingOption{s.WithMaxDrift(-time.Second)}, []error{s.ErrInvalidTask, s.ErrInvalidSchedule, s.ErrInvalidOption}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			scheduler := s.New()
			err := scheduler.RegisterTask(tc.task, tc.schedule, tc.opts...)
			if assert.Error(t, err) {
				for _, want := range tc.wantErrs {
					assert.ErrorIs(t, err, want)
				}
			}
			assert.Empty(t, scheduler.ScheduledTasks(), "nothing may be stored for a rejected registration")
		})
	}
}

func TestRegisterTaskValidRegistrationsAreAccepted(t *testing.T) {
	t.Parallel()
	now := time.Now()
	okTask := func(ctx context.Context, meta s.TriggerMetadata) error { return nil }

	for name, opts := range map[string][]s.SchedulingOption{
		"defaults":       nil,
		"zero drift":     {s.WithMaxDrift(0)},
		"valid range":    {s.WithNotBefore(now), s.WithNotAfter(now.Add(time.Hour))},
		"backfill":       {s.WithNotBefore(now.Add(-time.Hour)), s.WithBackfill()},
		"only not after": {s.WithNotAfter(now)},
		"all callbacks":  {s.WithSuccessCallback(func(context.Context) {}), s.WithErrorCallback(func(context.Context, error) {}), s.WithCancelCallback(func(context.Context) {})},
	} {
		t.Run(name, func(t *testing.T) {
			scheduler := s.New()
			assert.NoError(t, scheduler.RegisterTask(okTask, s.WithIntervalSchedule(time.Second), opts...))
			assert.Len(t, scheduler.ScheduledTasks(), 1)
		})
	}
}

func TestRejectedRegistrationLeavesExistingTaskIntact(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		var runs, cancels atomic.Int32

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { runs.Add(1); return nil },
			s.WithIntervalSchedule(time.Second),
			s.WithID("a"),
			s.WithCancelCallback(func(ctx context.Context) { cancels.Add(1) }),
		))

		err := scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error { return nil },
			s.WithIntervalSchedule(0),
			s.WithID("a"),
		)
		assert.ErrorIs(t, err, s.ErrInvalidSchedule)

		tickXSeconds(3 * time.Second)
		scheduler.Stop()

		assert.Equal(t, int32(3), runs.Load(), "original task keeps running on its schedule")
		assert.Equal(t, int32(0), cancels.Load(), "original task is not cancelled")
		assert.Len(t, scheduler.ScheduledTasks(), 1)
	})
}

type recordedWindow struct {
	start, end, ran time.Time
}

// windowRecorder records the trigger window and start time of every run.
type windowRecorder struct {
	mu   sync.Mutex
	runs []recordedWindow
}

func (w *windowRecorder) task(runFor time.Duration) s.Task {
	return func(ctx context.Context, meta s.TriggerMetadata) error {
		w.mu.Lock()
		w.runs = append(w.runs, recordedWindow{meta.TriggerWindowStart, meta.TriggerWindowEnd, time.Now()})
		w.mu.Unlock()
		time.Sleep(runFor)
		return nil
	}
}

func (w *windowRecorder) get() []recordedWindow {
	w.mu.Lock()
	defer w.mu.Unlock()
	return append([]recordedWindow(nil), w.runs...)
}

func TestFirstWindowOfTaskRegisteredBeforeStartBeginsAtStart(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		rec := &windowRecorder{}
		var cancels atomic.Int32

		assert.NoError(t, scheduler.RegisterTask(rec.task(0), s.WithIntervalSchedule(10*time.Second), s.WithID("idle"),
			s.WithCancelCallback(func(context.Context) { cancels.Add(1) })))

		time.Sleep(time.Minute)
		start := time.Now()
		assert.NoError(t, scheduler.Start(t.Context()))
		time.Sleep(10*time.Second + time.Millisecond)
		synctest.Wait()
		scheduler.Stop()

		runs := rec.get()
		if assert.Len(t, runs, 1) {
			assert.True(t, start.Equal(runs[0].start), "window start %s, want Start time %s", runs[0].start, start)
			assert.True(t, start.Add(10*time.Second).Equal(runs[0].end))
		}
		assert.Equal(t, int32(0), cancels.Load(), "first run must not be cancelled")
	})
}

func TestWindowsStayOnGridWhenRunsStartLate(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		rec := &windowRecorder{}
		origin := time.Now()

		assert.NoError(t, scheduler.Start(t.Context()))
		// Each run takes 12s, so every following run starts after its window ended.
		assert.NoError(t, scheduler.RegisterTask(rec.task(12*time.Second), s.WithIntervalSchedule(10*time.Second), s.WithID("late")))

		time.Sleep(100*12*time.Second + 20*time.Second)
		synctest.Wait()
		scheduler.Stop()

		runs := rec.get()
		if assert.GreaterOrEqual(t, len(runs), 100) {
			for i, r := range runs[:100] {
				assert.True(t, origin.Add(time.Duration(i)*10*time.Second).Equal(r.start), "window %d start %s", i, r.start.Sub(origin))
				assert.True(t, origin.Add(time.Duration(i+1)*10*time.Second).Equal(r.end), "window %d end %s", i, r.end.Sub(origin))
			}
			assert.True(t, origin.Add(1000*time.Second).Equal(runs[99].end))
		}
	})
}

// recordingSchedule is an interval schedule that records the reference time it is called with.
type recordingSchedule struct {
	interval time.Duration
	mu       sync.Mutex
	refs     []time.Time
}

func (r *recordingSchedule) Next(ctx context.Context, start time.Time, lastRun *time.Time) (time.Time, error) {
	if lastRun == nil {
		return start.Add(r.interval), nil
	}
	r.mu.Lock()
	r.refs = append(r.refs, *lastRun)
	r.mu.Unlock()
	return lastRun.Add(r.interval), nil
}

func TestScheduleReferenceTimeIsPreviousWindowEnd(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		rec := &windowRecorder{}
		sch := &recordingSchedule{interval: 10 * time.Second}

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(rec.task(13*time.Second), sch, s.WithID("ref")))

		time.Sleep(60 * time.Second)
		synctest.Wait()
		scheduler.Stop()

		runs := rec.get()
		sch.mu.Lock()
		defer sch.mu.Unlock()
		assert.GreaterOrEqual(t, len(sch.refs), 3)
		for i, ref := range sch.refs {
			if i < len(runs) {
				assert.True(t, runs[i].end.Equal(ref), "reference %d is %s, want previous window end %s (run started %s)", i, ref, runs[i].end, runs[i].ran)
			}
		}
	})
}

func TestLongRunningTaskIsNotCancelled(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		rec := &windowRecorder{}
		var cancels atomic.Int32
		origin := time.Now()

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(rec.task(25*time.Second), s.WithIntervalSchedule(10*time.Second), s.WithID("slow"),
			s.WithMaxDrift(time.Second),
			s.WithCancelCallback(func(context.Context) { cancels.Add(1) })))

		time.Sleep(36 * time.Second)
		synctest.Wait()
		scheduler.Stop()

		runs := rec.get()
		assert.Equal(t, int32(0), cancels.Load(), "a long run must not cancel its successor")
		if assert.Len(t, runs, 2) {
			assert.True(t, origin.Add(10*time.Second).Equal(runs[1].start), "second window starts at the first window's end")
			assert.True(t, origin.Add(20*time.Second).Equal(runs[1].end), "second window spans one interval")
			assert.True(t, origin.Add(35*time.Second).Equal(runs[1].ran), "second run starts right after the first finishes (10s + 25s)")
		}
	})
}

// blockingOnceSchedule fires once after `first`; computing its next run blocks the main loop for `block`.
type blockingOnceSchedule struct {
	first, block time.Duration
}

func (b blockingOnceSchedule) Next(ctx context.Context, start time.Time, lastRun *time.Time) (time.Time, error) {
	if lastRun == nil {
		return start.Add(b.first), nil
	}
	time.Sleep(b.block)
	return start.Add(24 * time.Hour), nil
}

func TestLateRunIsCancelledAndMergedIntoNextWindow(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		rec := &windowRecorder{}
		var cancels atomic.Int32
		origin := time.Now()

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(rec.task(0), s.WithIntervalSchedule(10*time.Second), s.WithID("task"),
			s.WithMaxDrift(5*time.Second),
			s.WithCancelCallback(func(context.Context) { cancels.Add(1) })))
		// Runs at 19s; rescheduling it blocks the main loop until 27s, so the task's 20s timer is handled at 27s.
		assert.NoError(t, scheduler.RegisterTask(func(context.Context, s.TriggerMetadata) error { return nil },
			blockingOnceSchedule{first: 19 * time.Second, block: 8 * time.Second}, s.WithID("blocker"), s.WithMaxDrift(time.Hour)))

		time.Sleep(29 * time.Second)
		synctest.Wait()
		assert.Equal(t, int32(1), cancels.Load(), "late run is cancelled once")
		assert.Len(t, rec.get(), 1, "no run is executed at 27s")

		time.Sleep(2 * time.Second)
		synctest.Wait()
		scheduler.Stop()

		runs := rec.get()
		if assert.Len(t, runs, 2) {
			assert.True(t, origin.Add(10*time.Second).Equal(runs[1].start), "merged window starts at the cancelled window's start, got %s", runs[1].start.Sub(origin))
			assert.True(t, origin.Add(30*time.Second).Equal(runs[1].end), "merged window ends at the next grid point, got %s", runs[1].end.Sub(origin))
			assert.True(t, origin.Add(30*time.Second).Equal(runs[1].ran), "merged run fires at its normal time, got %s", runs[1].ran.Sub(origin))
		}
		assert.Equal(t, int32(1), cancels.Load())
	})
}

func TestRestartAfterLongStopMergesMissedWindows(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New()
		rec := &windowRecorder{}
		var cancels atomic.Int32
		origin := time.Now()

		assert.NoError(t, scheduler.Start(t.Context()))
		assert.NoError(t, scheduler.RegisterTask(rec.task(0), s.WithIntervalSchedule(10*time.Second), s.WithID("task"),
			s.WithMaxDrift(5*time.Second),
			s.WithCancelCallback(func(context.Context) { cancels.Add(1) })))

		time.Sleep(15 * time.Second)
		assert.NoError(t, scheduler.Stop())
		time.Sleep(68 * time.Second) // now 83s
		assert.NoError(t, scheduler.Start(t.Context()))
		synctest.Wait()

		assert.Equal(t, int32(1), cancels.Load(), "pending window is cancelled once on restart")
		assert.Len(t, rec.get(), 1, "no run at 83s")

		time.Sleep(17*time.Second + time.Millisecond) // now 100s
		synctest.Wait()
		scheduler.Stop()

		runs := rec.get()
		if assert.Len(t, runs, 3) {
			assert.True(t, origin.Add(10*time.Second).Equal(runs[1].start), "got %s", runs[1].start.Sub(origin))
			assert.True(t, origin.Add(90*time.Second).Equal(runs[1].end), "got %s", runs[1].end.Sub(origin))
			assert.True(t, origin.Add(90*time.Second).Equal(runs[1].ran), "got %s", runs[1].ran.Sub(origin))
			assert.True(t, origin.Add(90*time.Second).Equal(runs[2].start))
			assert.True(t, origin.Add(100*time.Second).Equal(runs[2].end))
		}
		assert.Equal(t, int32(1), cancels.Load())
	})
}

// deadlineRecorder records the deadline of each run's context relative to the run start.
func deadlineRecorder(mu *sync.Mutex, deadlines *[]*time.Duration, block bool) s.Task {
	return func(ctx context.Context, meta s.TriggerMetadata) error {
		start := time.Now()
		var d *time.Duration
		if dl, ok := ctx.Deadline(); ok {
			v := dl.Sub(start)
			d = &v
		}
		mu.Lock()
		*deadlines = append(*deadlines, d)
		mu.Unlock()
		if block {
			<-ctx.Done()
		}
		return ctx.Err()
	}
}

func TestTaskTimeouts(t *testing.T) {
	day := 24 * time.Hour
	tests := map[string]struct {
		schedulerOpts []s.SchedulerOption
		taskOpts      []s.SchedulingOption
		want          *time.Duration
	}{
		"default 30 days":       {nil, nil, ptr(30 * day)},
		"scheduler-wide 5s":     {[]s.SchedulerOption{s.WithTaskTimeout(5 * time.Second)}, nil, ptr(5 * time.Second)},
		"per-task 1s override":  {[]s.SchedulerOption{s.WithTaskTimeout(5 * time.Second)}, []s.SchedulingOption{s.WithTimeout(time.Second)}, ptr(time.Second)},
		"scheduler disabled":    {[]s.SchedulerOption{s.WithTaskTimeout(0)}, nil, nil},
		"per-task disabled":     {[]s.SchedulerOption{s.WithTaskTimeout(5 * time.Second)}, []s.SchedulingOption{s.WithTimeout(0)}, nil},
		"per-task negative off": {nil, []s.SchedulingOption{s.WithTimeout(-time.Second)}, nil},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			synctest.Test(t, func(t *testing.T) {
				scheduler := s.New(tc.schedulerOpts...)
				var mu sync.Mutex
				var deadlines []*time.Duration

				assert.NoError(t, scheduler.Start(t.Context()))
				opts := append([]s.SchedulingOption{s.WithID("t")}, tc.taskOpts...)
				assert.NoError(t, scheduler.RegisterTask(deadlineRecorder(&mu, &deadlines, false), s.WithIntervalSchedule(time.Second), opts...))

				tickXSeconds(time.Second)
				scheduler.Stop()

				mu.Lock()
				defer mu.Unlock()
				if assert.Len(t, deadlines, 1) {
					if tc.want == nil {
						assert.Nil(t, deadlines[0], "run context must have no deadline")
					} else if assert.NotNil(t, deadlines[0]) {
						assert.Equal(t, *tc.want, *deadlines[0])
					}
				}
			})
		})
	}
}

func TestTaskTimeoutCancelsRun(t *testing.T) {
	t.Parallel()
	synctest.Test(t, func(t *testing.T) {
		scheduler := s.New(s.WithTaskTimeout(5 * time.Second))
		var mu sync.Mutex
		var gotErr error
		var ended time.Time

		assert.NoError(t, scheduler.Start(t.Context()))
		start := time.Now()
		assert.NoError(t, scheduler.RegisterTask(
			func(ctx context.Context, meta s.TriggerMetadata) error {
				<-ctx.Done()
				mu.Lock()
				if ended.IsZero() {
					ended = time.Now()
				}
				mu.Unlock()
				return ctx.Err()
			},
			s.WithIntervalSchedule(time.Second), s.WithID("slow"),
			s.WithErrorCallback(func(ctx context.Context, err error) {
				mu.Lock()
				if gotErr == nil {
					gotErr = err
				}
				mu.Unlock()
			}),
		))

		time.Sleep(7 * time.Second)
		synctest.Wait()
		scheduler.Stop()

		mu.Lock()
		defer mu.Unlock()
		assert.ErrorIs(t, gotErr, context.DeadlineExceeded)
		assert.True(t, start.Add(6*time.Second).Equal(ended), "run (started at 1s) cancelled 5s later, got %s", ended.Sub(start))
	})
}

// ctxSchedule records its context deadline and, after its first call, blocks until the context is done.
type ctxSchedule struct {
	mu       sync.Mutex
	calls    int
	deadline *time.Duration
}

func (c *ctxSchedule) Next(ctx context.Context, start time.Time, lastRun *time.Time) (time.Time, error) {
	c.mu.Lock()
	c.calls++
	calls := c.calls
	if dl, ok := ctx.Deadline(); ok && calls == 1 {
		d := dl.Sub(time.Now())
		c.deadline = &d
	}
	c.mu.Unlock()
	if calls == 1 {
		return start.Add(time.Second), nil
	}
	<-ctx.Done()
	return time.Time{}, ctx.Err()
}

func TestScheduleTimeouts(t *testing.T) {
	t.Run("default deadline is 1 minute", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			scheduler := s.New()
			sch := &ctxSchedule{}
			assert.NoError(t, scheduler.Start(t.Context()))
			assert.NoError(t, scheduler.RegisterTask(func(context.Context, s.TriggerMetadata) error { return nil }, sch, s.WithID("t")))
			synctest.Wait()
			scheduler.Stop()

			sch.mu.Lock()
			defer sch.mu.Unlock()
			if assert.NotNil(t, sch.deadline) {
				assert.Equal(t, time.Minute, *sch.deadline)
			}
		})
	})

	t.Run("exceeded timeout stops rescheduling", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			scheduler := s.New(s.WithScheduleTimeout(time.Second))
			sch := &ctxSchedule{}
			var runs, canary atomic.Int32
			assert.NoError(t, scheduler.Start(t.Context()))
			assert.NoError(t, scheduler.RegisterTask(func(context.Context, s.TriggerMetadata) error { runs.Add(1); return nil }, sch, s.WithID("t")))
			assert.NoError(t, scheduler.RegisterTask(func(context.Context, s.TriggerMetadata) error { canary.Add(1); return nil },
				s.WithIntervalSchedule(time.Second), s.WithID("canary")))

			tickXSeconds(10 * time.Second)
			scheduler.Stop()

			sch.mu.Lock()
			defer sch.mu.Unlock()
			assert.Equal(t, int32(1), runs.Load(), "task is not scheduled again after its schedule timed out")
			assert.Equal(t, 2, sch.calls)
			assert.GreaterOrEqual(t, canary.Load(), int32(8), "scheduler keeps running other tasks")
		})
	})

	t.Run("disabled timeout has no deadline", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			scheduler := s.New(s.WithScheduleTimeout(0))
			sch := &ctxSchedule{}
			assert.NoError(t, scheduler.Start(t.Context()))
			assert.NoError(t, scheduler.RegisterTask(func(context.Context, s.TriggerMetadata) error { return nil }, sch, s.WithID("t")))
			synctest.Wait()
			scheduler.Stop()

			sch.mu.Lock()
			defer sch.mu.Unlock()
			assert.Nil(t, sch.deadline)
		})
	})
}

func ptr[T any](v T) *T { return &v }
