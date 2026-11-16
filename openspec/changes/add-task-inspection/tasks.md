## 1. TaskInfo and snapshot plumbing

- [ ] 1.1 Add exported `TaskInfo` struct (ID, NextRun, WindowStart, WindowEnd, LastRun, LastSuccess, LastError time, Count, SuccessCount, ErrorCount, LastErr, Running) with doc comments; verify `go vet ./...` passes
- [ ] 1.2 Track next run time and running flag in main-loop-owned task state (set when arming the timer / launching a run, cleared on completion or schedule error); verify with a synctest test that inspects a task before, during and after a run
- [ ] 1.3 Add an inspection request channel served by the main loop, with a direct registry read path when idle; verify with a test that inspects on idle and running schedulers

## 2. Public API

- [ ] 2.1 Add `Task(id string) (TaskInfo, bool)` to `Scheduler` and implement it; verify tests for known and unknown IDs
- [ ] 2.2 Change `ScheduledTasks()` to return `[]TaskInfo` sorted by ID and update existing tests in `scheduler_test.go`; verify `go test ./...` passes
- [ ] 2.3 Add a snapshot-isolation test (earlier snapshot unchanged after another run) and a concurrent-inspection test; verify `go test -race ./...` passes
- [ ] 2.4 Document the breaking `ScheduledTasks` change in package docs/README; verify the doc example compiles (example test)
