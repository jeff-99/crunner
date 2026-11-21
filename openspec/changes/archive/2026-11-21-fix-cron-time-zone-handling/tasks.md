## 1. Implementation

- [x] 1.1 In `WithCronSchedule` (schedule.go), default a nil `location` to `time.Local`; verify with a unit test that the returned schedule's `Location` is `time.Local`
- [x] 1.2 Rewrite `cronSchedule.Next` to choose the base time (`now` or `*lastRun`), convert it to the location (nil treated as `time.Local`) and call the parser once; verify `go vet ./...`

## 2. Tests

- [x] 2.1 Test: nil location — `Next` with and without `lastRun` does not panic (also for a zero-value `cronSchedule` with a parsed schedule)
- [x] 2.2 Test: `0 9 * * *` in `America/New_York`, first run from 2026-01-01T12:00:00Z returns 14:00:00Z (fails on main)
- [x] 2.3 Test: same schedule with `lastRun` 2026-01-01T14:00:00Z returns 2026-01-02T14:00:00Z
- [x] 2.4 Test: `CRON_TZ=Asia/Tokyo 0 9 * * *` with `America/New_York` fires at 09:00 Tokyo
- [x] 2.5 Test (testing/synctest): a registered cron task with a non-UTC location receives a first trigger window ending at the expected instant
- [x] 2.6 Run `go test -race ./...` and confirm all tests pass (run with -count=20 -skip /5_days plus a full -count=1)
