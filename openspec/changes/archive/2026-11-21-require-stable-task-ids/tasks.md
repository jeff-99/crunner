## 1. Stable ID assignment

- [x] 1.1 Remove random ID generation from `WithDefaultOptions()` (scheduler.go ~57); verify it returns an empty `ID`
- [x] 1.2 In `RegisterTask`, after applying options, set `opts.ID = "task-" + rand.Text()` when it is empty; verify by test that the stored map key starts with `task-`
- [x] 1.3 Make `schedulingUnit.ID()` return `su.options.ID` without generating; verify `go vet ./...` passes

## 2. Tests

- [x] 2.1 Test (testing/synctest): a task registered with `WithID("")` on a running scheduler runs on its interval schedule (fails on main, passes after fix)
- [x] 2.2 Test: a task registered without `WithID` appears once in `ScheduledTasks()` and can be removed using the key it was stored under (read via the internal map in an in-package test)
- [x] 2.3 Test (testing/synctest): two tasks registered without `WithID` both run and do not replace each other
- [x] 2.4 Run `go test -race ./...` and confirm all tests pass (-count=20 -skip /5_days plus a full -count=1)
