# Design

## Context

`Stop()` (scheduler.go ~599-632) CASes running→stopping, cancels the scheduler context, stops timers under `stateMu`, calls `s.wg.Wait()` and stores idle. `Start` (~562) only accepts idle. Tasks run in goroutines tracked by `s.wg` (~426).

## Goals / Non-Goals

**Goals:**
- Bound the time Stop blocks, using idiomatic `context.Context` (same shape as `http.Server.Shutdown`).
- Keep the scheduler's state consistent when Stop gives up early.

**Non-Goals:**
- Forcibly killing goroutines (impossible in Go).
- A scheduler-wide default stop timeout option; the caller's context is enough.

## Decisions

- **`Stop(ctx)` instead of a `WithStopTimeout` option.** The deadline is a property of the shutdown call (signal handler grace period), not of the scheduler. Alternative: option + unchanged signature — rejected because it can't be adapted per call and hides the timeout from the caller. Breaking, acceptable pre-v1.
- **Waiting.** Run `wg.Wait()` in a goroutine that closes a `done` channel, then `select` on `done` and `ctx.Done()`. On timeout return `fmt.Errorf("crunner: stop: %w", ctx.Err())`.
- **State after timeout.** Stay in `stateStopping`; the waiter goroutine stores `stateIdle` when `done` closes. Thus `Start` (which requires idle) fails while draining, satisfying "Start while still draining". A second `Stop` in `stateStopping` waits on the same `done` channel (stored on the scheduler) bounded by its own ctx, instead of returning nil immediately as today — so callers can retry.
- **Nil ctx** is treated as `context.Background()`.

## Risks / Trade-offs

- [Abandoned tasks keep running after Stop returns and may touch shared resources] → Documented; their context is cancelled and the error tells the caller.
- [Behaviour change for a second Stop during stopping: it now waits instead of returning nil] → Only affects concurrent/repeated Stop callers; stated in the spec scenario set.

## Migration Plan

Replace `s.Stop()` with `s.Stop(ctx)`; `context.Background()` preserves the old unbounded wait.
