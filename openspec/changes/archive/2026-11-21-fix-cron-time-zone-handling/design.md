## Context

`cronSchedule.Next` (schedule.go ~50) converts `lastRun` with `.In(c.Location)` but passes `now` unconverted on the first run. robfig/cron's `SpecSchedule.Next(t)` evaluates fields in `t`'s location unless the spec carried `CRON_TZ=`, in which case it converts internally. `time.Time.In(nil)` panics.

## Goals / Non-Goals

**Goals:**
- No panic for nil location; consistent zone for all computations.

**Non-Goals:**
- DST semantics documentation and the parser itself (`replace-robfig-cron-parser`).
- Changing which time is used as the base for the next run (`make-trigger-windows-contiguous`).

## Decisions

- **Default nil to `time.Local` at construction** in `WithCronSchedule`, so `Next` never sees nil. Alternative: UTC — rejected; robfig/cron and gocron default to local, and changing it would silently shift existing schedules that pass nil. Users wanting UTC pass `time.UTC`.
- **Single code path in `Next`**: pick `base := now` or `*lastRun`, then `c.schedule.Next(base.In(c.Location))`. Removes the asymmetric branch.
- **Guard zero-value `cronSchedule{}`** (Location nil, not built via constructor): treat nil as `time.Local` inside `Next` too, cheap and avoids panics.
- **Prefix precedence** is inherited from robfig (it converts to the prefix zone internally); covered by a test so the parser replacement keeps it.

## Risks / Trade-offs

- [First run of existing cron tasks with a non-local location moves to a different instant] → This is the bug fix; mention in changelog.
- [`time.Local` differs between hosts/containers] → Documented; recommend passing an explicit location.
