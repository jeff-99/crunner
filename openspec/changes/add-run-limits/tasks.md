## 1. Options and validation

- [ ] 1.1 Add `MaxRuns` to `SchedulingOptions` and `WithMaxRuns(n)`; validate NotAfter > NotBefore and MaxRuns > 0 in `RegisterTask` before any state change; verify with unit tests that invalid registrations return an error and leave an existing same-ID task running

## 2. Limits

- [ ] 2.1 Add the completion helper (main-loop `CompareAndDelete`, no cancel callback); verify via the tests below that the cancel callback is not called
- [ ] 2.2 Enforce NotAfter in `scheduleTask` against the computed window end (inclusive); verify with synctest tests for an interval task stopping at NotAfter, the exact-boundary case, and backfill bounded by NotAfter
- [ ] 2.3 Enforce MaxRuns after each executed run; verify with synctest tests for exactly N runs and for drift-skipped runs not counting
- [ ] 2.4 Update doc comments of `WithNotAfter` and `WithMaxRuns`; verify `go test -race ./...` passes
