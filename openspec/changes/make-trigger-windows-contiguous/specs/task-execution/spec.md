## MODIFIED Requirements

### Requirement: Trigger windows
The windows of a task SHALL be contiguous and non-overlapping: each window SHALL start exactly at the previous window's end, and its end SHALL be the schedule's next time computed from the previous window's end, independent of when runs actually started or finished. The first window SHALL start when the task is first scheduled — at registration on a running scheduler, at Start for a task registered while the scheduler was idle — or at NotBefore when backfilling. Skipped windows SHALL still count as covered for the purpose of contiguity.

#### Scenario: First window
- **WHEN** a task without backfill is registered on a running scheduler and runs for the first time
- **THEN** its window starts at the registration time and ends at the schedule's next run time

#### Scenario: First window of a task registered before Start
- **WHEN** an interval task of 10 seconds is registered on an idle scheduler and the scheduler is started 1 minute later
- **THEN** its first window starts at the Start time and ends 10 seconds later, and the run is not skipped

#### Scenario: Late start does not shift windows
- **WHEN** an interval task of 10 seconds with window [T, T+10s) starts its run 2 seconds late
- **THEN** its next window is [T+10s, T+20s), not [T+10s, T+22s)

#### Scenario: Windows stay on the grid over time
- **WHEN** an interval task of 10 seconds runs 100 times, each run starting with a small delay
- **THEN** the 100th window ends exactly 1000 seconds after the first window's start

### Requirement: Maximum drift
The expected trigger time of a run SHALL be the later of its window end and the time the task's previous run finished. The scheduler SHALL skip a run whose actual start is later than its expected trigger time by more than the task's maximum drift (default 30 seconds) and invoke its cancel callback. For a task that is not backfilling, the skip SHALL cover every window that ended more than the maximum drift before the current time, with a single cancel callback invocation, and the next run SHALL be the first window that has not.

#### Scenario: Excessive drift
- **WHEN** a run starts more than MaxDrift after its expected trigger time
- **THEN** the task is not executed, the cancel callback is called and the next run is scheduled

#### Scenario: Long-running task is not skipped
- **WHEN** an interval task of 10 seconds with MaxDrift 1 second runs for 25 seconds
- **THEN** its next window starts at the previous window's end, the next run starts immediately after the long run finishes, and the cancel callback is not called

#### Scenario: Several missed windows skipped at once
- **WHEN** the timer of an interval task of 10 seconds with MaxDrift 5 seconds fires 60 seconds late
- **THEN** the cancel callback is called once, no run is executed for windows that ended more than 5 seconds ago, and the next executed window starts where the last skipped window ended
