# Proposal

## Why

The scheduler logs per-task summaries at info level on a fixed ticker (`emitTaskSummaries`, scheduler.go ~472-505, driven from `mainLoop` ~511-523, configured by `WithSummaryInterval`). What to report and how (logs, Prometheus, OpenTelemetry) is an application concern. The library also reads task metadata there while task goroutines write it (review finding 1). Applications instead need structured event hooks they can turn into metrics or logs.

## What Changes

- Add scheduler-wide event hooks via `WithHooks(Hooks)`: `BeforeRun`, `AfterRun`, `OnError`, `OnPanic`, `OnSkip`. Each receives an `Event` with task ID, trigger window, expected/actual trigger time, and (where applicable) duration and error.
- **BREAKING**: remove `WithSummaryInterval` and the periodic summary logging, including the summary ticker in the main loop and the `lastReported*` fields.
- Per-task callbacks (`OnSuccess`/`OnError`/`OnCancel`) are unchanged.

## Capabilities

### New Capabilities
- `task-events`: scheduler-wide lifecycle hooks for task runs (before, after, error, panic, skip), intended for metrics and observability.

### Modified Capabilities
- `scheduler-lifecycle`: construction no longer has a summary interval; the "Periodic task summaries" requirement is removed.

## Impact

- Code: `scheduler.go` (`WithSummaryInterval`, `summaryInterval`, `emitTaskSummaries`, ticker in `mainLoop`, `lastReported*` fields removed; hook calls in `processTask` and task goroutine), new `hooks.go`.
- API: **BREAKING** — `WithSummaryInterval` removed; applications wanting summaries build them from hooks. Additive: `Hooks`, `Event`, `WithHooks`.
- Dependencies: depends on `recover-task-panics` (the `OnPanic` hook fires on a recovered panic). Textual overlap with `replace-zerolog-with-slog` (summary log calls) and `split-cancel-callback` (skip path); rebase whichever lands later.
