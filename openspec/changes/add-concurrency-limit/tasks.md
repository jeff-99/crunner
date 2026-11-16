## 1. Option

- [ ] 1.1 Add `LimitMode` (`LimitWait`, `LimitSkip`) and `WithConcurrencyLimit(n, mode)`, panicking for n < 1; verify with a unit test for valid and invalid values

## 2. Slot accounting

- [ ] 2.1 Track in-flight runs in the main loop and start runs only when a slot is free; verify with a synctest test (limit 2, five due tasks → max two concurrent, measured with an atomic counter)
- [ ] 2.2 Implement wait mode FIFO queue with drift re-check at dequeue; verify synctest tests for "starts when slot frees" and "queued too long → drift skip + skip callback + next window scheduled"
- [ ] 2.3 Implement skip mode; verify a synctest test that a run due while the slot is busy is skipped with the skip callback and rescheduled
- [ ] 2.4 Drop queued runs on Stop and on remove/replace; verify tests that those runs never execute and Stop returns
- [ ] 2.5 Verify no-limit behaviour is unchanged: existing `scheduler_test.go` passes with `go test -race ./...`
- [ ] 2.6 Document the option, modes and MaxDrift interplay in doc comments; verify `go doc` output
