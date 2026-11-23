# Spec Delta

## Purpose

Defines how the crunner daemon discovers, validates and applies declarative job definitions from configuration files, including live reload that never disturbs unchanged jobs.

## ADDED Requirements

### Requirement: Config layout
The daemon SHALL read daemon settings from one main TOML file (default `/etc/crunner/crunner.toml`) and job definitions from every `*.toml` file in a drop-in directory (default `/etc/crunner/conf.d`), in lexical order. Hidden files and files with other extensions SHALL be ignored. A file MAY define several jobs. Both paths SHALL be overridable by flags.

#### Scenario: Drop-in files
- **WHEN** `conf.d` contains `10-etl.toml`, `20-backup.toml` and `notes.txt`
- **THEN** jobs from the two TOML files are loaded and `notes.txt` is ignored

#### Scenario: Missing drop-in directory
- **WHEN** the drop-in directory does not exist
- **THEN** the daemon starts with no jobs and logs a warning

### Requirement: Job definition
A job SHALL have a unique name matching `[A-Za-z0-9._-]{1,128}` that serves as its task ID, exactly one schedule (cron spec or `@` macro, fixed interval, or one-time `at`), and one handler. It MAY set a time zone, timeout, max drift, catch-up, jitter, not-before, not-after, max runs and an enabled flag, with the meaning of the corresponding scheduler options.

#### Scenario: Valid job
- **WHEN** a job has name `nightly-export`, schedule `0 2 * * *`, time zone `Europe/Amsterdam` and an exec handler
- **THEN** it is registered under task ID `nightly-export` and runs at 02:00 local time

#### Scenario: Several schedules
- **WHEN** a job sets both `schedule` and `every`
- **THEN** the job is rejected as invalid

#### Scenario: Disabled job
- **WHEN** a job sets `enabled = false`
- **THEN** it is validated but not registered, and any registered job of that name is removed

### Requirement: Schedule macros
The daemon SHALL accept `@yearly`, `@annually`, `@monthly`, `@weekly`, `@daily`, `@midnight` and `@hourly` as cron aliases, and `@reboot` as a job that runs once each time the daemon starts.

#### Scenario: Daily macro
- **WHEN** a job uses `@daily`
- **THEN** it runs at midnight in the job's time zone

#### Scenario: Reboot macro
- **WHEN** a job uses `@reboot` and the daemon starts
- **THEN** the job runs once at start, and not again until the next start

### Requirement: Catch-up
A job's catch-up setting SHALL be `none` (default) or `all`. With `none`, windows missed while the daemon was stopped are skipped as drift under the job's max drift. With `all`, every missed window is run in order. Catch-up SHALL rely on persisted state.

#### Scenario: Catch up all
- **WHEN** a job with `catch_up = "all"` and a one-hour schedule was down for three hours
- **THEN** after start it runs once per missed window, in order, each with its own trigger window

#### Scenario: Catch up none
- **WHEN** a job with `catch_up = "none"` was down for three hours
- **THEN** the missed span is skipped and the job continues on schedule

### Requirement: Validation
The daemon SHALL validate every job before registering it: syntax, schedule and time zone, durations, handler fields, template syntax and user resolution. An invalid job SHALL be rejected with an error naming the file and job. Two jobs with the same name SHALL be rejected, keeping the first in load order. One invalid job SHALL NOT prevent other jobs from loading.

#### Scenario: Invalid cron spec
- **WHEN** a job has schedule `61 * * * *`
- **THEN** the job is rejected with an error naming its file and name, and other jobs are loaded

#### Scenario: Duplicate names
- **WHEN** `10-a.toml` and `20-b.toml` both define a job named `sync`
- **THEN** the job from `10-a.toml` is kept and the one from `20-b.toml` is rejected with an error naming both files

#### Scenario: Syntax error in a file
- **WHEN** a file is not valid TOML
- **THEN** no job from that file is applied and an error is logged, while other files are loaded

### Requirement: Check mode
The daemon SHALL support a check mode that loads and validates all configuration, prints every error, starts nothing, and exits non-zero if any error was found.

#### Scenario: Check with errors
- **WHEN** `crunnerd --check` runs and one job is invalid
- **THEN** the error is printed and the exit status is non-zero

#### Scenario: Check clean
- **WHEN** `crunnerd --check` runs on valid configuration
- **THEN** it exits with status 0 and no job runs

### Requirement: Config trust
When the daemon runs as root, it SHALL load only configuration files and directories that are owned by root and not writable by group or others, and SHALL NOT follow symbolic links in the drop-in directory. Violations SHALL be reported and the affected file ignored.

#### Scenario: World-writable file
- **WHEN** a file in `conf.d` is writable by others
- **THEN** it is ignored and an error naming the file and the violation is logged

#### Scenario: Symlink in drop-in directory
- **WHEN** `conf.d` contains a symbolic link to a TOML file
- **THEN** the link is ignored with an error

### Requirement: Reload triggers
The daemon SHALL reload configuration on SIGHUP, on `crunnerctl reload`, and when it detects that a configuration file was added, changed or removed, within a configurable poll interval (default 5 seconds).

#### Scenario: File edited
- **WHEN** a job file is edited while the daemon runs
- **THEN** the change is applied within the poll interval without a restart

#### Scenario: SIGHUP
- **WHEN** the daemon receives SIGHUP
- **THEN** configuration is reloaded immediately

### Requirement: Reconciliation
A reload SHALL compare the desired jobs with the registered ones. New jobs SHALL be registered, removed or disabled jobs removed, and jobs whose definition changed replaced. Jobs whose definition is unchanged SHALL NOT be touched: they keep their timers, running runs and persisted state.

#### Scenario: Unchanged job during reload
- **WHEN** a job is running and an unrelated file is edited and reloaded
- **THEN** the running job continues uninterrupted and its schedule is unchanged

#### Scenario: Changed schedule
- **WHEN** a job's schedule is edited
- **THEN** the job is replaced and runs on the new schedule

#### Scenario: Removed job
- **WHEN** a job's definition is deleted
- **THEN** the job is removed, its running process is stopped as for a removal, and it no longer runs

### Requirement: Last good configuration
A reload that finds a job invalid SHALL keep that job's last good version registered and report the error. A file that fails to parse SHALL keep its previously loaded jobs. Reload errors SHALL be available through `crunnerctl`.

#### Scenario: Bad edit
- **WHEN** a valid job is edited into an invalid one and reloaded
- **THEN** the previous version keeps running and the error is logged and visible through `crunnerctl status`
