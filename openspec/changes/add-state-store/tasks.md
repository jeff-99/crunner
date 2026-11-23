# Tasks

## 1. API

- [ ] 1.1 Add `state.go` with `State`, `StateStore`, a mutex-protected in-memory implementation returned by `NewMemoryStateStore`, and godoc stating at-least-once semantics; verify with a unit test for Load-missing / Save / Load-present and `go test -race`
- [ ] 1.2 Add `WithStateStore` scheduler option and default the scheduler to the in-memory store in `New`; verify with a test that a custom store receives calls and that all existing tests still pass unchanged

## 2. Loading state

- [ ] 2.1 Load a task's state off the main loop before its first window is computed (from `Start` and `RegisterTask`) and pass it with the unit; verify with a synctest test that a task with saved `WindowEnd` E has a first window starting at E
- [ ] 2.2 With backfill, start from `max(NotBefore, WindowEnd)`; verify with a synctest test that windows before the saved end are not replayed and that a saved end before NotBefore is ignored
- [ ] 2.3 On load error, log and do not schedule the task; verify with a synctest test using a failing store that the task never runs

## 3. Saving state

- [ ] 3.1 Save `{WindowStart, WindowEnd}` in the task goroutine after success and after error, before rescheduling; verify with a synctest test that the store holds the last window after each run, for both outcomes
- [ ] 3.2 Save state on the drift-skip path (off the main loop); verify with a synctest test that a skipped window is recorded
- [ ] 3.3 On save error, log and continue scheduling; verify with a synctest test using a store whose Save fails that the task keeps running on schedule
- [ ] 3.4 Simulate a restart (stop scheduler, new scheduler sharing the same store, re-register) and verify with a synctest test that coverage is contiguous across the restart

## 4. Docs

- [ ] 4.1 Document state persistence and the "handled" semantics in the README/package doc; verify `go vet ./...` and `go doc` render the new symbols
