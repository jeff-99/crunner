## 1. Reproduce

- [x] 1.1 Add a stress test that concurrently registers, replaces and removes tasks and calls Stop while short tasks run with a tiny summary interval; verify it reports a data race with `go test -race -run Stress` on current main
- [x] 1.2 Fix the data races in the existing test helpers (`TestTriggerMetadataInTask`, `TestTaskTriggerWindows`, `TestTaskTriggerWindowsWithBackfill`: metadata recorded by the task goroutine and read by the test body without synchronisation) using atomics or a mutex; verify these tests report no races that originate in test code

## 2. Main loop owns run results

- [x] 2.1 Copy trigger metadata on the main loop before starting the task goroutine and pass the copy to the task; verify existing metadata tests still pass
- [x] 2.2 Introduce a run-result message and `doneCh`; make the task goroutine send it instead of writing `meta` and sending on `newTaskCh`, and handle it in `mainLoop` (update counts/times/`LastError`, reschedule if still current); verify callbacks/count tests pass and the summary part of the stress test is race-free

## 3. Main loop owns timers

- [x] 3.1 Remove timer access from `RegisterTask` and `RemoveTask`; stop a superseded unit's timer on the main loop; add a `testing/synctest` test that a replaced/removed task never runs when its old timer fires; verify it passes
- [x] 3.2 Make `mainLoop` stop all owned timers on exit and signal completion; make `Stop` wait for it before `wg.Wait()`; remove `stateMu`; verify Stop/restart tests pass

## 4. Verify

- [x] 4.1 Run `go test -race -count=20 ./...` and verify no races or failures, including the stress test from 1.1 (note: run with `-skip /5_days` for count=20, as the 5-day cases exhaust memory under the race detector at high counts; the full suite including them passed at `-count=3`, and the stress test at `-count=50`)
