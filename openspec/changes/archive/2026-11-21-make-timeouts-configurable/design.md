# Design

## Context

There are three hardcoded values:
- `const channelBufferSize = 100` in `New` (scheduler.go:179), used for both `newTaskCh` and `tasksIn`;
- `context.WithTimeout(s.schedulerCtx, 1*time.Minute)` in `scheduleTask` (scheduler.go:286);
- `context.WithTimeout(ctx, 30*24*time.Hour)` in the run goroutine (scheduler.go:433).

`New` returns no error, so options can't reject bad values.

## Goals / Non-Goals

**Goals:**
- Expose all three values with today's defaults.
- Allow a per-task run timeout.

**Non-Goals:**
- A stop timeout. That is `add-stop-timeout`.
- Separate sizes for the two queues.
- A per-task schedule timeout.

## Decisions

- **Option names.**
  - Scheduler level: `WithTaskTimeout`, `WithScheduleTimeout`, `WithQueueSize`.
  - Task level: `WithTimeout`, stored in a new `SchedulingOptions.Timeout *time.Duration`.
  - The task-level option is a pointer so that "unset" (inherit the scheduler value) is different from "0" (disabled).
  - Alternative considered: a zero value meaning "inherit". That would leave no way to disable the timeout for one task.
- **`<= 0` disables the timeout.** This keeps `New` free of errors. The "disabled" case uses `context.WithCancel` instead of `WithTimeout`.
- **Clamp the queue size to at least 1.** Keeps the internal channels buffered, as they always have been, without making `New` return an error.
- **Resolve the timeout when the run starts.** `processTask` reads the unit's option or falls back to the scheduler value. Since there is no runtime mutation, no locking is needed.

## Risks / Trade-offs

- [Clamping hides bad input silently] → Documented on the option. The value is trivially inspectable in tests.
