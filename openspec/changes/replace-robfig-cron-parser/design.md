# Design

## Context

`schedule.go` builds a robfig parser with `Minute|Hour|Dom|Month|Dow|Descriptor` (plus `Second` when `includeSeconds`) and wraps the result in `cronSchedule`. Behaviour of robfig/cron v3.0.1 was reviewed in `parser.go`, `spec.go` and `constantdelay.go` and probed with a small program; the observed behaviour is what the `schedules` delta spec pins down. Notable quirks found:

- `TZ=`/`CRON_TZ=` handling slices to the first space without checking for one: `TZ=UTC` panics.
- `getField` splits with `strings.FieldsFunc`, so empty list items are dropped; `,` yields no bits and `Next` returns the zero time.
- `mustParseInt` uses `strconv.Atoi`, so `+5` is accepted.
- Names are lower-cased before lookup; descriptors are not (`@Daily` is rejected).
- `*`/`?` set a "star bit" that is cleared by a step > 1; day matching ANDs dom/dow when either has the star bit, ORs otherwise.
- `Every` rounds down to whole seconds with a 1 s minimum; its `Next` truncates the input's nanoseconds.
- A schedule without `TZ=` uses `time.Local` as a marker meaning "evaluate in the input time's zone"; a prefix zone always wins.
- `Next` gives up after 5 years and returns `time.Time{}`, which `cronSchedule.Next` passes on with a nil error.

## Goals / Non-Goals

**Goals:**
- Same accepted language and same next times as today for every expression crunner accepts today, apart from the three tightenings listed in the proposal.
- Standard library only; small enough to review (~300 lines).

**Non-Goals:**
- Extensions (`L`, `W`, `#`, `H`, year field, `7` for Sunday, upper-case descriptors).
- Changing how the location argument is defaulted or applied on the first run — owned by `fix-cron-time-zone-handling`.

## Decisions

**Port the algorithm, not invent one.** Keep the bitset representation (one `uint64` per field plus a star flag) and robfig's field-by-field increment-with-wraparound search, including its DST handling (adding hours/days to wall-clock time, re-normalising when midnight is skipped). That is what makes identical next times across DST transitions cheap to guarantee. Alternative: a minute-by-minute scan — simpler but slow for sparse expressions with seconds and would need its own DST rules.

**Differential tests during development.** Before removing the dependency, a test compares the new parser with robfig for a table of expressions (all spec scenarios plus a generated set over every field form) and a range of reference times including DST transitions in Europe/Amsterdam, America/Sao_Paulo (historic midnight skip) and Australia/Lord_Howe (30-minute shift). The comparison test is deleted in the same change when the dependency is removed; the table is kept as golden expectations.

**Exported type.** `type CronExpression interface { Next(time.Time) time.Time }`, returned by `ParseCron`. It has the same method set as `cron.Schedule`, so most callers keep compiling. `Next` still returns the zero time on exhaustion at this level; `cronSchedule.Next` converts that into an error (spec "Cron search limit"), so the scheduler's existing schedule-error path applies.

**Tightenings are errors, not silent fixes.** Panics and never-firing schedules become parse/next errors so callers find out at registration time.

**Location precedence.** Unchanged: prefix zone, else the zone of the time passed to `Next` (which `cronSchedule` sets from its location argument). Keeping the sentinel semantics internal avoids touching `fix-cron-time-zone-handling`'s territory.

## Risks / Trade-offs

- [Subtle divergence from robfig on an untested input] → differential test across generated expressions and DST zones before the dependency is dropped.
- [Users depending on `cron.Schedule` type] → called out as BREAKING in the proposal; the method set is identical.
- [Tightening `,` to an error breaks a config that "worked"] → it never fired; an error is strictly more useful.

## Migration Plan

Single change: add parser with differential tests → switch `ParseCron` → delete comparison test → `go mod tidy` removes robfig. Rollback is reverting the change.
