## 1. API

- [ ] 1.1 Add `RunNow(id string) error` to `Scheduler`, `Manual bool` to `TriggerMetadata`, and exported errors `ErrTaskNotFound`, `ErrTaskRunning`, `ErrSchedulerNotRunning`; verify `go build ./...` and `go vet ./...` pass

## 2. Main loop

- [ ] 2.1 Add per-unit `running` flag managed by the main loop and park scheduled triggers that arrive while running; verify with a synctest test that a scheduled trigger during a manual run starts after it returns, and is skipped when that is beyond MaxDrift
- [ ] 2.2 Implement the `runNowCh` request handling, building the manual window from the last completed scheduled window end; verify with a synctest test that the manual window is 10:00-10:20 and the next scheduled window stays 10:00-11:00
- [ ] 2.3 Keep the value passed to `Schedule.Next` independent of manual runs; verify with a synctest test that an interval task's scheduled trigger times are unchanged by a manual run
- [ ] 2.4 Return errors for unknown ID, running task and idle scheduler; verify with unit tests for each case under `go test -race ./...`
- [ ] 2.5 Verify callbacks and counts for manual runs (success and error) with a test, and document `RunNow` and the manual window in its doc comment
