## 1. Schedule

- [ ] 1.1 Add `ErrNoMoreRuns` and `oneTimeSchedule`/`At(t)` in `schedule.go`; verify with unit tests in `schedule_test.go` that `Next` returns `t` for a nil last run and `ErrNoMoreRuns` otherwise
- [ ] 1.2 Add doc comments for `At` and `ErrNoMoreRuns` describing window and drift behaviour; verify `go doc` shows them

## 2. Scheduler completion

- [ ] 2.1 In `scheduleTask`, on `errors.Is(err, ErrNoMoreRuns)` remove the unit via `CompareAndDelete` without calling the cancel callback and log at debug level; verify with a synctest test that a one-time task runs once, disappears from `ScheduledTasks`, and its cancel callback is not called
- [ ] 2.2 Add synctest tests for a past time within MaxDrift (runs immediately once) and beyond MaxDrift (never runs, task removed); verify both pass
- [ ] 2.3 Add a test that a custom schedule returning a different error keeps today's behaviour (task not rescheduled, still registered); verify it passes
- [ ] 2.4 Add a test that re-registering the same ID after completion schedules the new task; verify it passes with `go test -race ./...`
