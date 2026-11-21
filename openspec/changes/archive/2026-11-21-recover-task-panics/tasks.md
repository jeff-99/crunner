## 1. Panic error type

- [x] 1.1 Add exported `PanicError` (`Value any`, `Stack []byte`, `Error()`, `Unwrap()`) in `task.go`; verify with a unit test that `errors.As` finds it and `errors.Is` matches a wrapped error value

## 2. Recovery

- [x] 2.1 Add a deferred `recover` in the inner task closure in `processTask` (scheduler.go ~432) that sets `err` to a `*PanicError` with `debug.Stack()`; verify `go vet ./...` and existing tests pass
- [x] 2.2 Test (testing/synctest): an interval task that panics with `"boom"` triggers the error callback with a `*PanicError` whose `Value == "boom"` and non-empty `Stack`, increments `ErrorCount`, does not call the success callback, and the test process does not crash
- [x] 2.3 Test (testing/synctest): a task that panics on every run is run again on subsequent windows, and a second healthy task registered alongside keeps running on schedule
- [x] 2.4 Test: a task panicking with `io.EOF` yields an error for which `errors.Is(err, io.EOF)` is true

## 3. Verification

- [x] 3.1 Run `go test -race ./...` and confirm all tests pass (run with -count=20 -skip /5_days plus a full -count=1)
