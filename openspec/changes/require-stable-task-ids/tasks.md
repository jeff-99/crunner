## 1. Stable ID assignment

- [ ] 1.1 Remove random ID generation from `WithDefaultOptions()` (scheduler.go ~57); verify it returns an empty `ID`
- [ ] 1.2 In `RegisterTask`, after applying options, set `opts.ID = "task-" + rand.Text()` when it is empty; verify by test that the stored map key starts with `task-`
- [ ] 1.3 Make `schedulingUnit.ID()` return `su.options.ID` without generating; verify `go vet ./...` passes

## 2. Tests

- [ ] 2.1 Test (testing/synctest): a task registered with `WithID("")` on a running scheduler runs on its interval schedule (fails on main, passes after fix)
- [ ] 2.2 Test: a task registered without `WithID` appears once in `ScheduledTasks()` and can be removed using the key it was stored under (read via the internal map in an in-package test)
- [ ] 2.3 Test (testing/synctest): two tasks registered without `WithID` both run and do not replace each other
- [ ] 2.4 Run `go test -race ./...` and confirm all tests pass
