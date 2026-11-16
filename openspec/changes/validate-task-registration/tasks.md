# Tasks

## 1. Errors and validation helper

- [ ] 1.1 Add exported `ErrInvalidTask`, `ErrInvalidSchedule`, `ErrInvalidOption` with doc comments; verify `go vet ./...` passes
- [ ] 1.2 Add `validate() error` to the interval schedule (rejects interval <= 0) and to the cron schedule (rejects a zero value); verify with table tests in `schedule_test.go`
- [ ] 1.3 Implement the validation helper covering a nil or typed-nil task/schedule, schedule `validate()`, negative MaxDrift, NotAfter <= NotBefore, backfill without NotBefore and nil callbacks, joining all errors; verify with a table test that asserts `errors.Is` for each case and for a combined case

## 2. Wire into RegisterTask

- [ ] 2.1 Call the validation helper in `RegisterTask` right after the options are applied and before the overwrite handling; verify with a test that an invalid registration stores nothing (`ScheduledTasks` length unchanged)
- [ ] 2.2 Add a `testing/synctest` test: register a valid task under ID "a", start the scheduler, attempt an invalid registration under "a", then advance time; verify the original task keeps running and its cancel callback is not called
- [ ] 2.3 Update existing tests that rely on now-invalid input (e.g. backfill without NotBefore), and update the `RegisterTask` doc comment to list the rejected inputs; verify `go test -race ./...` passes
