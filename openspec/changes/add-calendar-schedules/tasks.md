## 1. Types and validation

- [ ] 1.1 Add `TimeOfDay` and `NewTimeOfDay` with range validation; verify with table-driven unit tests for valid and invalid values
- [ ] 1.2 Add `Daily`, `Weekly`, `Monthly` constructors returning `(Schedule, error)` with input validation (empty times/days, out-of-range days); verify with unit tests for each error case

## 2. Next computation

- [ ] 2.1 Implement `calendarSchedule.Next` (strictly-after base, nil location → `time.Local`); verify table-driven tests for daily multi-time, weekly weekday selection, monthly day-31 skipping and the exact-match case
- [ ] 2.2 Implement DST gap and overlap handling; verify tests around 2026 DST transitions in Europe/Amsterdam and America/New_York
- [ ] 2.3 Add a synctest scheduler test registering a daily task and asserting window start/end in the received metadata; verify it passes with `go test -race ./...`
- [ ] 2.4 Add doc comments and a package example (`Example_daily`) for the calendar schedules; verify `go test` runs the example
