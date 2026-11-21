## MODIFIED Requirements

### Requirement: Registering a task
The scheduler SHALL accept a task function, a schedule and per-task options, and store the task under its ID. Registration is allowed whether or not the scheduler is running.

#### Scenario: Register while running
- **WHEN** a task is registered on a running scheduler
- **THEN** the task is scheduled immediately according to its schedule

#### Scenario: Register while idle
- **WHEN** a task is registered on an idle scheduler
- **THEN** the task is stored and scheduled when the scheduler starts

#### Scenario: Valid registration returns no error
- **WHEN** a non-nil task is registered with a non-nil schedule and valid options
- **THEN** RegisterTask returns nil

## ADDED Requirements

### Requirement: Registration validation
RegisterTask SHALL reject, with a non-nil error and without storing or scheduling anything, a nil task, a nil schedule, an interval schedule with a non-positive interval, a negative MaxDrift, a NotAfter that is not after NotBefore, backfill without NotBefore, and a nil callback function. The error SHALL match an exported sentinel error via errors.Is, and all detected problems SHALL be reported together.

#### Scenario: Nil task
- **WHEN** RegisterTask is called with a nil task
- **THEN** an error matching ErrInvalidTask is returned and no task is stored

#### Scenario: Nil schedule
- **WHEN** RegisterTask is called with a nil schedule
- **THEN** an error matching ErrInvalidSchedule is returned and no task is stored

#### Scenario: Zero interval
- **WHEN** RegisterTask is called with an interval schedule of 0
- **THEN** an error matching ErrInvalidSchedule is returned

#### Scenario: Negative drift
- **WHEN** RegisterTask is called with a negative MaxDrift
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Inverted time range
- **WHEN** RegisterTask is called with NotAfter equal to or before NotBefore
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Backfill without start
- **WHEN** RegisterTask is called with backfill enabled and no NotBefore
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Nil callback
- **WHEN** RegisterTask is called with a success, error or cancel callback option holding a nil function
- **THEN** an error matching ErrInvalidOption is returned

#### Scenario: Multiple problems
- **WHEN** RegisterTask is called with a nil task and a negative MaxDrift
- **THEN** the returned error matches both ErrInvalidTask and ErrInvalidOption

### Requirement: Rejected registration leaves existing task intact
A registration that fails validation SHALL NOT replace, cancel or reschedule a task already registered under the same ID.

#### Scenario: Invalid overwrite
- **WHEN** a task is registered under ID "a" and then an invalid registration is attempted under ID "a"
- **THEN** the original task keeps running on its schedule and its cancel callback is not called
