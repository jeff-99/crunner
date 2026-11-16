# Tasks

## 1. Remove summary logging

- [ ] 1.1 Remove `WithSummaryInterval`, the `summaryInterval` field and default, `emitTaskSummaries`, the `lastReported*` fields and the ticker case in `mainLoop`; update any tests using them; verify `go build ./...` and `go test -race ./...` pass and `grep -n -i summary *.go` is empty

## 2. Hooks API

- [ ] 2.1 Add `hooks.go` with `Event`, `Hooks`, `WithHooks` and a `callHook` helper that recovers and logs hook panics; verify a unit test that a scheduler without hooks runs a task normally (scenario "No hooks configured")
- [ ] 2.2 Build `Event` from a metadata snapshot and call BeforeRun before invoking the task and AfterRun (with duration and error) after it, on the task goroutine; verify with `testing/synctest` tests for "Before-run fires once per run", "Event identifies the run", "After-run on success" (2s duration) and "After-run on failure"
- [ ] 2.3 Call OnError before AfterRun on a non-nil error; verify a test asserting order (scenario "Error reported")
- [ ] 2.4 Call OnPanic (not OnError) before AfterRun on a recovered panic, using the recovery from `recover-task-panics`; verify a test (scenario "Panic reported")
- [ ] 2.5 Call OnSkip on the drift-skip path with the scheduler context, without BeforeRun/AfterRun; verify a `testing/synctest` test (scenario "Drift skip reported")
- [ ] 2.6 Verify a test where BeforeRun panics and the task still runs and is rescheduled (scenario "Panicking hook")

## 3. Documentation

- [ ] 3.1 Document hooks (ordering, synchronous/non-blocking contract) in doc comments, add a README example that counts runs via `AfterRun`, and a CHANGELOG entry for the removed `WithSummaryInterval`; verify `go doc github.com/jeff-99/crunner Hooks` shows it
