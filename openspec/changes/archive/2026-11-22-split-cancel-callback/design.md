# Design

## Context

In `processTask`, the drift path (a run that would start more than MaxDrift after its expected trigger time) already does the following on the main loop:
1. logs the skip;
2. starts `*su.onCancel` in a goroutine tracked by `s.wg`, with a context derived from the scheduler context that is cancelled when the callback returns;
3. computes the next, merged window with `scheduleWindow(su, &cancelledEnd)`. That call overwrites `su.meta` (window start/end, expected trigger time).

`RegisterTask` (replacement) and `RemoveTask` also call `onCancel`, in a goroutine, with the scheduler context (or `context.Background()` when the scheduler never started). `validate.go` rejects a nil func inside an `OnSuccess`/`OnError`/`OnCancel` option.

## Goals / Non-Goals

**Goals:**
- Give drift skips a dedicated callback with the metadata of the skipped run.
- Reserve the other callback for removal and replacement, and name it accordingly.
- Keep user code off the main loop and keep `Stop` waiting for it.

**Non-Goals:**
- Changing when a skip happens or how the next window is computed (`make-trigger-windows-contiguous`).
- Scheduler-wide hooks (`replace-summary-logging-with-hooks`).
- Telling removal and replacement apart, or passing a reason (can be added later).
- Scheduler-level start/stop callbacks. A scheduler's `Start` and `Stop` are synchronous calls whose caller already knows when they return; lifecycle hooks can be added to the hooks change later if needed.
- Passing the merged window to the skip callback.

## Decisions

- **Two names, no fallback.** `WithSkipCallback` for drift; `WithRemovedCallback` (renamed from `WithCancelCallback`) for removal and replacement. The rename turns a silent behaviour change into a compile error. No deprecated alias, since the project is pre-1.0. "Removed" covers both `RemoveTask` and replacement by a new registration: in both cases the task will never run again.
- **Signature: `SkipCallback func(ctx context.Context, meta TriggerMetadata)`.** It reuses the metadata type tasks already receive. The copy is taken on the main loop **before** `scheduleWindow` rewrites `su.meta`, with `ActualTriggerTime` set to the moment the skip was decided so callers can compute the drift. Only the skipped run's own window is reported; the merged window that covers it is not part of the callback. `RemovedCallback` stays `func(ctx context.Context)`.
- **Asynchronous, tracked, same context as today.** The skip callback runs in a goroutine counted in `s.wg`, with the scheduler-derived context cancelled when it returns, so `Stop` waits for it and a blocking callback never delays other tasks. The removed callback keeps its current behaviour (a goroutine with the scheduler context).
- **Backfilling tasks never skip,** so their skip callback is never called; documented on the option.
- **Storage follows the existing callback style.** `OnSkip` and `OnRemoved` fields in `SchedulingOptions`, `onSkip` and `onRemoved` in `schedulingUnit`, pointer-to-func like the other callbacks on main; `tidy-public-api` converts all of them together. Validation treats the new callbacks exactly like the existing ones.

## Risks / Trade-offs

- [Existing `WithCancelCallback` users must update] → The rename makes every affected call site fail to compile; the PR lists the mapping.
- [Many skip callbacks can pile up as goroutines if a task keeps getting skipped] → At most one skip per scheduled run, so the rate is bounded by the schedule.
- [Skip callback runs concurrently with the task's next run] → Documented; it already behaves like this for the cancel callback today.
