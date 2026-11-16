## Purpose

Lets applications observe every task run, failure, panic and drift skip through scheduler-wide hooks, so they can derive metrics and logs without the library deciding how to report.

## ADDED Requirements

### Requirement: Hook registration
The scheduler SHALL accept an optional set of event hooks at construction. Each hook is optional; unset hooks SHALL be ignored. Hooks apply to all tasks of the scheduler.

#### Scenario: No hooks configured
- **WHEN** a scheduler is created without hooks and a task runs
- **THEN** the task runs normally and nothing is reported

### Requirement: Event payload
Every hook SHALL receive an event containing the task ID, the trigger window start and end, and the expected trigger time. Run-related events SHALL also contain the actual start time; after-run, error and panic events SHALL also contain the run duration; error and panic events SHALL also contain the error.

#### Scenario: Event identifies the run
- **WHEN** a task with ID `report` runs for window [T0, T1)
- **THEN** the before-run event carries ID `report`, window start T0 and window end T1

### Requirement: Before-run hook
The scheduler SHALL call the before-run hook immediately before invoking a task, after the drift check has passed.

#### Scenario: Before-run fires once per run
- **WHEN** a task runs three times
- **THEN** the before-run hook is called three times, each before the corresponding task invocation

### Requirement: After-run hook
The scheduler SHALL call the after-run hook after every task invocation completes, whether it succeeded, returned an error or panicked, with the run's duration and its error, if any.

#### Scenario: After-run on success
- **WHEN** a task returns nil after 2 seconds
- **THEN** the after-run hook is called with a duration of 2 seconds and no error

#### Scenario: After-run on failure
- **WHEN** a task returns an error
- **THEN** the after-run hook is called with that error

### Requirement: Error hook
The scheduler SHALL call the error hook when a task returns a non-nil error, before the after-run hook.

#### Scenario: Error reported
- **WHEN** a task returns error E
- **THEN** the error hook is called with E and then the after-run hook is called

### Requirement: Panic hook
The scheduler SHALL call the panic hook when a task panics and the panic is recovered, with the error derived from the panic value, before the after-run hook. The error hook SHALL NOT be called for a panic.

#### Scenario: Panic reported
- **WHEN** a task panics
- **THEN** the panic hook is called, the error hook is not called, and the after-run hook is called with the panic error

### Requirement: Skip hook
The scheduler SHALL call the skip hook when a run is skipped because it exceeded the task's maximum drift. Neither the before-run nor the after-run hook SHALL be called for a skipped run.

#### Scenario: Drift skip reported
- **WHEN** a run starts more than MaxDrift after its expected trigger time
- **THEN** the skip hook is called with the run's window and expected trigger time, and the task is not invoked

### Requirement: Hook isolation
A panic inside a hook SHALL be recovered and SHALL NOT stop the scheduler, affect the task's run statistics, or prevent subsequent runs.

#### Scenario: Panicking hook
- **WHEN** the before-run hook panics
- **THEN** the task still runs and is rescheduled as usual
