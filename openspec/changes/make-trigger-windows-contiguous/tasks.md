## 1. Window computation

- [ ] 1.1 Remove `schedulingEpoch` from `internalScheduler` and `Start`; verify `go build ./...`
- [ ] 1.2 Add a window origin recorded when a unit is first scheduled while running (registration on a running scheduler or `Start`) and a `coveredUntil` per unit (initially the origin, or NotBefore when backfilling); start every window at `coveredUntil`; verify with test 3.1
- [ ] 1.3 In `scheduleTask`, always pass the previous window end (or nil for the first window) to `Schedule.Next`, removing the `LastRunTime` branch; advance `coveredUntil` when a run's result is handled; verify with tests 3.2 and 3.3
- [ ] 1.4 Update `Schedule` and `intervalSchedule` doc comments in `schedule.go` (reference time is a previous window end, never a run start) the `TriggerMetadata` doc (a window can span more than one interval) and the `WithMaxDrift` doc (a cancelled run's work waits for the next scheduled point, up to one interval); verify by reading `go doc`

## 2. Expected trigger time and drift

- [ ] 2.1 Add the run's finish time to `runResult`, store it on the unit in the main loop, and set expected trigger time / `fireAfter` to `max(windowEnd, previousFinish)`; verify with test 3.4
- [ ] 2.2 Replace the drift path's `LastRunTime = now` with cancel-and-merge: keep the async cancel callback, leave `coveredUntil` unchanged, advance the window end from the cancelled end with `Schedule.Next` until it is after now, and schedule the next run at that end; verify with tests 3.5 and 3.6

## 3. Tests (testing/synctest)

- [ ] 3.1 Task registered on an idle scheduler, Start 1 minute later: first window starts at Start time, run is not cancelled
- [ ] 3.2 Interval 10s task whose runs start late: after 100 windows the last window ends exactly origin + 1000s and every window starts at the previous end (fails on main)
- [ ] 3.3 Custom schedule records its `lastRun` argument: equals previous window end, not run start
- [ ] 3.4 Interval 10s, MaxDrift 1s, task runs 25s: cancel callback not called, next window starts at previous window end and spans one interval, next run starts immediately after the long run (fails on main)
- [ ] 3.5 Interval 10s, MaxDrift 5s, timer for [10s, 20s) fires at 27s (e.g. a blocking task delays it): callback once, no run at 27s, next run window [10s, 30s) triggered at 30s
- [ ] 3.6 Same task, Stop at 15s, Start at 83s: callback once, next run window [10s, 90s) triggered at 90s, following window [90s, 100s)
- [ ] 3.7 Update `TestMassDriftSkipDoesNotStallScheduler` and `TestBlockingDriftSkipCallbackDoesNotBlockOtherTasks` to the cancel-and-merge semantics; verify they still cover mass cancellation, blocking callbacks and Stop waiting
- [ ] 3.8 Existing `TestTaskTriggerWindows` and `TestTaskTriggerWindowsWithBackfill` still pass unchanged

## 4. Verification

- [ ] 4.1 Run `go test -race ./...` (high counts with `-skip /5_days`) and confirm all tests pass
- [ ] 4.2 PR description lists the breaking behaviour changes from the proposal
