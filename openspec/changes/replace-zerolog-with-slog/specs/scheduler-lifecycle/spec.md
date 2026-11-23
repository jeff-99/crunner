## MODIFIED Requirements

### Requirement: Scheduler construction
The library SHALL provide a constructor that returns an idle scheduler, configurable through functional options for the logger and the task summary interval. The logger option SHALL accept a standard-library structured logger (`*slog.Logger`); when no logger or a nil logger is given, the scheduler SHALL discard all log output.

#### Scenario: Default construction
- **WHEN** a scheduler is created without options
- **THEN** it is idle, logs nothing, and uses a task summary interval of 30 seconds

#### Scenario: Custom structured logger
- **WHEN** a scheduler is created with a `*slog.Logger` whose handler records entries and a task runs
- **THEN** the handler receives structured records carrying the task ID as an attribute

#### Scenario: Nil logger
- **WHEN** a scheduler is created with a nil logger
- **THEN** it behaves as if no logger was given and logs nothing

## ADDED Requirements

### Requirement: Trace log level
The library SHALL export a `LevelTrace` log level equal to `slog.Level(-8)` and SHALL log its most verbose internal scheduling messages at that level, below debug.

#### Scenario: Trace messages hidden at debug level
- **WHEN** the configured handler's minimum level is debug and a task is scheduled and run
- **THEN** no records at `LevelTrace` are emitted

#### Scenario: Trace messages visible at trace level
- **WHEN** the configured handler's minimum level is `LevelTrace` and a task is scheduled and run
- **THEN** records at `LevelTrace` describing scheduling steps are emitted
