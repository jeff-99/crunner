## MODIFIED Requirements

### Requirement: Concurrency safety
All exported scheduler methods SHALL be safe to call concurrently with each other and with running tasks, including Start and Stop themselves, and the scheduler SHALL NOT perform unsynchronised concurrent access to per-task state such as timers and trigger metadata, or to its own lifecycle state such as the scheduler context.

#### Scenario: Concurrent register, remove and stop under the race detector
- **WHEN** tasks are registered, replaced and removed from several goroutines while the scheduler is running and tasks are executing, and then Stop is called, with the race detector enabled
- **THEN** no data race is reported and Stop returns after all running tasks have returned

#### Scenario: Metadata updated while summaries are emitted
- **WHEN** tasks complete runs while the periodic summary is being produced, with the race detector enabled
- **THEN** no data race is reported and run counts reflect every completed run

#### Scenario: Replaced task's pending timer does not run it
- **WHEN** a task with a pending timer is replaced or removed and its old timer fires afterwards
- **THEN** the old task is not executed

#### Scenario: Concurrent Start, Stop, registration and removal
- **WHEN** Start and Stop are called repeatedly from several goroutines while other goroutines register and remove tasks, with the race detector enabled
- **THEN** no data race is reported, calls that do not apply to the current state return their usual error or no-op, and a final Stop leaves the scheduler idle
