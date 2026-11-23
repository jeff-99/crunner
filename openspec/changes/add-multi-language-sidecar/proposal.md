# Proposal

## Why

crunner is a Go-only library, but its trigger-window semantics (backfill, drift skipping, non-overlapping runs) are valuable to Python and TypeScript users. Re-implementing the scheduler per language would create three codebases that drift apart. Instead, the Go core stays the single source of truth and is exposed to other languages as a sidecar process with a statically typed protocol from which thin clients are generated.

## What Changes

- Add a Protobuf schema (`proto/crunner/v1/scheduler.proto`) defining one bidirectional message session between a client and the scheduler, with `ClientMessage` and `ServerMessage` envelopes.
- Add a Go sidecar binary (`cmd/crunner-sidecar`) that wraps the existing scheduler. It serves the protocol over stdio using varint length-delimited frames, and turns every task run into a `RunRequest` sent to the client.
- Add buf configuration for linting, breaking-change detection and code generation (Go, TypeScript via protobuf-es, Python).
- Add a TypeScript client and a Python client, each spawning the sidecar as a child process and exposing a native-feeling API (async functions, `AbortSignal` / asyncio cancellation).
- Add a language-neutral conformance suite that drives the sidecar over the protocol with a fake clock and asserts event traces, shared by all clients.
- v1 supports cron and interval schedules only; custom `Schedule` implementations remain Go-only.
- No change to the existing Go library API or behaviour.

## Capabilities

### New Capabilities
- `sidecar-protocol`: the wire contract between a client and the scheduler: framing, message envelopes, request/response matching, task registration, run delegation, cancellation, notifications, errors, and versioning.
- `sidecar-process`: the sidecar binary's lifecycle: how it is started, how it initialises and shuts down, how it handles client loss, back-pressure and log output.

### Modified Capabilities
<!-- None: the in-process scheduler's requirements are unchanged. -->

## Impact

- New code: `proto/`, `cmd/crunner-sidecar/`, `clients/typescript/`, `clients/python/`, `buf.yaml`, `buf.gen.yaml`, conformance suite, release tooling (e.g. GoReleaser) for per-platform binaries.
- New dependencies: `google.golang.org/protobuf` in the sidecar only. The root `crunner` package gains none, in line with the project's preference for the standard library. Client packages depend on protobuf runtimes.
- No breaking API changes. Does not depend on other proposed changes.
- Extension points reserved for later changes: custom schedules via a `next_time` callback, daemon mode over gRPC, opaque task payloads.
