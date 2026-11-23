# Spec Delta

## Purpose

Defines how the crunner sidecar executable is started, serves one client session, applies back-pressure and shuts down, so that any language client can embed the scheduler as a child process.

## ADDED Requirements

### Requirement: Stdio session
The sidecar SHALL read the session's `ClientMessage` frames from standard input and write `ServerMessage` frames to standard output. It SHALL write nothing else to standard output; diagnostics that cannot be sent as `LogRecord` SHALL go to standard error.

#### Scenario: Clean stdout
- **WHEN** the sidecar runs a session with logging enabled
- **THEN** standard output contains only valid frames

### Requirement: Shutdown on client loss
When standard input reaches end of file, the sidecar SHALL stop the scheduler, cancel delegated runs, and exit with status 0 without waiting for answers from the client.

#### Scenario: Host process exits
- **WHEN** the client process terminates and standard input is closed
- **THEN** the sidecar exits within the scheduler's stop handling and leaves no running tasks

### Requirement: Graceful stop
When the client sends a stop request and then closes the session, the sidecar SHALL wait for delegated runs to finish or be cancelled, as the in-process `Stop` does, before exiting.

#### Scenario: Stop request
- **WHEN** the client sends `StopRequest` while a run is delegated
- **THEN** the run is cancelled with reason `STOPPING` and the response to `StopRequest` is sent once the scheduler has stopped

### Requirement: Concurrency limit towards the client
The sidecar SHALL NOT have more `RunRequest`s outstanding than the `max_concurrent_runs` given at initialization (0 meaning unlimited). Runs that are due while the limit is reached SHALL wait for a free slot, and drift skipping SHALL apply to them as usual.

#### Scenario: Limit reached
- **WHEN** the limit is 1 and two different tasks are due at the same time
- **THEN** the second `RunRequest` is sent only after the first is answered or cancelled

#### Scenario: Waiting past maximum drift
- **WHEN** a run waits for a slot longer than its task's maximum drift
- **THEN** it is skipped instead of being delegated

### Requirement: Version reporting
The sidecar SHALL report its version and supported protocol version in the initialization response and with a `--version` flag.

#### Scenario: Version flag
- **WHEN** the sidecar is run with `--version`
- **THEN** it prints its version and protocol version and exits with status 0

### Requirement: Distribution
Releases SHALL provide sidecar binaries for Linux, macOS and Windows on amd64 and arm64, and each client release SHALL be pinned to a compatible sidecar version.

#### Scenario: Client finds its binary
- **WHEN** a client package is installed on a supported platform
- **THEN** it locates the matching sidecar binary without requiring a Go toolchain or `protoc`
