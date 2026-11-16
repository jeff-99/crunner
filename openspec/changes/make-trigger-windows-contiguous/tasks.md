## 1. Window computation

- [ ] 1.1 Remove `schedulingEpoch` from `internalScheduler` and `Start` (scheduler.go ~147, ~573-574); verify `go build ./...`
- [ ] 1.2 Add a window origin recorded when a unit is first scheduled while running (registration on a running scheduler or `Start`) and use it instead of `createdAt` for the first non-backfill window; verify with test 3.1
- [ ] 1.3 In `scheduleTask`, always pass the previous window end (or nil for the first window) to `Schedule.Next`, removing the `LastRunTime` branch (~314-325); verify with tests 3.2 and 3.3
- [ ] 1.4 Update `Schedule` and `intervalSchedule` doc comments in `schedule.go` to say the reference time is the previous window end; verify by reading `go doc`

## 2. Expected trigger time and drift

- [ ] 2.1 Record the previous run's finish time when the task goroutine hands the unit back for rescheduling, and set expected trigger time / `fireAfter` to `max(windowEnd, previousFinish)`; verify with test 3.4
- [ ] 2.2 Replace the skip path's `LastRunTime = now` (~406) with grouped skipping: advance windows while their end is more than MaxDrift before now, call the cancel callback once, then schedule the first remaining window; verify with test 3.5

## 3. Tests (testing/synctest)

- [ ] 3.1 Task registered on an idle scheduler, Start 1 minute later: first window starts at Start time, run is not skipped
- [ ] 3.2 Interval 10s task whose runs take 2s: after 100 windows the last window ends exactly origin + 1000s and every window starts at the previous end (fails on main)
- [ ] 3.3 Custom schedule records its `lastRun` argument: equals previous window end, not run start
- [ ] 3.4 Interval 10s, MaxDrift 1s, task runs 25s: cancel callback not called, next window starts at previous window end, next run starts immediately after the long run (fails on main)
- [ ] 3.5 Interval 10s, MaxDrift 5s, main loop/timer delayed 60s (e.g. block the first run's queueing): cancel callback called exactly once and the next executed window starts at the end of the last skipped window
- [ ] 3.6 Existing `TestTaskTriggerWindows` and `TestTaskTriggerWindowsWithBackfill` still pass unchanged

## 4. Verification

- [ ] 4.1 Run `go test -race ./...` and confirm all tests pass
