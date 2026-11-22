## 1. Expected trigger time

- [ ] 1.1 Add synctest tests with the 10s interval / MaxDrift 5s example (runs 7.5s, 12.5s, 1s, 19s): runs start at 10s, 20s, 32.5s, 40s; the window [40s, 50s) is skipped at 59s with one skip callback; the next run has window [40s, 60s) and starts at 60s; verify the 19s case fails on current main
- [ ] 1.2 In `scheduleWindow`, set the expected trigger time to the window end and remove the `previousFinish` raise; verify the new tests pass
- [ ] 1.3 Remove `schedulingUnit.previousFinish`, `runResult.finish` and their updates (task goroutine, `handleRunResult`); verify `go build ./...` and `go vet ./...`

## 2. Existing tests and docs

- [ ] 2.1 Update `TestLongRunningTaskIsNotCancelled` to the new behaviour (within MaxDrift: runs immediately; beyond MaxDrift: skipped with merged window), and give `TestWindowsStayOnGridWhenRunsStartLate` a MaxDrift large enough for 100 consecutive overrunning runs; verify all drift and window tests pass
- [ ] 2.2 Update the `WithMaxDrift` doc comment (overruns count as drift; raise MaxDrift for tasks that may overrun); verify by reading `go doc`

## 3. Verification

- [ ] 3.1 Run `go test -race ./...` (high counts with `-skip /5_days`) and confirm all tests pass; the PR description states the behavioural break
