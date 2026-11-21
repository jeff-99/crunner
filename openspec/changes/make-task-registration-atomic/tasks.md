## 1. Atomic registry operations

- [x] 1.1 Compute the task ID once in `RegisterTask` and replace Load+Store with `sync.Map.Swap`, cancelling only the returned previous unit; add a test that registers N tasks under one ID from N goroutines (scheduler running, `testing/synctest` for timing) and asserts exactly one runs and N-1 cancel callbacks fire once each; verify it passes with `go test -race`
- [x] 1.2 Replace Load+Delete in `RemoveTask` with `LoadAndDelete`; add a test calling RemoveTask concurrently for one ID asserting exactly one nil error and one cancel callback; verify it passes with `go test -race`

## 2. Verify

- [x] 2.1 Run `go test -race -count=20 ./...` and verify existing overwrite/remove tests and the new tests pass (note: run with `-skip /5_days` at count=20, see fix-shared-state-races; full suite incl. 5-day cases passes at count=1)
