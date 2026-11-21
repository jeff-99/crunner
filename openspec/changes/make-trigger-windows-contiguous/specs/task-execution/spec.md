## MODIFIED Requirements

### Requirement: Trigger windows
The windows of a task SHALL be contiguous and non-overlapping: each window SHALL start exactly at the end of the task's previous executed window, and its end SHALL be a point of the schedule's sequence computed from the previous window's end, independent of when runs actually started or finished. The first window SHALL start when the task is first scheduled — at registration on a running scheduler, at Start for a task registered while the scheduler was idle — or at NotBefore when backfilling. A window normally spans one step of the schedule; after a run is cancelled for drift, the next window spans several steps (see Maximum drift). Window starts and ends SHALL always be points of the schedule's sequence.

#### Scenario: First window
- **WHEN** a task without backfill is registered on a running scheduler and runs for the first time
- **THEN** its window starts at the registration time and ends at the schedule's next run time

#### Scenario: First window of a task registered before Start
- **WHEN** an interval task of 10 seconds is registered on an idle scheduler and the scheduler is started 1 minute later
- **THEN** its first window starts at the Start time and ends 10 seconds later, and the run is not cancelled

#### Scenario: Late start does not shift windows
- **WHEN** an interval task of 10 seconds with window [T, T+10s) starts its run 2 seconds late
- **THEN** its next window is [T+10s, T+20s), not [T+10s, T+22s)

#### Scenario: Windows stay on the grid over time
- **WHEN** an interval task of 10 seconds runs 100 times, each run starting with a small delay
- **THEN** the 100th window ends exactly 1000 seconds after the first window's start

### Requirement: Maximum drift
The expected trigger time of a run SHALL be the later of its window end and the time the task's previous run finished. When a task that is not backfilling would start more than its maximum drift (default 30 seconds) after its expected trigger time, the run SHALL NOT be executed and its cancel callback SHALL be invoked once. The next run's window SHALL start at the cancelled window's start and end at the first point of the schedule's sequence after the current time, and that run SHALL be triggered at that end as normal. No part of the timeline SHALL be left uncovered. The cancel callback SHALL be invoked asynchronously so that it never blocks scheduling, and any number of simultaneous cancellations SHALL NOT stall the scheduler.

#### Scenario: Excessive drift
- **WHEN** an interval task of 10 seconds with MaxDrift 5 seconds has window [10s, 20s) and its timer fires at 27s
- **THEN** the run is not executed, the cancel callback is called once, and the next run has window [10s, 30s) and is triggered at 30s

#### Scenario: Long pause
- **WHEN** the same task's timer for window [10s, 20s) fires at 83s because the process was paused
- **THEN** the cancel callback is called once, no run is executed at 83s, and the next run has window [10s, 90s) and is triggered at 90s

#### Scenario: Restart after a long stop
- **WHEN** the scheduler is stopped at 15s while the same task's window [10s, 20s) is pending and started again at 83s
- **THEN** the cancel callback is called once, and the next run has window [10s, 90s) and is triggered at 90s

#### Scenario: Long-running task is not cancelled
- **WHEN** an interval task of 10 seconds with MaxDrift 1 second runs for 25 seconds
- **THEN** its next window starts at the previous window's end and spans one interval, the next run starts immediately after the long run finishes, and the cancel callback is not called

#### Scenario: Blocking cancel callback
- **WHEN** a run is cancelled for drift and its cancel callback blocks indefinitely
- **THEN** other tasks keep running on their schedules and the cancelled task's next run is still scheduled

#### Scenario: Mass skip
- **WHEN** more runs are cancelled for drift at the same moment than the scheduler's internal queue capacity (more than 100 tasks)
- **THEN** every cancelled task is rescheduled and the scheduler continues to run tasks

#### Scenario: Stop waits for skip callbacks
- **WHEN** Stop is called while a drift cancel callback is running
- **THEN** Stop returns only after that callback has returned
