## Why

crunner depends on `github.com/rs/zerolog` only for its logger option (`scheduler.go:11`, `WithLogger(zerolog.Logger)` at line 22). Forcing a third-party logging library on every user of an open-source scheduler is unnecessary now that the standard library ships `log/slog`, and the project prefers the standard library.

## What Changes

- **BREAKING**: `WithLogger` takes a `*slog.Logger` instead of a `zerolog.Logger`.
- Default logger discards all output (same observable behaviour as today's `zerolog.Nop()`); passing a nil logger also means discard.
- New exported constant `LevelTrace = slog.Level(-8)` used for the scheduler's very verbose internal messages (today's `Trace()` calls); other levels map 1:1 (Debug, Info, Warn, Error).
- All log calls in `scheduler.go` are rewritten to slog with structured attributes (`task`, `windowStart`, `windowEnd`, ...), keeping the same messages and levels.
- `github.com/rs/zerolog` is removed from `go.mod`.

## Capabilities

### New Capabilities

### Modified Capabilities
- `scheduler-lifecycle`: the "Scheduler construction" requirement now specifies a standard-library structured logger, discard-by-default, and an exported trace level.

## Impact

- Code: `scheduler.go` (option, struct field, every log call); tests that construct loggers, if any.
- API: **BREAKING** — `WithLogger(zerolog.Logger)` becomes `WithLogger(*slog.Logger)`. Callers wrap their existing handler or use `slog.Default()`. New exported `LevelTrace`.
- Dependencies: removes `github.com/rs/zerolog` (and its transitive deps) from `go.mod`/`go.sum`; adds none.
- Depends on: none. Other changes that touch log calls (e.g. `replace-summary-logging-with-hooks`, `fix-shared-state-races`) are easier to land after this one but do not require it.
