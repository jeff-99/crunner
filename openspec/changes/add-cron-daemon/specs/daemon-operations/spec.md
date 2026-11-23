# Spec Delta

## Purpose

Defines the daemon's operational behaviour on a Linux host: lifecycle, single instance, persisted state, systemd integration, the control socket with its client, and logging with an audit trail.

## ADDED Requirements

### Requirement: Daemon lifecycle
On start the daemon SHALL acquire its instance lock, load configuration, start the scheduler and then signal readiness. On SIGTERM or SIGINT it SHALL stop the scheduler, terminate running jobs as for cancellation, and exit with status 0. It SHALL exit non-zero if it cannot start.

#### Scenario: Graceful termination
- **WHEN** the daemon receives SIGTERM while jobs run
- **THEN** the jobs are terminated and the daemon exits with status 0

#### Scenario: Unusable configuration directory
- **WHEN** the main config file is invalid TOML at start
- **THEN** the daemon logs the error and exits non-zero without starting jobs

### Requirement: Single instance
The daemon SHALL hold an exclusive lock on a file in its state directory for its whole lifetime, and a second instance using the same state directory SHALL exit non-zero immediately.

#### Scenario: Second daemon
- **WHEN** a second daemon is started with the same state directory
- **THEN** it exits non-zero with a message that another instance is running

### Requirement: Persisted state
The daemon SHALL persist each job's window state in its state directory (default `/var/lib/crunner`) using one JSON file per job, written atomically, so that jobs resume after a restart as described by the state store. Unchanged jobs SHALL keep their state across reloads; a removed job's state SHALL NOT be reused by a later job of the same name with a different schedule.

#### Scenario: Restart continuity
- **WHEN** the daemon is stopped, and restarted after a missed run of a job with `catch_up = "all"`
- **THEN** the missed window is run after start

#### Scenario: Atomic state file
- **WHEN** the daemon is killed while saving state
- **THEN** the job's file contains either the previous or the new complete state

### Requirement: systemd integration
When started under systemd with a notification socket, the daemon SHALL send `READY=1` after the scheduler started, `RELOADING=1` and `READY=1` around a reload, and `STOPPING=1` when stopping. An example unit file with reload and hardening options SHALL be provided.

#### Scenario: Ready notification
- **WHEN** the daemon starts with `NOTIFY_SOCKET` set
- **THEN** `READY=1` is sent once jobs are registered

#### Scenario: Not under systemd
- **WHEN** `NOTIFY_SOCKET` is not set
- **THEN** no notification is attempted

### Requirement: Control socket
The daemon SHALL serve a local control API on a Unix socket (default `/run/crunner/control.sock`, mode 0660, configurable group). It SHALL identify the caller by peer credentials. Read-only requests (list, status) SHALL be allowed for every caller that can open the socket. Mutating requests (manual run, reload) SHALL require uid 0, or for manual runs a uid listed in the job's `allow_run_users`.

#### Scenario: Unauthorised manual run
- **WHEN** a user not listed in the job's `allow_run_users` and not root requests a manual run
- **THEN** the request is rejected as forbidden and no run starts

#### Scenario: Read-only access
- **WHEN** any user with socket access lists jobs
- **THEN** the daemon returns them

### Requirement: crunnerctl
`crunnerctl` SHALL provide `list` (jobs with next run and last result), `status <job>` (the job's snapshot, effective definition, last error and any config error), `run <job>` (manual run), `reload`, and an offline `validate` that checks configuration files without a running daemon. Each command SHALL offer `--json` output and use exit status 0 only on success.

#### Scenario: List jobs
- **WHEN** `crunnerctl list` runs against a daemon with two jobs
- **THEN** it prints both with their next run times, in a stable order

#### Scenario: Manual run of a running job
- **WHEN** `crunnerctl run` targets a job that is currently running
- **THEN** it reports that the job is already running and exits non-zero

#### Scenario: Unknown job
- **WHEN** `crunnerctl status nope` targets an unknown job
- **THEN** it reports not found and exits non-zero

#### Scenario: Daemon not running
- **WHEN** the control socket does not exist
- **THEN** `crunnerctl list` reports that the daemon is not running and exits non-zero

### Requirement: Concurrency limit
The daemon SHALL support an optional `max_concurrent_jobs` setting bounding the number of jobs running at once; runs that exceed it wait and are then subject to max drift.

#### Scenario: Limit reached
- **WHEN** the limit is 1 and two jobs are due at once
- **THEN** the second starts after the first finishes, or is skipped if its max drift has passed

### Requirement: Logging and audit
The daemon SHALL log structured records to standard error. For every run it SHALL log a start record and an end record with job name, run window, run-as user, command, outcome (exit status or signal), and duration.

#### Scenario: Audit of a run
- **WHEN** a job runs as `etl` and exits with status 0
- **THEN** a start and an end record are logged with the job, user, command, status 0 and duration
