# Spec Delta

## Purpose

Defines the typed, language-neutral wire contract between a client in any language and the crunner scheduler, so that clients can be generated from one schema and behave identically.

## ADDED Requirements

### Requirement: Single schema as source of truth
The protocol SHALL be defined by one versioned Protobuf schema in package `crunner.v1`. Generated client code SHALL be derived from it, and changes to it SHALL be checked for backward compatibility.

#### Scenario: Breaking schema change
- **WHEN** a change removes or renumbers a field or reuses a field number of the published schema
- **THEN** the compatibility check fails

#### Scenario: Unknown fields
- **WHEN** a peer receives a message containing fields it does not know
- **THEN** it ignores them and processes the rest of the message

### Requirement: Framed session
A session SHALL be a bidirectional sequence of Protobuf messages, each preceded by its length as a varint. Client-to-scheduler messages SHALL be `ClientMessage` and scheduler-to-client messages SHALL be `ServerMessage`, each carrying exactly one body from a `oneof`.

#### Scenario: Well-formed frame
- **WHEN** a peer writes a varint length followed by that many bytes of an encoded message
- **THEN** the receiver decodes exactly one message

#### Scenario: Message without a body
- **WHEN** the scheduler receives a `ClientMessage` whose body is not set
- **THEN** it replies with an `INVALID_ARGUMENT` error for that message id and the session continues

#### Scenario: Oversized or corrupt frame
- **WHEN** the scheduler receives a frame exceeding the maximum frame size or one that cannot be decoded
- **THEN** it terminates the session

### Requirement: Request and response matching
Every client request SHALL carry a non-zero `id`, and the scheduler's response SHALL carry the same `id` and be either the matching response body or an `Error`. Messages with id 0 SHALL be notifications and receive no response. Scheduler-initiated requests SHALL be correlated by `run_id` instead.

#### Scenario: Concurrent requests
- **WHEN** a client has several requests outstanding at once
- **THEN** each response carries the id of its request, in any order

#### Scenario: Failed request
- **WHEN** a request is rejected, for instance registering an invalid task
- **THEN** the response is an `Error` with the request's id and an error code

### Requirement: Initialization handshake
The first message of a session SHALL be an `InitializeRequest` carrying the client's protocol version and scheduler options (task timeout, schedule timeout, summary interval, queue size, maximum concurrent runs, log level). The scheduler SHALL answer with its protocol version and server version. Any other request before initialization SHALL be rejected with `INVALID_STATE`.

#### Scenario: Compatible version
- **WHEN** the client initializes with a protocol version the scheduler supports
- **THEN** the scheduler constructs its scheduler with the given options and responds with its versions

#### Scenario: Unsupported version
- **WHEN** the client initializes with a protocol version the scheduler does not support
- **THEN** the scheduler responds with an `INVALID_ARGUMENT` error and terminates the session

#### Scenario: Request before initialization
- **WHEN** the client sends a `StartRequest` before initializing
- **THEN** the scheduler responds with an `INVALID_STATE` error

#### Scenario: Repeated initialization
- **WHEN** the client sends a second `InitializeRequest`
- **THEN** the scheduler responds with an `INVALID_STATE` error and the existing configuration is unchanged

### Requirement: Scheduler control requests
The protocol SHALL provide requests to start and stop the scheduler, to register a task, to remove a task, to list tasks and to get one task. Their outcomes SHALL match the in-process scheduler: the same validation, the same replacement and removal semantics, and the same lifecycle errors mapped to error codes.

#### Scenario: Register and replace
- **WHEN** the client registers a task id that already exists
- **THEN** the existing task is replaced exactly as `RegisterTask` replaces it in-process

#### Scenario: Invalid registration
- **WHEN** the client registers a task with an empty id, no schedule, an invalid cron spec or an unknown time zone
- **THEN** the response is an `INVALID_ARGUMENT` error and any existing task with that id is left intact

#### Scenario: Unknown task
- **WHEN** the client removes or gets a task id that is not registered
- **THEN** the response is a `NOT_FOUND` error

#### Scenario: Start when already running
- **WHEN** the client sends `StartRequest` while the scheduler is running
- **THEN** the response is an `INVALID_STATE` error

### Requirement: Supported schedules
A task SHALL be registered with either a cron schedule (spec, IANA time zone, optional seconds field) or a fixed interval. Custom schedules SHALL NOT be available over the protocol in this version.

#### Scenario: Cron schedule with time zone
- **WHEN** a task is registered with cron `0 9 * * *` and time zone `Europe/Amsterdam`
- **THEN** its runs are due at 09:00 local time, including across daylight saving changes

#### Scenario: Interval schedule
- **WHEN** a task is registered with a one minute interval
- **THEN** its runs are due one minute apart

### Requirement: Registration options
Registration SHALL support maximum drift, not-before and not-after bounds, backfill, a per-task timeout, and flags stating whether the client wants skip and removal notifications. Options SHALL have the same meaning as the corresponding in-process options.

#### Scenario: Drift skipping
- **WHEN** a run would start later than the task's maximum drift after its expected trigger time
- **THEN** the run is not delegated to the client and, if the task requested skip notifications, a `TaskSkipped` notification is sent

#### Scenario: Notifications not requested
- **WHEN** a task registered without skip or removal notifications is skipped or removed
- **THEN** no `TaskSkipped` or `TaskRemoved` message is sent for it

### Requirement: Run delegation
When a run is due, the scheduler SHALL send a `RunRequest` containing a scheduler-unique `run_id`, the task id, the full trigger metadata (trigger window, expected and actual trigger times, last run, last success and last error times, counts, last error) and the run's deadline. The client SHALL answer with exactly one `RunResult` for that `run_id`, either success or a `TaskError` with a kind (returned error, exception, cancelled), message and optional stack trace.

#### Scenario: Successful run
- **WHEN** the client answers a `RunRequest` with success
- **THEN** the scheduler counts the run as successful and schedules the next run

#### Scenario: Failed run
- **WHEN** the client answers with a `TaskError`
- **THEN** the scheduler counts the run as errored, records the message as the task's last error and schedules the next run

#### Scenario: Metadata matches in-process behaviour
- **WHEN** a task runs over the protocol and an equivalent task runs in-process under the same clock and schedule
- **THEN** both receive the same trigger windows and trigger times

#### Scenario: Non-overlapping runs
- **WHEN** a run of a task is still unanswered at the time its next run is due
- **THEN** no second `RunRequest` for that task is sent until the first is answered or cancelled

### Requirement: Run cancellation and timeouts
The scheduler SHALL own all timing. When a run exceeds its timeout, when the scheduler is stopping or when the task is removed or replaced, it SHALL send `CancelRun` with the reason. A run SHALL be treated as finished once cancellation was sent and its timeout elapsed or the scheduler stopped, even if the client never answers; a late `RunResult` SHALL be ignored.

#### Scenario: Timeout
- **WHEN** a run exceeds its timeout
- **THEN** a `CancelRun` with reason `TIMEOUT` is sent and the run is counted as errored with a timeout error

#### Scenario: Unresponsive client
- **WHEN** the client does not answer a cancelled run
- **THEN** the scheduler does not wait for the answer indefinitely and a later `RunResult` for that `run_id` has no effect

#### Scenario: Stop while running
- **WHEN** the scheduler is stopped while runs are delegated
- **THEN** `CancelRun` with reason `STOPPING` is sent for each of them

#### Scenario: Unknown run id
- **WHEN** the scheduler receives a `RunResult` for a `run_id` it does not know
- **THEN** it ignores the message

### Requirement: Task snapshots
Get and list responses SHALL carry task snapshots with the same content as in-process `TaskInfo`, with times that do not apply left unset, and listing SHALL be in a stable order.

#### Scenario: Snapshot after run
- **WHEN** a task has completed one successful run and the client gets the task
- **THEN** the snapshot reports count 1, success count 1, error count 0 and the next run time

### Requirement: Scheduler log forwarding
The scheduler SHALL forward its log records to the client as `LogRecord` notifications with level, message and string fields, honouring the configured log level.

#### Scenario: Log level filtering
- **WHEN** the client initializes with log level `WARN`
- **THEN** no `INFO` or `DEBUG` records are sent

### Requirement: Error codes
Failed requests SHALL return an `Error` with one of the codes `INVALID_ARGUMENT`, `ALREADY_EXISTS`, `NOT_FOUND` and `INVALID_STATE`, and a human-readable message. Enum values SHALL reserve 0 as unspecified.

#### Scenario: Unspecified value
- **WHEN** a request carries an enum field with value 0 where a value is required
- **THEN** the scheduler responds with `INVALID_ARGUMENT`
