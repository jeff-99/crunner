## ADDED Requirements

### Requirement: Run limit option validation
Registering a task SHALL fail with an error, and leave any existing task with the same ID untouched, when NotAfter is not later than NotBefore or when the maximum number of runs is zero or negative.

#### Scenario: NotAfter before NotBefore
- **WHEN** a task is registered with NotBefore 12:00 and NotAfter 11:00
- **THEN** RegisterTask returns an error and the task is not registered

#### Scenario: Non-positive max runs
- **WHEN** a task is registered with a maximum of 0 runs
- **THEN** RegisterTask returns an error and the task is not registered
