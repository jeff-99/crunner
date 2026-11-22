# Tasks

## 1. Skip callback

- [x] 1.1 Add `SkipCallback`, `SchedulingOptions.OnSkip`, `schedulingUnit.onSkip` and `WithSkipCallback` with a doc comment (called once per drift skip, async, never for backfilling tasks); add the nil check for it in `validate.go`; verify `go vet ./...` passes
- [x] 1.2 On the drift path in `processTask`, copy the metadata before `scheduleWindow` overwrites it, set `ActualTriggerTime` to the skip time, and start `onSkip` in a goroutine tracked by `s.wg` with the derived context; stop calling the removal callback there; verify with a `testing/synctest` test that forces drift: the skip callback receives the skipped window, expected trigger time and skip time, and the removed callback is not called
- [x] 1.3 Test: a blocking skip callback does not delay another task, and `Stop` waits for it

## 2. Removed callback rename and semantics

- [x] 2.1 Rename `WithCancelCallback`, `CancelCallback`, `SchedulingOptions.OnCancel` and `schedulingUnit.onCancel` to `WithRemovedCallback`, `RemovedCallback`, `OnRemoved` and `onRemoved`; update `validate.go`, doc comments ("called only when the task is removed or replaced") and all tests; verify `TestRemoveTaskCancelCallback`, `TestOverwriteTaskCancelCallback` and the atomic-registration tests pass under the new names
- [x] 2.2 Test: replacing or removing a task that has a skip callback does not call it; skipping a run for a task that has a removed callback does not call that
- [x] 2.3 Update the tests that used the cancel callback as the drift signal (late-run, restart, long-running, mass and blocking drift tests) to use `WithSkipCallback`, keeping their assertions that the removed callback is not called; add `WithSkipCallback(nil)` and `WithRemovedCallback(nil)` cases to `TestRegisterTaskValidation`

## 3. Verification

- [x] 3.1 Run `go test -race ./...` (high counts with `-skip /5_days`) and confirm all tests pass; the PR description lists the rename mapping and the behaviour change (-count=20 -skip /5_days plus a full -count=1)
