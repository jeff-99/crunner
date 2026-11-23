# Proposal

## Why

`Stop()` cancels the scheduler context and then blocks on `s.wg.Wait()` (scheduler.go ~626) with no bound. A task that ignores context cancellation makes shutdown hang forever, which is unacceptable for services with a termination grace period (e.g. Kubernetes SIGTERM handling).

## What Changes

- **BREAKING**: `Stop() error` becomes `Stop(ctx context.Context) error`. Stop still cancels the scheduler context and stops all timers immediately; it then waits for running tasks until they return or `ctx` is done.
- If `ctx` is done first, Stop returns an error wrapping `ctx.Err()`; the still-running tasks are abandoned (their contexts are already cancelled), and the scheduler moves to idle only once they have all returned.
- While tasks are still draining after a timed-out Stop, `Start` returns an error and a further `Stop` call waits again (bounded by its own ctx).

## Capabilities

### New Capabilities

### Modified Capabilities
- `scheduler-lifecycle`: "Stopping the scheduler" gains a caller-supplied bound on waiting for running tasks and defines state after a timed-out stop.

## Impact

- Code: `scheduler.go` (`Scheduler` interface, `Stop`, `Start` state check); all `scheduler.Stop()` calls in `scheduler_test.go`.
- API: **BREAKING** — `Scheduler.Stop` signature changes; callers migrate with `Stop(context.Background())` for today's unbounded behaviour.
- Dependencies: none added. No dependency on other changes. Textually overlaps with `fix-shared-state-races` (Stop iterating timers); rebase whichever lands later.
