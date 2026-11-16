## Context

See proposal.md. `mainLoop` handles `tasksIn` by calling `processTask` inline (`scheduler.go:550`). The non-skip path hands off to a goroutine that later sends on `newTaskCh` from outside the main loop, which is fine. Only the skip path (`scheduler.go:396-419`) sends from the main loop itself and runs the cancel callback inline. `RegisterTask` and `Start` also send on `newTaskCh` but from other goroutines.

## Goals / Non-Goals

**Goals:**
- The main loop never sends on a channel it is the only reader of.
- No user callback runs on the main loop.

**Non-Goals:**
- Changing the drift semantics or which window the next run covers (`make-trigger-windows-contiguous`).
- Separating "skipped" from "cancelled" callbacks (`split-cancel-callback`).

## Decisions

- **Call `scheduleTask(su)` inline on the skip path.** The main loop is already the goroutine that runs `scheduleTask` for `newTaskCh` messages, and the unit is known to be current (checked just before `processTask`), so the identity re-check that the channel round-trip provided is unnecessary. Alternatives rejected: a larger buffer (only moves the threshold); a non-blocking send with fallback goroutine (reorders messages and adds complexity).
- **Run the cancel callback in a goroutine counted in `s.wg`** with the run's derived context, cancelling that context when the callback returns. Counting it in `wg` keeps the existing Stop guarantee ("wait for running tasks") meaningful for callbacks too. Alternative of fire-and-forget `go f(ctx)` (as replace/remove do today) rejected because it lets callbacks outlive Stop.
- **Keep `LastRunTime = now` on skip** exactly as today, so the next window computation is unchanged by this fix.

## Risks / Trade-offs

- [Callback now concurrent with the next run of the same task] → Documented in proposal Impact; matches how replace/remove already invoke it.
- [Stop now also waits for a blocking skip callback] → Callback receives a context cancelled by Stop; documented in the spec scenario.
