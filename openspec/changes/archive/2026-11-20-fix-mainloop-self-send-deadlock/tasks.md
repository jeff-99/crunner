## 1. Reproduce

- [x] 1.1 Add a `testing/synctest` test that registers >100 tasks with small MaxDrift, forces them all to be late at once (e.g. a blocking cancel callback or by advancing fake time past their trigger), and asserts that another task keeps running; verify it hangs/fails on current main

## 2. Fix skip path

- [x] 2.1 Replace the `newTaskCh` send in `processTask`'s skip path with a direct `scheduleTask(su)` call; verify the mass-skip test from 1.1 passes
- [x] 2.2 Invoke the drift-skip cancel callback in a goroutine tracked by `s.wg`, cancelling the derived context after it returns; add `testing/synctest` tests that a blocking skip callback does not delay other tasks and that Stop waits for it; verify they pass

## 3. Verify

- [x] 3.1 Run `go test -race ./...` and verify the existing drift tests and new tests pass (note: the new drift tests pass under `-race` in isolation; the full suite still reports pre-existing races, fixed by `fix-shared-state-races`)
