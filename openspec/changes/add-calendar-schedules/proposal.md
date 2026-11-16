## Why

The most common real schedules are "every day at 02:00", "Mondays and Thursdays at 09:30" or "the 1st of the month at midnight". Today these require writing cron expressions (`cronSchedule` in `schedule.go`), which is error-prone and stringly typed. gocron v2 offers `DailyJob`, `WeeklyJob` and `MonthlyJob`; crunner has no equivalent.

## What Changes

- Add a `TimeOfDay` value type (hour, minute, second) with a constructor that validates its fields.
- Add calendar schedule constructors, each returning a `Schedule` or an error on invalid input:
  - `Daily(loc *time.Location, times ...TimeOfDay)`
  - `Weekly(loc *time.Location, days []time.Weekday, times ...TimeOfDay)`
  - `Monthly(loc *time.Location, days []int, times ...TimeOfDay)` (days 1-31; months without that day are skipped)
- Calendar times are evaluated in the given location (nil means `time.Local`), with defined DST behaviour.
- Window semantics are unchanged: each window ends at the next calendar occurrence.

## Capabilities

### New Capabilities

### Modified Capabilities
- `schedules`: adds daily, weekly and monthly calendar schedules.

## Impact

- `schedule.go` (or new `calendar.go`): new types and constructors. Pure standard library.
- Not breaking: additive API only.
- Dependencies: none required. Follows the same nil-location convention as `fix-cron-time-zone-handling` (nil → `time.Local`); if `tidy-public-api` changes constructor naming conventions, apply them here too.
