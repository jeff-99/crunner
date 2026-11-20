package crunner_test

import (
	"context"
	"fmt"
	"runtime"
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
		
		schedule, _ := s.WithCronSchedule("* * * * * *", time.Now().Location() ,true)
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

		// expected five runs and not 4 because when the scheduler started again
		// the interval was ready to trigger again so it triggers directly and then 2 additional intervals were waited so two additional triggers
		assert.Equal(t, int32(5), count.Load())

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

		var expected, actual, triggerWindowStart, triggerWindowEnd time.Time

		task := func(ctx context.Context, meta s.TriggerMetadata) error {
			expected = meta.ExpectedTriggerTime
			actual = meta.ActualTriggerTime

			triggerWindowStart = meta.TriggerWindowStart
			triggerWindowEnd = meta.TriggerWindowEnd
			return nil
		}

		scheduler.Start(t.Context())

		scheduler.RegisterTask(
			task,
			s.WithIntervalSchedule(2*time.Second),
			s.WithID("test-task"),
		)

		// trigger the first execution
		tickXSeconds(1 * time.Second)

		assert.Equal(t, time.Time{}, expected)
		assert.Equal(t, time.Time{}, actual)
		assert.Equal(t, time.Time{}, triggerWindowStart)
		assert.Equal(t, time.Time{}, triggerWindowEnd)

		// trigger the second execution
		tickXSeconds(1 * time.Second)

		assert.Equal(t, schedulingStart.Add(2*time.Second), expected)
		assert.Equal(t, schedulingStart.Add(2*time.Second), actual)
		assert.Equal(t, schedulingStart, triggerWindowStart)
		assert.Equal(t, schedulingStart.Add(2*time.Second), triggerWindowEnd)

		// schedule to next trigger
		tickXSeconds(2 * time.Second)

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

				scheduler.Stop()
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
				scheduler.Stop()
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
			task2,                               // Different task function
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
		schedules := make([]*lateOnceSchedule, lateTasks)

		for i := range lateTasks {
			schedules[i] = &lateOnceSchedule{}
			err := scheduler.RegisterTask(
				func(ctx context.Context, meta s.TriggerMetadata) error { return nil },
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

		assert.Equal(t, int32(lateTasks), cancelled.Load(), "every late task should be skipped")
		for i, sch := range schedules {
			assert.Equal(t, int32(2), sch.calls.Load(), "late task %d should be rescheduled once", i)
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
