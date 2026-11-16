# Proposal

## Why

`OnCancel` currently covers two unrelated events:
- a run that was skipped because it started more than MaxDrift late (scheduler.go:408). The task is still scheduled and will run again;
- a task that was removed (scheduler.go:667) or replaced by a new registration (scheduler.go:245). The task will never run again.

Callers can't tell these apart, yet they need different reactions: alert or record a missed window versus clean up. The skip call also receives no information about which window was missed. It runs synchronously on the main loop, so a slow callback stalls all scheduling.

## What Changes

- New per-task option `WithSkipCallback(f SkipCallback)`, where `SkipCallback func(ctx context.Context, meta TriggerMetadata)`. It is called when a run is skipped for drift. The metadata describes the skipped run: its window, its expected trigger time, and the actual time it would have started.
- The skip callback is called asynchronously, so it doesn't block the main loop.
- **BREAKING**: the cancel callback (`WithCancelCallback`) is no longer called for drift skips. It now fires only when the task is removed or replaced.
- Specs are updated so that "cancelled" means removed or replaced, and "skipped" means a drift skip.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: the maximum-drift requirement invokes the skip callback instead of the cancel callback, and the callbacks requirement documents the skip callback.
- `task-registration`: the replacing and removing requirements state that the cancel callback is reserved for these cases.

## Impact

- Code: `scheduler.go` (`SchedulingOptions.OnSkip`, `schedulingUnit.onSkip`, `WithSkipCallback`, `processTask` skip path); tests in `scheduler_test.go`.
- **BREAKING**: code that relied on `OnCancel` firing on drift skips must move to `WithSkipCallback`.
- Dependencies on other changes: none required.
- Coordination with other changes:
  - `fix-mainloop-self-send-deadlock` reworks the same skip path (the self-send to `newTaskCh`). Whichever lands second rebases onto the other's skip path, but the behaviour stays the same.
  - `tidy-public-api` turns callback pointers into plain funcs. The new field follows whatever style is on main when this change is implemented.
  - `replace-summary-logging-with-hooks` adds a scheduler-wide skip hook. This per-task callback stays alongside it.
