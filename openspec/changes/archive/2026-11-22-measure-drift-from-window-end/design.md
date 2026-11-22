## Context

`scheduleWindow` computes `expectedTriggerTime` as the window end, raised to `su.previousFinish` when the previous run finished later. `previousFinish` is recorded by `handleRunResult` from `runResult.finish`, which the task goroutine fills in. The raise was added by `make-trigger-windows-contiguous` so that a long run could not make its successor late. The drift check in `processTask` then compares the actual start with that expected time.

## Goals / Non-Goals

**Goals:**
- Drift is measured from the window end, so an overrun beyond MaxDrift is skipped and merged into the next window (bounded lag, visible through the skip callback).
- Keep everything else of the trigger-window design: contiguous windows, `coveredUntil`, skip-and-merge, restart behaviour.

**Non-Goals:**
- Overlap policies (`add-overlap-policy`), which build on this behaviour as the "wait" default.
- Any change to how the merged window or the next trigger time is computed.

## Decisions

- **Expected trigger time = window end, nothing else.** Remove the `previousFinish` raise. Alternative: keep the raise and add a separate skip for overruns — rejected: it would duplicate the drift mechanism and the skip callback semantics are already "this run did not happen and its window was merged".
- **Delete `previousFinish` and `runResult.finish`.** They are only used by the removed rule, and `handleRunResult` still records the other result fields. Alternative: keep the finish time for later features (inspection) — rejected for now; `add-task-inspection` can add its own.
- **Immediate start when late within MaxDrift.** Unchanged: when the previous run finishes after the next window end but within MaxDrift, the timer fires at once and the run starts immediately (`fireAfter` is clamped to 0 as today).
- **Lag stays bounded by MaxDrift.** Because drift now includes overruns, each run starts at most MaxDrift after its window end; a task that always overruns is skipped repeatedly with merged windows, and each skip invokes the skip callback.

## Risks / Trade-offs

- [Tasks that always overrun now get skip callbacks and merged, wider windows instead of silently lagging] → That is the intent: it is visible and bounded. Documented on `WithMaxDrift` (overruns count as drift; raise MaxDrift for tasks that may overrun).
- [Reverses a decision of `make-trigger-windows-contiguous`] → Called out in the proposal; its other guarantees (contiguity, grid windows, skip-and-merge) are untouched.
- [Existing tests that run longer than their interval with the default 30s MaxDrift] → Lag now accumulates until MaxDrift is exceeded; tests that need many consecutive windows set a large MaxDrift.
