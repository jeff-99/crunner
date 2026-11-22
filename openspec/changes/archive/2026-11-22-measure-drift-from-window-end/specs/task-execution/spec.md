## MODIFIED Requirements

### Requirement: Drift skipping
The expected trigger time of a run SHALL be the end of its trigger window. When a task that is not backfilling would start more than its maximum drift (default 30 seconds) after its expected trigger time, for whatever reason including the time taken by its own previous run, the run SHALL be skipped: it SHALL NOT be executed and the task's skip callback SHALL be invoked once. The next run's window SHALL start at the skipped window's start and end at the first point of the schedule's sequence after the current time, and that run SHALL be triggered at that end as normal. No part of the timeline SHALL be left uncovered. A run that would start late by at most the maximum drift SHALL start immediately. The skip callback SHALL be invoked asynchronously so that it never blocks scheduling, and any number of simultaneous skips SHALL NOT stall the scheduler. The removed callback SHALL NOT be invoked for a skipped run.

#### Scenario: Excessive drift
- **WHEN** an interval task of 10 seconds with MaxDrift 5 seconds has window [10s, 20s) and its timer fires at 27s
- **THEN** the run is not executed, the skip callback is called once, and the next run has window [10s, 30s) and is triggered at 30s

#### Scenario: Long pause
- **WHEN** the same task's timer for window [10s, 20s) fires at 83s because the process was paused
- **THEN** the skip callback is called once, no run is executed at 83s, and the next run has window [10s, 90s) and is triggered at 90s

#### Scenario: Restart after a long stop
- **WHEN** the scheduler is stopped at 15s while the same task's window [10s, 20s) is pending and started again at 83s
- **THEN** the skip callback is called once, and the next run has window [10s, 90s) and is triggered at 90s

#### Scenario: Long-running task is not skipped
- **WHEN** an interval task of 10 seconds with MaxDrift 5 seconds has window [10s, 20s), its run starts at 20s and takes 12.5 seconds, and the next window is [20s, 30s)
- **THEN** the next run starts immediately at 32.5s with window [20s, 30s), because it is late by 2.5 seconds, which is within MaxDrift, and the skip callback is not called

#### Scenario: Overrun beyond MaxDrift
- **WHEN** the same task has window [40s, 50s) and the previous run, started at 40s, takes 19 seconds
- **THEN** at 59s the run for window [40s, 50s) is skipped because it would be 9 seconds late, the skip callback is called once, and the next run has window [40s, 60s) and is triggered at 60s

#### Scenario: Blocking skip callback
- **WHEN** a run is skipped for drift and its skip callback blocks indefinitely
- **THEN** other tasks keep running on their schedules and the skipped task's next run is still scheduled

#### Scenario: Mass skip
- **WHEN** more runs are skipped for drift at the same moment than the scheduler's internal queue capacity (more than 100 tasks)
- **THEN** every skipped task is rescheduled and the scheduler continues to run tasks

#### Scenario: Stop waits for skip callbacks
- **WHEN** Stop is called while a skip callback is running
- **THEN** Stop returns only after that callback has returned

#### Scenario: Skip does not call the removed callback
- **WHEN** a run is skipped for drift on a task that has both a skip callback and a removed callback
- **THEN** only the skip callback is called

#### Scenario: Skip metadata
- **WHEN** a run with window [10s, 20s) is skipped for drift at 27s
- **THEN** the skip callback receives trigger metadata with window start 10s, window end 20s, the run's expected trigger time, and 27s as actual trigger time
