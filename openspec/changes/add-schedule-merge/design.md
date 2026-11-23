## Context

`Schedule.Next(ctx, schedulingStart, lastRun)` (`schedule.go`) returns the end of the next window. The scheduler passes `lastRun` = the previous window end (nil for the first window) and may call `Next` repeatedly with its own previous result after drift skips. The contract says the result is strictly after the reference, so a combinator only needs to take the minimum.

## Goals / Non-Goals

**Goals:**
- Express "any of these schedules" without a new parser or per-field logic.
- Work with every `Schedule`, not only cron.

**Non-Goals:**
- Intersections or exclusions ("every day except holidays").
- Per-child windows or per-child options; the task still has one window sequence.
- Changing DST behaviour; each child keeps its own.

## Decisions

- **Stateless fan-out:** `Next` calls each child with the same arguments and returns the earliest time. No state is kept in the merged schedule, so it stays safe for concurrent use as long as its children are. Alternative: keep a heap of cached next times per child. Rejected: it would break the "called repeatedly with the previous result" contract and make the schedule stateful.
- **Reference time semantics:** children receive the merged window end as `lastRun`. Relative schedules such as `Every(d)` therefore count from the previous merged window end, not from their own last occurrence. This is documented, and it is the only behaviour consistent with a single window sequence per task.
- **Errors propagate:** the first child error is returned. Alternative: skip failing children and use the rest. Rejected: it hides a broken child and silently changes the schedule.
- **Duplicates collapse for free:** the minimum of equal instants is one instant, so two children firing at the same time produce one run.
- **Validation at construction:** an empty list and a nil child are errors. One child returns that child unchanged to avoid a pointless wrapper.
- **No dependency on `add-one-time-schedule`:** exhaustion handling is specified there. When both exist, exhausted children are dropped and `ErrNoMoreRuns` is only returned when every child is exhausted.

## Risks / Trade-offs

- [A child with hidden state returns different values when called once per merge call] → document that children must derive their result from the arguments only; all built-in schedules do.
- [Many children make each `Next` O(children)] → negligible versus scheduling frequency.
- [Mixing a DST-aware and a default cron child gives each its own DST rule] → documented; behaviour is per child.
