# Tasks

## 1. Schema and tooling

- [ ] 1.1 Add `buf.yaml` and `buf.gen.yaml` and verify `buf lint` and `buf generate` run cleanly on an empty module
- [ ] 1.2 Write `proto/crunner/v1/scheduler.proto` from design.md and verify `buf lint` passes
- [ ] 1.3 Add a CI step running `buf breaking` against main and a stale-codegen check, and verify both fail when a field is renumbered or generated code is outdated
- [ ] 1.4 Generate Go code and verify `go build ./...` succeeds with the root package free of a protobuf dependency (`go list -deps .`)

## 2. Sidecar core

- [ ] 2.1 Implement frame reading and writing with varint length and a maximum frame size, and verify unit tests cover valid, oversized, truncated and corrupt frames
- [ ] 2.2 Implement the session loop with id matching, the initialization handshake and error mapping, and verify tests for version mismatch, request before initialize and repeated initialize
- [ ] 2.3 Implement control requests (start, stop, register, remove, list, get) on top of `crunner.New`, and verify tests for replacement, validation errors and unknown tasks
- [ ] 2.4 Implement the task adapter (`RunRequest`, `RunResult`, `CancelRun`, late-result dropping), and verify with `testing/synctest` tests for success, failure, timeout, unresponsive client and non-overlapping runs
- [ ] 2.5 Implement skip and removed notifications gated by `notify_skip` and `notify_removed`, and verify tests for both flags on and off
- [ ] 2.6 Implement `max_concurrent_runs` back-pressure, and verify with `testing/synctest` tests for limit reached and skip after waiting past maximum drift
- [ ] 2.7 Forward scheduler logs as `LogRecord` with level filtering, and verify that stdout contains only valid frames
- [ ] 2.8 Implement `cmd/crunner-sidecar` with stdio wiring, stdin-EOF shutdown, graceful stop and `--version`, and verify with a subprocess test that closing stdin exits with status 0
- [ ] 2.9 Run the sidecar session tests under the race detector (`go test -race ./...`) and verify they pass

## 3. Conformance suite

- [ ] 3.1 Define the data format for conformance cases (clock, registrations, scripted client behaviour, expected trace) and verify it is documented with two example cases
- [ ] 3.2 Write cases covering windows, backfill, drift, timeouts, cancellation, notifications, back-pressure and snapshots, and verify they pass against the sidecar with the Go harness
- [ ] 3.3 Add an in-process parity test asserting the same windows and trigger times as the in-process scheduler for the same schedules, and verify it passes

## 4. TypeScript client

- [ ] 4.1 Generate protobuf-es code and add the package skeleton, and verify the package builds
- [ ] 4.2 Implement the process spawn, framing and request matching, and verify unit tests against a stub sidecar
- [ ] 4.3 Implement the public API (`Scheduler.spawn`, `register`, `start`, `stop`, `remove`, `list`, `get`, callbacks) with `AbortSignal` cancellation, and verify tests for exceptions, async errors and cancellation
- [ ] 4.4 Run the conformance suite through the client and verify all cases pass
- [ ] 4.5 Write the README with a quick-start example and verify the example runs as written
- [ ] 4.6 Set up per-platform npm packages for the sidecar binary and verify an install on a clean machine finds its binary

## 5. Python client

- [ ] 5.1 Generate Python code with type stubs and add the package skeleton, and verify the package builds and `mypy` passes on it
- [ ] 5.2 Implement the process spawn, framing and request matching with asyncio, and verify unit tests against a stub sidecar
- [ ] 5.3 Implement the public API (`Scheduler`, decorator-style registration, callbacks, cancellation via asyncio, optional thread pool for sync tasks), and verify tests for exceptions, cancellation and sync tasks
- [ ] 5.4 Run the conformance suite through the client and verify all cases pass
- [ ] 5.5 Write the README with a quick-start example and verify the example runs as written
- [ ] 5.6 Set up per-platform wheels containing the sidecar binary and verify an install in a clean virtualenv finds its binary

## 6. Release and integration

- [ ] 6.1 Add GoReleaser configuration for Linux, macOS and Windows on amd64 and arm64, and verify a snapshot build produces all binaries with checksums
- [ ] 6.2 Pin each client release to a compatible sidecar version and verify a mismatch is reported at initialization
- [ ] 6.3 Add a top-level document describing the architecture and protocol, and verify its links and commands are correct
- [ ] 6.4 Run both clients and the conformance suite in CI on all supported platforms and verify the pipeline is green

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
