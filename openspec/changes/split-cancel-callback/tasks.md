# Tasks

## 1. Skip callback

- [ ] 1.1 Add the `SkipCallback` type, the `SchedulingOptions.OnSkip` field, the `schedulingUnit.onSkip` field and the `WithSkipCallback` option with a doc comment; verify `go vet ./...` passes
- [ ] 1.2 On the drift-skip path in `processTask`, set `ActualTriggerTime`, copy the metadata and call `onSkip` in a goroutine; remove the `onCancel` call there; verify with a `testing/synctest` test that forces drift: the skip callback receives the expected window/trigger times and the cancel callback is not called
- [ ] 1.3 Add a `testing/synctest` test where the skip callback blocks; verify that a second task still runs on schedule

## 2. Cancel semantics and docs

- [ ] 2.1 Update the `WithCancelCallback` doc comment to say "removed or replaced only"; verify that the existing `TestRemoveTaskCancelCallback` and `TestOverwriteTaskCancelCallback` still pass
- [ ] 2.2 Add a test where replacing a task that has a skip callback does not call that skip callback; verify `go test -race ./...` passes
