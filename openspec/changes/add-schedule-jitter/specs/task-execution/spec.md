## MODIFIED Requirements

### Requirement: Maximum drift
The scheduler SHALL skip a run whose actual start is later than its expected trigger time by more than the task's maximum drift (default 30 seconds) plus its maximum jitter (default 0), invoke its cancel callback, and schedule the next run from the current time.

#### Scenario: Excessive drift
- **WHEN** a run starts more than MaxDrift after its expected trigger time
- **THEN** the task is not executed, the cancel callback is called and the next run is scheduled

#### Scenario: Jitter does not cause drift skips
- **WHEN** a task with MaxDrift 1 second and jitter 10 seconds starts 8 seconds after its expected trigger time because of jitter
- **THEN** the run executes

## ADDED Requirements

### Requirement: Start jitter
When a task has a maximum jitter J greater than zero, each scheduled run SHALL start at its expected trigger time plus a random delay in [0, J), chosen independently per run. Jitter SHALL NOT change the run's trigger window or expected trigger time, and SHALL NOT apply to backfill catch-up runs.

#### Scenario: Jittered start within bounds
- **WHEN** a 1-minute interval task with jitter 10 seconds runs repeatedly
- **THEN** each run starts between 0 and 10 seconds after its expected trigger time

#### Scenario: Window unchanged
- **WHEN** a jittered run is invoked
- **THEN** its metadata window start, window end and expected trigger time equal those of the same run without jitter

#### Scenario: Backfill not jittered
- **WHEN** a task with jitter and backfill catches up past windows
- **THEN** catch-up runs start without jitter delay

#### Scenario: Negative jitter
- **WHEN** a task is registered with a negative jitter
- **THEN** RegisterTask returns an error and the task is not registered
