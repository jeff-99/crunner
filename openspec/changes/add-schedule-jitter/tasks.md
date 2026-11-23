## 1. Option

- [ ] 1.1 Add `Jitter time.Duration` to `SchedulingOptions` and `WithJitter(d)`; make `RegisterTask` return an error for negative values; verify with a unit test

## 2. Behaviour

- [ ] 2.1 Add `rand.N(J)` delay to `fireAfter` in `scheduleTask` for non-backfill runs, leaving `ExpectedTriggerTime` and the window untouched; verify with a synctest test that starts fall within [trigger, trigger+J) and metadata windows equal the un-jittered ones
- [ ] 2.2 Extend the drift check in `processTask` to `MaxDrift + Jitter`; verify with a synctest test (MaxDrift 1s, jitter 10s → no skip) and that excessive drift beyond the sum still skips
- [ ] 2.3 Verify backfill catch-up runs are not delayed with a synctest test
- [ ] 2.4 Document `WithJitter` (start-only shift, drift allowance, keep J below the interval); verify `go doc` output and `go test -race ./...` passes
