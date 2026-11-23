# Tasks

## 1. Clock abstraction

- [ ] 1.1 Add `clock.go` with exported `Clock` and `Timer` interfaces and an unexported real clock wrapping `time.Now` and `time.AfterFunc`, with doc comments on the firing contract; verify `go build ./...` and `go vet ./...` pass
- [ ] 1.2 Add `WithClock(Clock) SchedulerOption` (nil ignored) and default the scheduler's clock to the real clock in `New`; verify with a unit test that a scheduler built without the option uses the real clock and `WithClock(nil)` does not panic

## 2. Route time through the clock

- [ ] 2.1 Replace `time.Now()` in `RegisterTask`, `scheduleTask`, the timer callback, `processTask` and the task goroutine with the scheduler clock, and `time.AfterFunc` in `scheduleTask` with `clock.AfterFunc`; change `schedulingUnit.timer` to `Timer`; verify `grep -n 'time.Now\|time.AfterFunc' scheduler.go` returns no scheduling-path hits and the existing test suite passes with `go test -race ./...`
- [ ] 2.2 Add a test with a hand-written fake clock (fixed `Now`, manually fired `AfterFunc`) asserting the first window starts at the fake time and ends one interval later (scenario "Custom clock drives scheduling"); verify it passes under `go test -race`
- [ ] 2.3 Add a test with the fake clock advanced past expected trigger + MaxDrift before firing the timer, asserting the run is skipped (scenario "Custom clock decides drift"); verify it passes
- [ ] 2.4 Add a `testing/synctest` test confirming default-clock behaviour is unchanged for an interval task (scenario "Default clock"); verify it passes

## 3. Documentation

- [ ] 3.1 Document `WithClock` with a short example in the package doc / README; verify `go doc github.com/jeff-99/crunner WithClock` shows it
