## MODIFIED Requirements

### Requirement: Callbacks
The scheduler SHALL call the task's success callback after a run that returns nil, and its error callback with the error after a run that returns an error or panics. A panicking run SHALL be treated as an errored run whose error is a panic error carrying the panic value and stack trace.

#### Scenario: Error callback
- **WHEN** a task returns an error
- **THEN** the error callback is called with that error and the error count is incremented

#### Scenario: Success callback
- **WHEN** a task returns nil
- **THEN** the success callback is called and the success count is incremented

#### Scenario: Panic reported as error
- **WHEN** a task panics with value `"boom"`
- **THEN** the error callback is called with a panic error whose value is `"boom"` and which includes a stack trace, the error count is incremented, and the success callback is not called

## ADDED Requirements

### Requirement: Panic containment
A panic raised by a task function SHALL NOT terminate the process or the scheduler. After a panicking run the task SHALL be rescheduled as after any errored run, and other tasks SHALL keep running.

#### Scenario: Scheduler survives a panicking task
- **WHEN** a task panics on every run
- **THEN** the process keeps running, the task is run again on its next scheduled window, and other registered tasks keep running on their schedules

#### Scenario: Panic with an error value
- **WHEN** a task panics with an `error` value
- **THEN** the panic error passed to the error callback unwraps to that original error
