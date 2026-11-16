## MODIFIED Requirements

### Requirement: Maximum drift
The scheduler SHALL skip a run whose actual start is later than its expected trigger time by more than the task's maximum drift (default 30 seconds), invoke its skip callback asynchronously with the trigger metadata of the skipped run, and schedule the next run from the current time. The cancel callback SHALL NOT be invoked for a skipped run.

#### Scenario: Excessive drift
- **WHEN** a run starts more than MaxDrift after its expected trigger time
- **THEN** the task is not executed, the skip callback is called and the next run is scheduled

#### Scenario: Skip does not cancel
- **WHEN** a run is skipped for drift on a task that has both a skip and a cancel callback
- **THEN** only the skip callback is called

#### Scenario: Skip metadata
- **WHEN** a run is skipped for drift
- **THEN** the skip callback receives the skipped run's trigger window start and end, its expected trigger time and the time at which it was skipped as actual trigger time

### Requirement: Callbacks
The scheduler SHALL call the task's success callback after a run that returns nil, and its error callback with the error after a run that returns an error. It SHALL call the task's skip callback when a run is skipped for drift, and the cancel callback only when the task is removed or replaced.

#### Scenario: Error callback
- **WHEN** a task returns an error
- **THEN** the error callback is called with that error and the error count is incremented

#### Scenario: Success callback
- **WHEN** a task returns nil
- **THEN** the success callback is called and the success count is incremented

#### Scenario: Slow skip callback
- **WHEN** a skip callback blocks for a long time
- **THEN** other tasks continue to be scheduled and run on time
