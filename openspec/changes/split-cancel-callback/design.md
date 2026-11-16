# Design

## Context

In `processTask` (scheduler.go:391-424), the drift-skip path does the following:
1. sets `LastRunTime = now`;
2. calls `*su.onCancel` synchronously on the main loop, with a context derived from the scheduler;
3. sends to `newTaskCh`.

The overwrite path (scheduler.go:245-255) and `RemoveTask` (scheduler.go:667-671) also call `onCancel`, but in a goroutine. Before the skip path runs, `ActualTriggerTime` has not been set yet for this run.

## Goals / Non-Goals

**Goals:**
- Give drift skips a dedicated callback with the metadata of the missed window.
- Stop running user code synchronously on the main loop for skips.

**Non-Goals:**
- Changing when a skip happens. That is covered by `make-trigger-windows-contiguous` and `add-overlap-policy`.
- Scheduler-wide hooks. That is `replace-summary-logging-with-hooks`.
- Telling "removed" and "replaced" apart in the cancel callback.

## Decisions

- **Signature: `func(ctx context.Context, meta TriggerMetadata)`.** It reuses the type that tasks already receive, so no new types are needed. Before the copy is taken, `ActualTriggerTime` is set to the skip time so that callers can compute the drift. Considered: a dedicated `SkipInfo{Drift, MaxDrift, ...}`. That can be added later through the hooks change.
- **Asynchronous call.** Take a copy of the metadata on the main loop, then call `go f(ctx, metaCopy)`. Copying on the main loop avoids racing on `su.meta`. The context is the scheduler context, the same as the cancel callback uses.
- **Storage follows the existing callback style.** Use `OnSkip *SkipCallback` in `SchedulingOptions` and `onSkip` in `schedulingUnit`, consistent with the other callbacks on main. `tidy-public-api` converts all of them together.
- **No fallback to `OnCancel`.** A clean split is the whole point. The break is documented in the proposal.

## Risks / Trade-offs

- [Silent behaviour change for existing `OnCancel` users who expected skip notifications] → Marked BREAKING. The changelog and doc comments on both options point to the other one.
- [Many skip callbacks can pile up as goroutines if a task keeps getting skipped] → At most one skip per scheduled run, so the rate is bounded by the schedule.
