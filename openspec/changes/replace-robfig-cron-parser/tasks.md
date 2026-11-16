# Tasks

## 1. Parser

- [ ] 1.1 Add `cron.go` with field bounds, month/day names, and parsing of lists, `*`/`?`, values (incl. leading `+`), ranges and steps into bitsets with a star flag; verify with table tests in `cron_test.go` covering every scenario of "Cron field count", "Cron field values", "Cron field bounds", "Cron names" and "Cron steps"
- [ ] 1.2 Add descriptor parsing (`@yearly` … `@hourly`, exact lower case) and `@every` with whole-second truncation and 1 s minimum; verify with table tests for "Cron descriptors" and "Every descriptor"
- [ ] 1.3 Add `TZ=`/`CRON_TZ=` prefix parsing that errors on unknown zones and on a missing expression; verify with tests for "Cron time-zone prefix", including that `TZ=UTC` returns an error without panicking

## 2. Next-time calculation

- [ ] 2.1 Port the field-wise next-time search with dom/dow AND/OR rule, location precedence and 5-year limit; verify with tests for "Cron day matching" and "Cron search limit"
- [ ] 2.2 Verify DST behaviour with tests for "Cron daylight-saving behaviour" (Europe/Amsterdam spring forward and fall back 2026) and a Sao Paulo historic midnight-skip case
- [ ] 2.3 Add a temporary differential test comparing `Next` against robfig/cron v3 for the spec table plus generated expressions and reference times across DST zones; verify it passes with `go test ./...`

## 3. Switch over

- [ ] 3.1 Export `CronExpression`, make `ParseCron` return it, and make `cronSchedule` use the new parser; verify existing `schedule_test.go` and `scheduler_test.go` pass unchanged
- [ ] 3.2 Make `cronSchedule.Next` return an error when no next time exists; verify with a synctest scheduler test that a task with `0 0 30 2 *` is not scheduled and the error is logged
- [ ] 3.3 Convert the differential test's results into golden expectations, delete the comparison test, run `go mod tidy`, and verify `go.mod` no longer lists `github.com/robfig/cron/v3` and `go test -race ./...` passes

## 4. Docs

- [ ] 4.1 Document the accepted cron syntax, descriptors, time-zone prefix and DST behaviour in the `ParseCron`/`WithCronSchedule` godoc and README, and note the BREAKING return-type change in the changelog; verify `go doc` renders it
