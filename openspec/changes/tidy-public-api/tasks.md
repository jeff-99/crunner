# Tasks

## 1. Schedule constructors

- [ ] 1.1 Rename `ParseCron` to `parseCron` (update `schedule_internal_test.go` too); verify `go doc github.com/jeff-99/crunner` no longer mentions `cron.Schedule`
- [ ] 1.2 Add `Every(d) Schedule`, replacing `WithIntervalSchedule`; update the tests and doc comments; verify the interval tests pass unchanged in behaviour
- [ ] 1.3 Add `CronOption`, `CronLocation` and `CronWithSeconds`, and `Cron(spec, ...CronOption) (Schedule, error)`, replacing `WithCronSchedule` (nil or missing location means `time.Local`); move the tests that read `schedule.Location` or build a zero-value schedule to in-package tests; verify the `schedule_test.go` cases pass and add a test that a 6-field spec without `CronWithSeconds` returns an error

## 2. Options and callbacks

- [ ] 2.1 Change `SchedulingOptions.OnSuccess`, `OnError` and `OnCancel` to plain func types and update the `With*Callback` options and all call sites; verify the callback tests (`TestSchedulerSuccessCallback`, `TestSchedulerErrorCallback`, `TestRemoveTaskCancelCallback`, `TestOverwriteTaskCancelCallback`) pass
- [ ] 2.2 Keep rejecting nil callbacks: each `With*Callback` option records a message on a private `invalid` field of `SchedulingOptions` when given a nil func, and `validateRegistration` reports each as `ErrInvalidOption` (replacing the pointer dereference in `validate.go`); verify the existing "nil success/error/cancel callback" cases in `TestRegisterTaskValidation` still pass unchanged, and that omitting the option is still accepted
- [ ] 2.3 Rename `WithDefaultOptions` to `DefaultSchedulingOptions` and update its doc comment; verify `go build ./...` passes

## 3. Integration

- [ ] 3.1 Update the doc comments on the new and renamed identifiers to the new API (the migration table lives in proposal.md); verify `go vet ./...` and `go test -race ./...` pass (high counts with `-skip /5_days`)
