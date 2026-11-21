## Context

See proposal.md. Today per-task state lives on `*schedulingUnit` and is touched from four kinds of goroutine: the main loop (`scheduleTask`, `processTask`, `emitTaskSummaries`), the `time.AfterFunc` callback (reads `su` only to send it), the per-run task goroutine (writes `meta`, `scheduler.go:440-456`), and API callers (`RegisterTask`/`RemoveTask`/`Stop` stop `su.timer`). `stateMu` only guards the readers, not the writer in `scheduleTask`. The main loop already discards stale units (`mainLoop` checks `currentUnit == receivedUnit`, ~528/545).

## Goals / Non-Goals

**Goals:**
- Single-goroutine ownership of all mutable per-task state, enforced by structure rather than locks.
- `go test -race ./...` clean, including a dedicated stress test.

**Non-Goals:**
- Making Load/Store in `RegisterTask` atomic (`make-task-registration-atomic`).
- Removing the self-send on the drift-skip path (`fix-mainloop-self-send-deadlock`).
- Exposing metadata to callers (`add-task-inspection`).

## Decisions

- **Ownership by the main loop, not a mutex per unit.** A per-unit mutex would fix the race but keep state reachable from many goroutines and invite future races; the main loop already serialises scheduling, so moving the remaining writes there is the smaller conceptual model. Alternative rejected: `sync.Mutex` on `schedulingUnit`.
- **Run completion message.** The task goroutine sends `runResult{unit, start, err}` on a new buffered `doneCh` (select with `schedulerCtx.Done()`), instead of mutating `meta` and sending the unit on `newTaskCh`. The main loop's new case updates `meta` (counts, last run times, `LastError`) and calls `scheduleTask` if the unit is still current. The task still receives an immutable copy of metadata taken on the main loop before the goroutine starts (as `metaCopy` today, but copied before `go`).
- **Callbacks stay in the task goroutine.** Success/error callbacks are invoked by the task goroutine with the run's own error before sending the result; they do not read shared state, so they need no change, and the ordering "callback before next scheduling" is preserved.
- **Timers are only created, replaced and stopped on the main loop.** `RegisterTask`/`RemoveTask` stop touching `timer`; a stale timer that fires sends its unit to `tasksIn`, where the existing identity check drops it. To avoid unbounded stale timers on replace/remove, the main loop stops the old unit's timer when it sees a newer unit for the same ID in `newTaskCh`, and `RemoveTask` is not required to stop it (it fires once and is dropped). On exit (`schedulerCtx.Done()`), `mainLoop` stops every timer it owns and closes a `loopDone` channel; `Stop` cancels the context, waits on `loopDone`, then `wg.Wait()`.
- **Timer callback must not block after stop.** The `AfterFunc` callback already selects on `schedulerCtx.Done()`; keep that.
- **Remove `stateMu`.** Nothing needs it once timers are main-loop only. The `state` atomic stays.

## Risks / Trade-offs

- [A removed task's timer may fire once more and be dropped] → Harmless; covered by the "replaced task's pending timer does not run it" scenario.
- [`doneCh` full when many tasks finish at once] → Buffer like the other channels; the task goroutine blocks only until the main loop drains, and selects on ctx so Stop cannot deadlock. The main loop never sends on `doneCh`.
- [Conflicts with other hardening changes editing the same functions] → Keep this change limited to ownership; no behavioural changes to drift/backfill logic.
