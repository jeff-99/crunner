# Tasks

## 1. Scheduler-wide options

- [ ] 1.1 Add `taskTimeout`, `scheduleTimeout` and `queueSize` fields to `internalScheduler` with defaults of 30 days, 1 minute and 100 in `New`; create the channels after the options are applied, clamping the queue size to >= 1; verify with a unit test of the defaults and the clamping
- [ ] 1.2 Add the `WithTaskTimeout`, `WithScheduleTimeout` and `WithQueueSize` options with doc comments stating the defaults and the `<= 0` behaviour; verify `go vet ./...` passes
- [ ] 1.3 Use `scheduleTimeout` in `scheduleTask` (WithCancel when <= 0); verify with a `testing/synctest` test where a custom schedule blocks on `ctx.Done()` and returns after the configured timeout, and the task is not rescheduled

## 2. Task timeout

- [ ] 2.1 Add `SchedulingOptions.Timeout *time.Duration` and the `WithTimeout(d)` option; verify that it compiles and that the existing tests pass
- [ ] 2.2 In `processTask`, resolve the effective timeout (task, then scheduler) and use WithCancel when <= 0; verify with `testing/synctest` tests for the default deadline, a scheduler-wide 5s timeout, a per-task 1s override and the disabled case (`ctx.Deadline()` reports no deadline)
- [ ] 2.3 Run `go test -race ./...` and verify that it passes
