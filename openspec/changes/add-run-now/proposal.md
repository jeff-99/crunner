## Why

Operators often need to trigger a scheduled task immediately (after a fix, from an admin endpoint, in tests) without waiting for its next trigger time. crunner has no way to do this short of re-registering the task, which resets its windows and fires the cancel callback (gocron v2 offers `Job.RunNow`).

## What Changes

- Add `RunNow(id string) error` to the `Scheduler` interface.
- A manual run executes the task immediately, outside the regular schedule: it does not consume, shift or shorten any scheduled window.
- The manual run covers the window from the end of the task's last completed scheduled window (or its first window's start, if none completed) up to the moment of the request, and its metadata marks it as manual.
- Manual runs are never skipped for drift, count towards the run statistics and invoke the success/error callbacks.
- A task never runs concurrently with itself: `RunNow` returns an error if the task is currently running; a scheduled trigger that fires during a manual run waits for it and is then subject to MaxDrift as usual.
- `RunNow` returns errors for unknown IDs and when the scheduler is not running.
- **BREAKING**: adds a method to the `Scheduler` interface (breaks third-party implementations/mocks) and a field `Manual bool` to `TriggerMetadata`.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: adds manual runs, their window and their interplay with scheduled runs.

## Impact

- `scheduler.go`: `Scheduler` interface, `TriggerMetadata.Manual`, a run-now request channel handled by the main loop, a per-unit "running" flag.
- Exported error values `ErrTaskNotFound`, `ErrTaskRunning`, `ErrSchedulerNotRunning`.
- Dependencies: depends on `fix-shared-state-races` (the "running" flag and window metadata must be owned by the main loop) and should land after `fix-mainloop-self-send-deadlock`. A scheduled trigger that falls due during a manual run behaves as during any other run: it starts afterwards if it is within MaxDrift late, otherwise it is skipped and merged into the next window.
