# Design

## Context

The scheduler calls back into user code: `Task` is a function, `Schedule.Next` is an interface, and there are success, error, skip and removed callbacks. Sharing it across languages therefore needs calls in both directions. See proposal.md for motivation. The `Task` signature is `func(ctx, TriggerMetadata) error`, and the existing "Non-overlapping runs" requirement already guarantees at most one run per task at a time, which the protocol inherits.

## Goals / Non-Goals

**Goals:**
- One schema, many generated clients; the Go core is the only implementation of scheduling logic.
- Clients are small and idiomatic: user code never sees the protocol.
- Embedded use: the client spawns the sidecar, no daemon to manage.
- Transport-independent messages, so a gRPC/daemon mode can be added later without schema changes.

**Non-Goals:**
- Custom `Schedule` implementations over the protocol (reserved for a later `next_time` callback).
- Daemon mode, multiple clients per scheduler, durable state, opaque task payloads.
- Changing the in-process Go API.

## Decisions

### Sidecar process instead of cgo, WASM, or native ports
The sidecar keeps the Go scheduler as the only implementation. cgo `c-shared` makes Go goroutines call into Python/Node (GIL, thread-safe functions), allows one Go runtime per process and breaks on fork. WASM handles goroutines and timers poorly. Native ports in each language would need a shared conformance suite to avoid drift and triple maintenance. A sidecar costs a shipped binary per platform, which is an accepted trade-off (the esbuild/ruff model).

### Protobuf over a single bidirectional session
gRPC methods only go client to server, but `run` goes the other way. So the whole conversation is two `oneof` envelopes (`ClientMessage`, `ServerMessage`) on one stream. Client requests are matched by `id`; scheduler-initiated runs by `run_id`.
Alternatives: JSON-RPC (no static types or codegen), Cap'n Proto/FlatBuffers (weaker TS/Python support, efficiency not needed), ConnectRPC (immature in Python).

### Transport: varint-delimited frames over stdio first
The client starts the sidecar and talks over stdin/stdout. Lifecycle is free (stdin EOF means shutdown), there are no ports, and each app gets its own scheduler. Only message codegen is needed, not a gRPC stack, which keeps the Python package light. The identical messages can later be served as `rpc Session(stream ClientMessage) returns (stream ServerMessage)` for daemon mode.
Framing is the protobuf standard delimited format (`protodelim` in Go, `sizeDelimited` in protobuf-es, a few lines in Python), with a maximum frame size (default 16 MiB).

### The Go side owns time
The sidecar enforces timeouts, drift and cancellation. `deadline` in `RunRequest` is only a hint for the client. If a client ignores `CancelRun`, the run is still treated as finished, and a late `RunResult` is dropped. This keeps scheduling deterministic regardless of client behaviour.

### Callbacks are client-side
`OnSuccess` and `OnError` need no messages because the client already sees the outcome of its own runs. `OnSkip` and `OnRemoved` become `TaskSkipped` / `TaskRemoved` notifications, sent only if the task asked for them (`notify_skip`, `notify_removed`).

### The sidecar's Task adapter
For each registration the sidecar builds a `crunner.Task` that sends a `RunRequest` and blocks on the answer or `ctx.Done()`. On `ctx.Done()` it sends `CancelRun` and returns the context error. This reuses all existing timeout, panic, window and drift logic. Client exceptions arrive as `TaskError` and are turned into Go errors, so counts and `LastError` behave as in-process.

### Back-pressure
`max_concurrent_runs` is enforced inside the adapter with a semaphore. Waiting for a slot happens before the `RunRequest` is sent, so a saturated client leads to normal drift skipping instead of an unbounded queue.

### Schema

```proto
syntax = "proto3";
package crunner.v1;

import "google/protobuf/duration.proto";
import "google/protobuf/timestamp.proto";

message ClientMessage {
  uint64 id = 1;                         // 0 for notifications
  oneof body {
    InitializeRequest   initialize    = 10;
    StartRequest        start         = 11;
    StopRequest         stop          = 12;
    RegisterTaskRequest register_task = 13;
    RemoveTaskRequest   remove_task   = 14;
    ListTasksRequest    list_tasks    = 15;
    GetTaskRequest      get_task      = 16;
    RunResult           run_result    = 20;   // answer to ServerMessage.run
  }
}

message ServerMessage {
  uint64 id = 1;                         // matches ClientMessage.id for responses
  oneof body {
    InitializeResponse initialize = 10;
    Ack                ack        = 11;
    Error              error      = 12;
    ListTasksResponse  list_tasks = 13;
    TaskInfo           task       = 14;
    RunRequest         run        = 20;
    CancelRun          cancel_run = 21;
    TaskSkipped        skipped    = 22;
    TaskRemoved        removed    = 23;
    LogRecord          log        = 24;
  }
}

message InitializeRequest {
  uint32 protocol_version = 1;
  google.protobuf.Duration task_timeout     = 2;
  google.protobuf.Duration schedule_timeout = 3;
  google.protobuf.Duration summary_interval = 4;
  uint32 queue_size          = 5;
  uint32 max_concurrent_runs = 6;
  LogLevel log_level         = 7;
}
message InitializeResponse { uint32 protocol_version = 1; string server_version = 2; }
message StartRequest {}
message StopRequest {}
message Ack {}

message Schedule {
  oneof kind {
    CronSchedule cron = 1;
    google.protobuf.Duration interval = 2;
  }
}
message CronSchedule { string spec = 1; string time_zone = 2; bool include_seconds = 3; }

message RegisterTaskRequest {
  string   task_id  = 1;
  Schedule schedule = 2;
  google.protobuf.Duration  max_drift  = 3;
  google.protobuf.Timestamp not_before = 4;
  google.protobuf.Timestamp not_after  = 5;
  bool backfill = 6;
  google.protobuf.Duration timeout = 7;
  bool notify_skip    = 8;
  bool notify_removed = 9;
}
message RemoveTaskRequest { string task_id = 1; }
message GetTaskRequest    { string task_id = 1; }
message ListTasksRequest  {}
message ListTasksResponse { repeated TaskInfo tasks = 1; }

message TaskInfo {                       // mirrors crunner.TaskInfo; unset time = not applicable
  string task_id = 1;
  google.protobuf.Timestamp next_run_time = 2;
  google.protobuf.Timestamp window_start = 3;
  google.protobuf.Timestamp window_end = 4;
  google.protobuf.Timestamp last_run_time = 5;
  google.protobuf.Timestamp last_successful_run_time = 6;
  google.protobuf.Timestamp last_errored_run_time = 7;
  uint32 count = 8;
  uint32 successful_count = 9;
  uint32 error_count = 10;
  string last_error = 11;
  bool running = 12;
}

message TriggerMetadata {                // mirrors crunner.TriggerMetadata
  google.protobuf.Timestamp expected_trigger_time = 1;
  google.protobuf.Timestamp actual_trigger_time = 2;
  google.protobuf.Timestamp window_start = 3;
  google.protobuf.Timestamp window_end = 4;
  google.protobuf.Timestamp last_successful_run_time = 5;
  google.protobuf.Timestamp last_errored_run_time = 6;
  google.protobuf.Timestamp last_run_time = 7;
  uint32 count = 8;
  uint32 successful_count = 9;
  uint32 error_count = 10;
  string last_error = 11;
}

message RunRequest {
  uint64 run_id = 1;
  string task_id = 2;
  TriggerMetadata meta = 3;
  google.protobuf.Timestamp deadline = 4;
}
message RunResult {
  uint64 run_id = 1;
  oneof outcome { Success ok = 2; TaskError err = 3; }
}
message Success {}
message TaskError { TaskErrorKind kind = 1; string message = 2; string stack = 3; }
enum TaskErrorKind { TASK_ERROR_KIND_UNSPECIFIED = 0; RETURNED = 1; EXCEPTION = 2; CANCELLED = 3; }

message CancelRun { uint64 run_id = 1; CancelReason reason = 2; }
enum CancelReason { CANCEL_REASON_UNSPECIFIED = 0; TIMEOUT = 1; STOPPING = 2; REMOVED = 3; }

message TaskSkipped { string task_id = 1; TriggerMetadata meta = 2; }
message TaskRemoved { string task_id = 1; }

message Error { ErrorCode code = 1; string message = 2; }
enum ErrorCode { ERROR_CODE_UNSPECIFIED = 0; INVALID_ARGUMENT = 1; ALREADY_EXISTS = 2; NOT_FOUND = 3; INVALID_STATE = 4; }
enum LogLevel  { LOG_LEVEL_UNSPECIFIED = 0; DEBUG = 1; INFO = 2; WARN = 3; ERROR = 4; }
message LogRecord { LogLevel level = 1; string message = 2; map<string, string> fields = 3; }
```

The sketch above is the intended shape; the final field names are settled when the `.proto` file is written and checked by `buf lint`. Enum value names will be prefixed (e.g. `LOG_LEVEL_DEBUG`) to satisfy buf's defaults.

### Tooling and packaging
- `buf` for lint, breaking-change detection (`buf breaking` against main) and generation from one `buf.gen.yaml`: `protoc-gen-go`, protobuf-es (TypeScript), official `protobuf` with `mypy-protobuf` stubs (Python).
- Generated code is committed in each client so users never need `protoc`; CI fails if it is stale.
- Sidecar binaries are built with GoReleaser. Python ships per-platform wheels; npm ships per-platform packages as `optionalDependencies` (e.g. `@crunner/linux-x64`).
- The sidecar lives in the same Go module under `cmd/`, but protobuf is only imported by packages below it, so the root package stays dependency-light.

### Conformance suite
Test cases are data: a fake clock, registrations, scripted client behaviour (answer, hang, fail) and the expected message trace. A Go harness runs them against the sidecar binary directly; each client runs the same cases through its public API. The sidecar uses an injectable clock (`testing/synctest` in Go tests) for deterministic traces.

## Risks / Trade-offs

- [Client ignores cancellation, leaking work] → The scheduler stops waiting after cancel and drops late results; clients document that cancellation is cooperative.
- [Clients drift apart in behaviour] → Shared conformance suite; clients contain no scheduling logic.
- [Binary distribution burden (platforms, signing, antivirus)] → GoReleaser matrix, checksum verification, version pinning between client and sidecar.
- [Slow Python tasks starve other tasks] → `max_concurrent_runs`; Python client runs tasks as asyncio tasks and offers a thread-pool option for sync functions.
- [No custom schedules in v1] → Documented; `next_time` request/response can be added without breaking the schema.
- [Frame size or malformed input] → Maximum frame size, and the session terminates on corrupt frames.

## Migration Plan

Purely additive; nothing to migrate. Rollout: schema and sidecar first, then the TypeScript client, then Python, each verified by the conformance suite. Reverting means removing the new directories.

## Open Questions

- Exact name and layout of the published packages (`crunner` on PyPI, `@crunner/...` on npm).
- Whether the Python client uses plain `protobuf` or `betterproto`; defaults to plain `protobuf` for maintenance reasons.
