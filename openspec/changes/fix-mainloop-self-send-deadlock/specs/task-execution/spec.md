## MODIFIED Requirements

### Requirement: Maximum drift
The scheduler SHALL skip a run whose actual start is later than its expected trigger time by more than the task's maximum drift (default 30 seconds), invoke its cancel callback, and schedule the next run from the current time. The cancel callback SHALL be invoked asynchronously so that it never blocks the scheduling of other tasks, and skipping any number of runs at once SHALL NOT stall the scheduler.

#### Scenario: Excessive drift
- **WHEN** a run starts more than MaxDrift after its expected trigger time
- **THEN** the task is not executed, the cancel callback is called and the next run is scheduled

#### Scenario: Blocking cancel callback
- **WHEN** a run is skipped for drift and its cancel callback blocks indefinitely
- **THEN** other tasks keep running on their schedules and the skipped task's next run is still scheduled

#### Scenario: Mass skip
- **WHEN** more runs are skipped for drift at the same moment than the scheduler's internal queue capacity (more than 100 tasks)
- **THEN** every skipped task is rescheduled and the scheduler continues to run tasks

#### Scenario: Stop waits for skip callbacks
- **WHEN** Stop is called while a drift-skip cancel callback is running
- **THEN** Stop returns only after that callback has returned
