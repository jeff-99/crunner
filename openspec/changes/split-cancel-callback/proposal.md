# Proposal

## Why

`WithCancelCallback` currently covers two unrelated events:
- a run that was skipped because it started more than MaxDrift late. The task is not gone: the skipped window is merged into the task's next run and the task keeps running;
- a task that was removed (`RemoveTask`) or replaced by a new registration with the same ID. The task will never run again.

Callers can't tell these apart, yet they need different reactions: alert or record a missed run versus clean up. The drift call also receives no information about which window was missed, only a context. Our own specs and tests already had to call a drift skip "cancelled", which shows the overloaded name.

## What Changes

- New per-task option `WithSkipCallback(f SkipCallback)`, where `SkipCallback func(ctx context.Context, meta TriggerMetadata)`. It is called once when a run is skipped for drift. The metadata describes the skipped run: its trigger window, its expected trigger time, and, as the actual trigger time, the moment the skip was decided.
- **BREAKING**: `WithCancelCallback` / `CancelCallback` / `SchedulingOptions.OnCancel` are renamed to `WithRemovedCallback` / `RemovedCallback` / `SchedulingOptions.OnRemoved`. The callback fires only when the task is removed or replaced, and no longer for drift skips. The rename makes the break loud: code that used the cancel callback to notice drift fails to compile and must choose `WithSkipCallback` or `WithRemovedCallback`.
- Both callbacks are called asynchronously (off the main loop) and `Stop` waits for them.
- `WithSkipCallback(nil)` and `WithRemovedCallback(nil)` are rejected by `RegisterTask` with `ErrInvalidOption`, like the other callbacks.
- Specs use "skipped" for a drift skip (its window is merged into the next run) and "removed" for removal or replacement.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-execution`: the "Maximum drift" requirement is replaced by "Drift skipping" (the OpenSpec validator does not allow renaming scenarios inside a modified requirement), which invokes the skip callback instead of the cancel callback; Callbacks documents the skip callback and its metadata.
- `task-registration`: Replacing and Removing a task invoke the removed callback and state that it is reserved for these events; Registration validation and Rejected registration cover the renamed and new callbacks.

## Impact

- Code: `scheduler.go` (`SchedulingOptions.OnSkip`/`OnRemoved`, `schedulingUnit` fields, option functions and types, the drift path in `processTask`, the removal and replacement paths), `validate.go`, tests in `scheduler_test.go`.
- **BREAKING**: `WithCancelCallback`, `CancelCallback` and `OnCancel` no longer exist; use `WithRemovedCallback`, `RemovedCallback` and `OnRemoved` for removal/replacement and `WithSkipCallback` for drift.
- Dependencies on other changes: none outstanding (`fix-mainloop-self-send-deadlock`, `make-trigger-windows-contiguous`, `validate-task-registration` are merged and preserved).
- Coordination with other changes:
  - `tidy-public-api` turns callback pointers into plain funcs. The new fields follow whatever style is on main when this change is implemented, and `tidy-public-api` must also cover the new callbacks.
  - `replace-summary-logging-with-hooks` adds a scheduler-wide skip hook. The per-task skip callback stays alongside it, and both describe the same event.
