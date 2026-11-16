## Context

See proposal.md for motivation. Every log statement in `scheduler.go` uses zerolog's chained builder (`s.logger.Trace().Str("task", ...).Msg(...)`); there are ~50 such calls, about half at Trace. The logger is stored by value in `internalScheduler.logger` and defaults to `zerolog.Nop()` (`scheduler.go:186`). Go version is 1.26, so `slog.DiscardHandler` (Go 1.24+) is available.

## Goals / Non-Goals

**Goals:**
- Zero third-party logging dependency; identical messages and levels after migration.
- Cheap when disabled: trace logging on hot paths must not allocate when the handler is not enabled for it.

**Non-Goals:**
- Changing what is logged or removing the periodic summary (that is `replace-summary-logging-with-hooks`).
- Context-propagated loggers (logger from ctx).

## Decisions

- **`*slog.Logger` rather than `slog.Handler`** in `WithLogger`: it is what applications already hold, and `logger.Handler()` is available if needed. Alternative (accept a Handler) was rejected as less ergonomic.
- **Default `slog.New(slog.DiscardHandler)`**, and `WithLogger(nil)` maps to the same; avoids nil checks at every call site. Alternative `slog.Default()` rejected: a library should be silent unless asked.
- **`LevelTrace = slog.Level(-8)`** exported as a constant so users can set `HandlerOptions{Level: crunner.LevelTrace}` and use `ReplaceAttr` to print it as "TRACE". -8 follows the slog convention of 4-step gaps below Debug(-4).
- **Call style**: `s.logger.Log(ctx, LevelTrace, "msg", slog.String("task", id), ...)` for trace; `s.logger.DebugContext/InfoContext/WarnContext/ErrorContext` for the rest, passing the scheduler context where available, else `context.Background()`. Errors use `slog.Any("error", err)`; durations `slog.Duration`; times `slog.Time`. A small private helper (`s.trace(ctx, msg, attrs...)`) keeps call sites short and guards with `s.logger.Enabled(ctx, LevelTrace)` before building attributes on hot paths.
- **Attribute keys** keep the zerolog field names (`task`, `fireAfter`, `expectedTriggerTime`, ...) so log-based dashboards keep working, except zerolog's `Err()` key `error`, which is kept as `error`.

## Risks / Trade-offs

- [Users depending on zerolog-specific output formatting] → Documented as breaking; they can bridge with a zerolog-backed `slog.Handler`.
- [Large mechanical diff may conflict with other in-flight changes touching `scheduler.go`] → Land early; the change is purely mechanical and has no behavioural coupling.
- [Formatting `Msgf` call at `scheduler.go:452`] → Converted to a fixed message with attributes; slog has no printf variant.

## Migration Plan

Callers replace `crunner.WithLogger(zlog)` with `crunner.WithLogger(slog.New(handler))`. Rollback is reverting the commit; no persisted state involved.
