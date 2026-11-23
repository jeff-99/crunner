## 1. Constructor and validation

- [ ] 1.1 Add `Merge(schedules ...Schedule) (Schedule, error)` with empty-list, nil-child and single-child handling; verify with unit tests for each case

## 2. Next computation

- [ ] 2.1 Implement the minimum-of-children `Next`, propagating child errors; verify table-driven tests for two cron children (06:00 and 18:30), mixed interval and cron children, equal instants and a failing child
- [ ] 2.2 Test repeated evaluation with the previous result as reference time to prove strictly increasing, complete output; verify it passes
- [ ] 2.3 Add a synctest scheduler test registering a task on a merged schedule and asserting the window start and end passed to the task for consecutive runs; verify it passes with `go test -race ./...`
- [ ] 2.4 Add doc comments, including the reference-time semantics for relative children and the stateless-children requirement, and an `Example_merge`; verify `go test` runs the example
