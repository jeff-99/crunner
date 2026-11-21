## Why

The cron schedule's time zone handling is broken in two ways (`cronSchedule.Next`, schedule.go ~50-56): a nil location makes `(*lastRun).In(nil)` panic, and the first run (`lastRun == nil`) evaluates the expression in the zone of `now` instead of the configured location. A schedule like `0 9 * * *` with `Europe/Amsterdam` can therefore fire its first run at 09:00 UTC/local of the host.

## What Changes

- A nil location passed to `WithCronSchedule` defaults to `time.Local` (the same default robfig/cron uses), resolved once at construction.
- Every next-run computation, including the first one, evaluates the cron expression in the configured location.
- Returned times are in the configured location (the instant is what matters; the zone is informative).
- An explicit `CRON_TZ=`/`TZ=` prefix in the spec keeps precedence over the location argument (today's parser behaviour, now specified).

## Capabilities

### New Capabilities

### Modified Capabilities
- `schedules`: Cron schedule requirement — defines default location and that all runs, including the first, honour it.

## Impact

- Code: `schedule.go` (`WithCronSchedule`, `cronSchedule.Next`), `schedule_test.go`.
- API: no signature change. Behaviour change: a nil location no longer panics, and first runs of cron tasks with a non-host location shift to the correct instant. Not breaking for correct usage.
- Dependencies on other changes: none. `replace-robfig-cron-parser` must preserve these semantics; `tidy-public-api` may rename the constructor.
