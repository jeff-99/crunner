# Design

## Context

Summaries are produced by `emitTaskSummaries` (scheduler.go ~472) on a ticker in `mainLoop` (~511-523), reading `su.meta` counts that the task goroutine writes (~440-456) — one of the races in review finding 1. Runs happen in `processTask` (~391): the drift check and skip path run on the main loop (~398-420); the task executes in a goroutine (~427-469) that then calls per-task callbacks. Panic recovery is introduced by `recover-task-panics`.

## Goals / Non-Goals

**Goals:**
- A small, stable hook surface sufficient for run counters, duration histograms, error/panic/skip counters.
- Remove the summary ticker and its racy reads.

**Non-Goals:**
- Shipping Prometheus/OpenTelemetry adapters (would add dependencies).
- Per-task hooks (per-task callbacks already exist); registration/removal/replace events (can be added later).
- Changing per-task callback semantics (`split-cancel-callback` owns that).

## Decisions

- **Struct of funcs, not an interface.**
  ```go
  type Event struct {
      TaskID string
      WindowStart, WindowEnd time.Time
      ExpectedTriggerTime, ActualTriggerTime time.Time
      Duration time.Duration
      Err error
  }
  type Hooks struct {
      BeforeRun, AfterRun, OnError, OnPanic, OnSkip func(ctx context.Context, e Event)
  }
  func WithHooks(h Hooks) SchedulerOption
  ```
  Users set only what they need; adding a field later is non-breaking. Alternative: `EventListener` interface (gocron style) — forces implementing every method. Alternative: single `func(Event)` with a `Kind` — loses type-level clarity; rejected.
- **Value `Event`, built from a metadata snapshot.** Hooks never see `*TriggerMetadata`, so they cannot race with the scheduler.
- **Synchronous calls.** BeforeRun/AfterRun/OnError/OnPanic run on the task goroutine; OnSkip runs where the skip is decided (currently the main loop). Synchronous keeps ordering (error → after-run) and avoids unbounded goroutines. Documented: hooks must be fast and non-blocking.
- **Ordering.** BeforeRun → task → (OnError | OnPanic) → AfterRun → per-task callbacks → reschedule.
- **Hook panics recovered** by a small `callHook` helper that logs at error level and continues.
- **Context**: the run context for run hooks; the scheduler context for OnSkip.
- **Removal** of `WithSummaryInterval`, `summaryInterval`, `emitTaskSummaries`, `lastReported*` fields and the ticker case in `mainLoop`.

## Risks / Trade-offs

- [Slow OnSkip hook blocks the main loop] → Documented; `fix-mainloop-self-send-deadlock` may move skip handling off the main loop anyway.
- [Users lose out-of-the-box summaries] → Migration note and a README example of counting via `AfterRun`.
- [Dependency on `recover-task-panics`] → If implemented first, `OnPanic` would have nothing to fire on; ordering is stated in the proposal.

## Migration Plan

Drop `WithSummaryInterval(...)`; add `WithHooks(Hooks{AfterRun: ...})` to aggregate and report.
