## Context

`processTask` (`scheduler.go:391-470`) starts a goroutine per due run with no bound. After `fix-shared-state-races`, the main loop owns run state and learns about completions via a message, which is the natural place to account for slots.

## Goals / Non-Goals

**Goals:**
- One scheduler-wide limit with wait/skip modes; zero cost when unset.

**Non-Goals:**
- Per-task concurrency (a task never overlaps itself; see `add-overlap-policy`).
- Priorities or per-tag limits.
- Distributed limits across replicas (`add-distributed-locking`).

## Decisions

- **Slot accounting in the main loop** (an `inFlight int` and a FIFO slice of queued units), not a semaphore channel acquired in task goroutines. The main loop already serialises starts/completions, so this is race-free and lets queued runs be dropped on remove/replace/stop by filtering the queue. Alternative: `chan struct{}` semaphore in each goroutine — rejected: blocked goroutines can't be cleanly discarded and drift can't be re-checked at start.
- **Drift re-checked at dequeue** in wait mode, reusing the existing drift check, so the MaxDrift contract stays "measured at actual start".
- **Backfill runs count** against the limit like any run.
- **Invalid limit (n < 1)**: `New` has no error return; the option panics with a clear message. Alternative: silently treat as unlimited — rejected as it hides misconfiguration. If `validate-task-registration` or a later change gives `New` an error return, switch to that.
- `type LimitMode int` with `LimitWait`, `LimitSkip`.

## Risks / Trade-offs

- [Starvation of frequently due tasks in wait mode] → FIFO by due time; documented.
- [Wait mode plus small MaxDrift causes many skips] → documented; observable through skip callback.
- [Panic on invalid option] → only for programmer error at construction time.
