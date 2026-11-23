## MODIFIED Requirements

### Requirement: Stopping the scheduler
Stopping a running scheduler SHALL cancel the scheduler context, stop all pending timers, and wait for running tasks to return until the caller-supplied context is done. When all tasks have returned, the scheduler SHALL be idle. If the caller's context ends first, Stop SHALL return an error wrapping the context's error and the scheduler SHALL become idle only after the remaining tasks return.

#### Scenario: Stop waits for running tasks
- **WHEN** Stop is called while a task is executing
- **THEN** the task's context is cancelled and Stop returns only after the task has returned

#### Scenario: Stop when not running
- **WHEN** Stop is called on an idle scheduler
- **THEN** Stop returns no error

#### Scenario: Restart after stop
- **WHEN** a stopped scheduler is started again
- **THEN** its registered tasks are scheduled again

#### Scenario: Stop bounded by context
- **GIVEN** a running task that ignores context cancellation
- **WHEN** Stop is called with a context that times out after 1 second
- **THEN** Stop returns after about 1 second with an error that matches `context.DeadlineExceeded`

#### Scenario: Start while still draining
- **GIVEN** Stop has returned a timeout error and a task is still running
- **WHEN** Start is called
- **THEN** Start returns an error and no tasks are scheduled

#### Scenario: Idle after draining
- **GIVEN** Stop has returned a timeout error
- **WHEN** the remaining task finally returns
- **THEN** the scheduler becomes idle and can be started again
