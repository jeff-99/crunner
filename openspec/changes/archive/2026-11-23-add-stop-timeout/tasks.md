# Tasks

## 1. Bounded Stop

- [x] 1.1 Change `Scheduler.Stop` to `Stop(ctx context.Context) error`, run `wg.Wait()` in a goroutine closing a stored `done` channel that also stores `stateIdle`, and `select` on `done`/`ctx.Done()`; update all `scheduler.Stop()` calls in `scheduler_test.go` to `Stop(context.Background())`; verify `go test -race ./...` passes
- [x] 1.2 Make Stop in `stateStopping` wait on the existing `done` channel bounded by its ctx, and treat a nil ctx as `context.Background()`; verify with a unit test that a second Stop during draining returns once the task returns
- [x] 1.3 Add a `testing/synctest` test: task ignores cancellation, Stop with a 1s timeout returns `errors.Is(err, context.DeadlineExceeded)` after 1s (scenario "Stop bounded by context"); verify it passes
- [x] 1.4 Add a `testing/synctest` test: after a timed-out Stop, Start returns an error while the task runs, and succeeds after it returns (scenarios "Start while still draining", "Idle after draining"); verify it passes

## 2. Documentation

- [x] 2.1 Document `Stop(ctx)` semantics and the migration (`context.Background()` for old behaviour) in the doc comment (the repo has no README/CHANGELOG yet); verify `go doc github.com/jeff-99/crunner Scheduler` shows it
