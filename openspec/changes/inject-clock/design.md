# Design

## Context

Time is read via `time.Now()` at registration (scheduler.go ~219), in `scheduleTask` (~292), in the timer callback (~372), in `processTask` (~396) and after a run (~438). Run timers are created with `time.AfterFunc` (~371) and stored in `schedulingUnit.timer *time.Timer`, which `RegisterTask`, `RemoveTask` and `Stop` call `Stop()` on. Contexts use `context.WithTimeout` (~286, ~433), and the summary loop uses `time.NewTicker` (~511). `Start` sets `schedulingEpoch` (~573), which is unused.

## Goals / Non-Goals

**Goals:**
- One seam for "now" and for scheduling run timers, so applications can drive the scheduler with a fake clock.
- Zero behaviour change with the default clock.

**Non-Goals:**
- Shipping a fake clock implementation in the public API (tests in this repo use `testing/synctest` instead).
- Controlling context deadlines (`context.WithTimeout`) through the clock; those stay on real time.
- The summary ticker: it is removed by `replace-summary-logging-with-hooks`; until then it keeps using `time.NewTicker`.

## Decisions

- **Minimal interface.**
  ```go
  type Clock interface {
      Now() time.Time
      AfterFunc(d time.Duration, f func()) Timer
  }
  type Timer interface { Stop() bool }
  ```
  Only what the scheduler uses. Alternative: mirror a full clock library (tickers, `Sleep`, `After`) — rejected as unnecessary surface; methods can be added later before v1 if needed.
- **`Timer` interface instead of `*time.Timer`.** A fake clock cannot construct a `*time.Timer` with custom firing. The real clock wraps `time.AfterFunc` and returns the `*time.Timer` directly (it already satisfies `Stop() bool`).
- **Option on the scheduler, not per task.** One time source per scheduler; per-task clocks would make windows of different tasks incomparable.
- **Nil clock** passed to `WithClock` is ignored (keeps the real clock) rather than panicking later in the main loop.
- **Context deadlines stay real.** Faking them would require a custom `context.Context` implementation; `testing/synctest` already fakes them in tests.

## Risks / Trade-offs

- [A user clock whose `AfterFunc` never fires stalls the task] → Documented on `Clock`: `AfterFunc` must call `f` on its own goroutine once `d` has elapsed on that clock.
- [Mixed time sources: run context deadlines use real time while windows use the clock] → Documented as a non-goal; acceptable since deadlines are coarse (30 days by default).
- [Textual conflicts with other in-flight changes touching the same lines] → Change is mechanical; rebase whichever lands later.
