# Proposal

## Why

Several limits are hardcoded in `scheduler.go`:
- every run gets a 30-day timeout (line 433);
- each `Schedule.Next` call gets a 1-minute context (line 286);
- both internal channels have a buffer of 100 (line 179).

Users can't tighten the run timeout for short jobs, give a slow custom schedule (for example, one that looks up a calendar remotely) more time, or size the buffers for many tasks.

## What Changes

- New scheduler options, defaulting to today's values:
  - `WithTaskTimeout(d)`: default 30 days. `d <= 0` disables the timeout, so the run is cancelled only when the scheduler stops.
  - `WithScheduleTimeout(d)`: default 1 minute. `d <= 0` disables the timeout.
  - `WithQueueSize(n)`: default 100. Values below 1 are clamped to 1.
- New per-task option `WithTimeout(d)`, which overrides the scheduler-wide task timeout for one task.
- No change in behaviour for users who don't set these options.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: the task timeout becomes configurable per scheduler and per task.
- `scheduler-lifecycle`: construction accepts the new timeout and queue-size options and documents their defaults.
- `schedules`: computing the next run time is bounded by a configurable timeout.

## Impact

- Code: `scheduler.go` (`New`, `internalScheduler` fields, `SchedulingOptions`, `scheduleTask`, `processTask`).
- API: purely additive. New `SchedulerOption`s, plus a new `SchedulingOption` and a `Timeout` field on `SchedulingOptions`. No breaking change.
- Dependencies on other changes: none.
- Coordination with other changes:
  - `validate-task-registration` could reject a negative per-task timeout. Here, negative values mean "no timeout", for consistency with the scheduler option.
  - `fix-mainloop-self-send-deadlock` (merged) removed the main loop's self-send, so a small queue size no longer risks that deadlock.
