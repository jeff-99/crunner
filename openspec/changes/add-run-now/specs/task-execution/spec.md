## ADDED Requirements

### Requirement: Manual runs
The scheduler SHALL allow a registered task to be run immediately by ID while the scheduler is running. A manual run SHALL NOT be subject to MaxDrift, SHALL invoke the success or error callback and SHALL count towards the run statistics.

#### Scenario: Run now
- **WHEN** RunNow is called for a registered, idle task on a running scheduler
- **THEN** the task starts immediately and its success callback is called when it returns nil

#### Scenario: Unknown task
- **WHEN** RunNow is called with an ID that is not registered
- **THEN** an error is returned

#### Scenario: Scheduler not running
- **WHEN** RunNow is called on an idle scheduler
- **THEN** an error is returned and the task does not run

### Requirement: Manual run window
A manual run SHALL cover the window from the end of the task's last completed scheduled window (or, if no scheduled run has completed, the start of its first window) to the time of the request, and its metadata SHALL mark it as manual. A manual run SHALL NOT change the windows of scheduled runs.

#### Scenario: Window of a manual run
- **WHEN** a 1-hour interval task whose last scheduled window ended at 10:00 is run manually at 10:20
- **THEN** the manual run's window is 10:00 to 10:20 and its metadata is marked manual

#### Scenario: Scheduled window unaffected
- **WHEN** that task's next scheduled run fires at 11:00
- **THEN** its window is still 10:00 to 11:00 and it is not marked manual

### Requirement: Manual runs do not overlap
A manual run SHALL NOT start while the same task is running. A scheduled trigger that fires during a manual run SHALL wait until the manual run returns and is then subject to MaxDrift.

#### Scenario: Task already running
- **WHEN** RunNow is called while the task is executing
- **THEN** an error is returned and no second instance is started

#### Scenario: Scheduled trigger during manual run
- **WHEN** a scheduled trigger fires while a manual run of the same task is executing
- **THEN** the scheduled run starts only after the manual run returns, or is skipped if it is then later than MaxDrift
