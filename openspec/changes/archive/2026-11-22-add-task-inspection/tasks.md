## 1. TaskInfo and snapshot publishing

- [x] 1.1 Add exported `TaskInfo` (ID, NextRunTime, WindowStart, WindowEnd, LastRunTime, LastSuccessfulRunTime, LastErroredRunTime, Count, SuccessfulCount, ErrorCount, LastError, Running) with doc comments, an `atomic.Pointer[TaskInfo]` and `nextRun`/`running` fields on `schedulingUnit`, and a `publish()` method that builds and stores a snapshot from main-loop state; verify `go vet ./...` passes
- [x] 1.2 Publish at registration, when a timer is armed or scheduling fails, when a run is launched, when a run result is handled, and on main loop exit (clearing next run and running); verify with a `testing/synctest` test that inspects a task before it is scheduled, while waiting, during a run, after a success and after Stop

## 2. Public API

- [x] 2.1 Add `Task(id string) (TaskInfo, bool)` to `Scheduler` and implement it by loading the published snapshot; verify tests for known and unknown IDs
- [x] 2.2 Change `ScheduledTasks()` to return `[]TaskInfo` sorted by ID and update existing tests in `scheduler_test.go`; verify `go test ./...` passes
- [x] 2.3 Add tests: errored run (error count, last error, last errored run time), schedule error clears next run, generated IDs are listed, snapshot isolation (an earlier snapshot is unchanged after another run), and concurrent inspection while tasks run, Start and Stop; verify `go test -race ./...` passes
- [x] 2.4 Document `TaskInfo` and the breaking `ScheduledTasks` change in doc comments and add an `Example` test; verify it compiles and passes

## 3. Race-free Start and Stop

- [x] 3.1 Hold the scheduler context and its cancel function in an atomic `runState`, serialise `Start` and `Stop` with a mutex, and capture the context on the main loop (timers, `processTask`, `mainLoop`); verify with a test that hammers Start/Stop from several goroutines while others register and remove tasks, and with the concurrent-inspection test, both under `-race`

## 4. Verification

- [x] 4.1 Run `go test -race ./...` (high counts with `-skip /5_days`) and confirm all tests pass (-count=20 -skip /5_days plus a full -count=1)
