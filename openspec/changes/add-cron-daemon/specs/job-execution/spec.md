# Spec Delta

## Purpose

Defines how the daemon's exec handler runs a job's command safely and predictably: identity, environment, templated arguments, process control, output capture and hooks.

## ADDED Requirements

### Requirement: Command without a shell
An exec handler SHALL run its command from an argument array without a shell. Each template in an argument SHALL expand within that single argument, and expanded values SHALL NOT be split or re-interpreted. A job MAY instead select a shell and a script, in which case run variables SHALL be available only through environment variables and templates in the script SHALL be rejected.

#### Scenario: Value with spaces and quotes
- **WHEN** an argument template expands to a value containing spaces, quotes and `;`
- **THEN** the command receives that value as exactly one argument

#### Scenario: Template in a script
- **WHEN** a job with a shell script contains a template in the script text
- **THEN** the job is rejected at validation

### Requirement: Run-as identity
When the daemon runs as root, a job SHALL name the user it runs as, and the command SHALL run with that user's uid, gid and supplementary groups, or with the job's group if set. A job without a user SHALL be rejected. If the identity cannot be applied, the run SHALL fail and the command SHALL NOT run. When the daemon does not run as root, a job's user, if set, SHALL equal the daemon's own.

#### Scenario: Run as another user
- **WHEN** a job names user `etl` and runs
- **THEN** the process runs with `etl`'s uid and groups

#### Scenario: Unknown user
- **WHEN** a job names a user that does not exist
- **THEN** the job is rejected at validation, and if the user disappears later the run fails without executing

#### Scenario: No user as root
- **WHEN** the daemon runs as root and a job has no `user`
- **THEN** the job is rejected, and running as root requires `user = "root"`

### Requirement: Execution environment
The command SHALL start with an empty environment plus `PATH` (a daemon setting with a safe default), `HOME`, `USER`, `LOGNAME` and `SHELL` taken from the run-as user, the job's `env` entries, and the `CRUNNER_*` run variables. It SHALL start in the job's working directory, with the job's umask (default 027), with standard input at `/dev/null`, with no inherited file descriptors other than stdio, and with privilege escalation disabled.

#### Scenario: Clean environment
- **WHEN** the daemon's own environment contains `SECRET=x`
- **THEN** the job's process does not see `SECRET`

#### Scenario: Descriptors not leaked
- **WHEN** the job lists its open file descriptors
- **THEN** only stdin, stdout and stderr are open

### Requirement: Run variables
Each run SHALL expose its trigger metadata and the job name to the command as environment variables prefixed `CRUNNER_` and to templates as CamelCase fields. Times SHALL be RFC 3339 in UTC, and empty when not applicable.

#### Scenario: Complete variable set
- **WHEN** a job runs
- **THEN** it sees `CRUNNER_JOB`, `CRUNNER_WINDOW_START`, `CRUNNER_WINDOW_END`, `CRUNNER_EXPECTED_TIME`, `CRUNNER_ACTUAL_TIME`, `CRUNNER_COUNT`, `CRUNNER_SUCCESS_COUNT`, `CRUNNER_ERROR_COUNT`, `CRUNNER_LAST_RUN`, `CRUNNER_LAST_SUCCESS`, `CRUNNER_LAST_ERROR_TIME` and `CRUNNER_MANUAL`, and templates can use `Job`, `WindowStart`, `WindowEnd`, `ExpectedTime`, `ActualTime`, `Count`, `SuccessCount`, `ErrorCount`, `LastRun`, `LastSuccess`, `LastErrorTime` and `Manual`

#### Scenario: Window in arguments
- **WHEN** a job has argument `--from={{.WindowStart}}` and runs for window [02:00, 03:00) UTC
- **THEN** the process receives `--from=` followed by the RFC 3339 time of 02:00 UTC

#### Scenario: Window in environment
- **WHEN** a job runs for a backfilled window
- **THEN** `CRUNNER_WINDOW_START` and `CRUNNER_WINDOW_END` hold that window, not the current time

#### Scenario: Manual run
- **WHEN** a run is started with `crunnerctl run`
- **THEN** `CRUNNER_MANUAL` is `true`

### Requirement: Template functions
Templates SHALL support only the run fields above and a fixed set of pure formatting functions (time formatting and shifting, upper/lower case, default). Templates SHALL NOT be able to execute commands, read files or access the daemon environment.

#### Scenario: Formatting a time
- **WHEN** a template uses the date formatting function on `WindowStart`
- **THEN** it renders the time in the requested layout

#### Scenario: Unknown function
- **WHEN** a template calls a function outside the supported set
- **THEN** the job is rejected at validation

### Requirement: Process lifecycle
Each run SHALL execute in its own process group. A run SHALL succeed only if the command exits with status 0; a non-zero exit, a termination by signal or a failure to start SHALL be an errored run whose error states the status or signal. When a run is cancelled by timeout, removal or daemon stop, the daemon SHALL send SIGTERM to the whole process group and, if processes remain after the kill grace period (default 10 seconds), SIGKILL.

#### Scenario: Non-zero exit
- **WHEN** the command exits with status 3
- **THEN** the run is errored and its error mentions exit status 3

#### Scenario: Timeout kills children
- **WHEN** a run exceeds its timeout and the command has spawned child processes that ignore SIGTERM
- **THEN** SIGTERM is sent to the group, and after the grace period SIGKILL, and no process of the run remains

#### Scenario: Stop while running
- **WHEN** the daemon is stopped while a job runs
- **THEN** the job's process group is terminated as for a timeout

### Requirement: Output capture
The daemon SHALL capture the command's stdout and stderr line by line into its log, tagged with job name, run window and stream. Capture per run SHALL be limited (default 1 MiB) and long lines truncated, with truncation reported once.

#### Scenario: Output logged
- **WHEN** a job prints a line to stderr
- **THEN** the daemon logs it with the job name and stream `stderr`

#### Scenario: Output limit
- **WHEN** a run prints more than the limit
- **THEN** output beyond the limit is dropped, a truncation notice is logged once, and the run is not failed because of it

### Requirement: Hooks
A job MAY define `on_success`, `on_error` and `on_skip` exec handlers. Hooks SHALL run after the event with the same templates and variables, plus `LastError` and the exit status for `on_error`, with their own timeout (default 1 minute), and by default as the job's user. A hook's failure SHALL be logged and SHALL NOT change the run's outcome or delay the next run.

#### Scenario: Error hook
- **WHEN** a job's command fails and the job defines `on_error`
- **THEN** the hook runs with the error available to its templates

#### Scenario: Skip hook
- **WHEN** a run is skipped for drift and the job defines `on_skip`
- **THEN** the hook runs with the skipped window in its variables

#### Scenario: Failing hook
- **WHEN** an `on_success` hook exits non-zero
- **THEN** the failure is logged and the run remains successful
