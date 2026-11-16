## Context

`scheduleTask` (`scheduler.go:279-389`) computes `expectedTriggerTime` and arms a timer for `expectedTriggerTime - now`; `processTask` (`scheduler.go:398-399`) skips the run if `now - ExpectedTriggerTime > MaxDrift`. Agreed decision: jitter shifts only the start, never the window, and adds to the allowed drift.

## Goals / Non-Goals

**Goals:**
- Spread starts of runs with identical triggers; no effect on window semantics.

**Non-Goals:**
- Scheduler-wide default jitter, deterministic (hash-based) jitter, negative/early jitter.

## Decisions

- **Apply delay when arming the timer**: `fireAfter = max(0, expected - now) + rand.N(J)` for non-backfill runs. `ExpectedTriggerTime` stays the un-jittered value so metadata and windows are unchanged. Alternative: shift `ExpectedTriggerTime` — rejected, it leaks into metadata and drift semantics.
- **Drift allowance `MaxDrift + J`** rather than subtracting the chosen delay: simpler, needs no per-run state, and a run is never skipped merely because of jitter.
- **`math/rand/v2` `rand.N`**: stdlib, goroutine-safe, no seeding. Tests assert bounds rather than exact values (optionally via an unexported rand hook).
- **Backfill excluded**: catch-up runs exist to close gaps fast; jittering them only delays catch-up.
- **Validation**: negative J → `RegisterTask` error. If `validate-task-registration` lands first, add the check there.

## Risks / Trade-offs

- [J larger than interval delays a run past the next trigger] → runs never overlap (next timer armed after completion); documented that J should be smaller than the interval.
- [Randomness in tests] → bound-based assertions under `testing/synctest`.
