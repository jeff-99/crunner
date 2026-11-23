# Proposal

## Why

The only third-party dependency used at runtime for scheduling is `github.com/robfig/cron/v3` (`schedule.go:8`), which is unmaintained since 2020, panics on some inputs (e.g. `TZ=UTC` without a following expression: slice bounds out of range in `Parser.Parse`), and silently returns a zero time for expressions that never fire, which crunner then schedules as "fire immediately". Owning a small parser removes the dependency and lets crunner define and document the exact syntax and DST behaviour, while existing expressions keep working.

## What Changes

- Implement a cron parser and next-time calculation inside the `crunner` package (standard library only) that accepts exactly the expressions robfig/cron v3.0.1 accepts with the options crunner uses today (`Minute|Hour|Dom|Month|Dow|Descriptor`, plus `Second` when seconds are enabled), and computes the same next times — documented precisely in the `schedules` spec.
- Return parse errors instead of panicking for a `TZ=`/`CRON_TZ=` prefix not followed by an expression.
- A list field with no values (e.g. `,`) is rejected at parse time instead of producing a schedule that never fires.
- A cron schedule that has no next time within five years returns an error from `Next` instead of a zero time.
- Document DST behaviour: a wall-clock time that does not exist on a day is skipped for that day; a wall-clock time that occurs twice fires on both occurrences (unchanged from today).
- Add an opt-in DST-safe mode for cron schedules, enabled with a cron option (`CronDSTSafe()` once `tidy-public-api` introduces `CronOption`): a wall-clock time that does not exist on a day runs at the equivalent instant after the gap, and a wall-clock time that occurs twice runs only at its first occurrence. Without the option, behaviour stays identical to robfig.
- **BREAKING**: `ParseCron` returns an exported crunner type (`CronExpression`, with the same `Next(time.Time) time.Time` method) instead of `cron.Schedule`.
- Remove `github.com/robfig/cron/v3` from `go.mod`.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `schedules`: the "Cron schedule" requirement is made precise (field syntax, ranges, names, steps, day matching, descriptors, time-zone prefix, DST, search limit) and its error behaviour tightened.

## Impact

- Code: `schedule.go` (`ParseCron`, `cronSchedule`), new `cron.go` + `cron_test.go`; `go.mod`/`go.sum`.
- API: **BREAKING** — `ParseCron`'s return type changes from `cron.Schedule` to `crunner.CronExpression`. Code that only calls `.Next` on the result keeps compiling; code that names `cron.Schedule` or type-asserts to `*cron.SpecSchedule` must change. Expressions accepted today remain accepted except a list field with no values (e.g. `,`), which never fired anyway.
- Dependencies: removes `github.com/robfig/cron/v3`.
- Depends on: `tidy-public-api` for the `CronOption` type that carries the DST-safe option (the parser rewrite itself does not need it). Coordinates with `fix-cron-time-zone-handling` (owns the location *argument* of the cron schedule and the nil-location / first-run fixes; this change keeps today's rule that a `TZ=` prefix overrides that argument) and `tidy-public-api` (may rename `ParseCron`/`WithCronSchedule`; whichever lands second rebases).
