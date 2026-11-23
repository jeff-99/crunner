# Proposal

## Why

crunner's trigger-window semantics (backfill, drift skipping, non-overlapping runs, persisted state) are what cron lacks, but today it can only be used from Go code. A daemon that reads declarative job files from `/etc` and runs commands on schedule makes crunner a drop-in replacement for cron/anacron on a single Linux host, with jobs that can use the trigger window of each run (for example `--from`/`--to` arguments) and run as other users.

## What Changes

- Add a `crunnerd` daemon that loads job definitions from `/etc/crunner/crunner.toml` and `/etc/crunner/conf.d/*.toml`, validates them, registers them with the scheduler, and reconciles on change (SIGHUP, file changes, `crunnerctl reload`) without disturbing unchanged jobs.
- Jobs are declarative: a name, a schedule (cron spec, `@` macros, fixed interval or one-time `at`), time zone, timeout, max drift, catch-up, jitter, bounds, run limit, and an `exec` handler.
- The `exec` handler runs a command (argv by default, no shell) as a configured user and group, in its own process group, with a clean environment, captured output and timeout/stop handling (SIGTERM, then SIGKILL).
- Each run's trigger metadata is exposed to the job as `CRUNNER_*` environment variables and as templates in arguments and environment values. `on_success`, `on_error` and `on_skip` hooks are handlers too and see the same variables.
- Add a control socket and a `crunnerctl` client: `list`, `status`, `run` (manual run), `reload`, and offline `validate`; the daemon also supports `--check`.
- Persist job state across restarts with a file-backed state store (one JSON file per job) shipped with the daemon, a single-instance lock, systemd readiness notification, and an example unit file.
- The daemon and its TOML dependency live in a separate Go module so the root `crunner` library stays standard-library only.
- No changes to the library API or behaviour.

## Capabilities

### New Capabilities
- `daemon-config`: config file layout, job definition fields, validation, trust checks, reload and reconciliation.
- `job-execution`: how the `exec` handler runs a job: identity, environment, templating, process control, output capture and hooks.
- `daemon-operations`: daemon lifecycle, single-instance lock, persisted state, systemd integration, control socket, `crunnerctl`, logging and audit.

### Modified Capabilities
<!-- None: the library's requirements are unchanged; the daemon consumes them. -->

## Impact
- New code: `cmd/crunnerd`, `cmd/crunnerctl`, and supporting packages in a nested module (e.g. `daemon/`), a systemd unit file, an example config, packaging notes.
- New dependency: a TOML parser, in the nested module only.
- No breaking API changes.
- Depends on these other proposed changes, which must land first: `add-state-store` (restart continuity and catch-up), `add-run-now` (manual runs via `crunnerctl run`, the `Manual` flag), `add-schedule-jitter` (`jitter`), `add-concurrency-limit` (`max_concurrent_jobs`), `add-run-limits` (`not_after`, `max_runs`) and `add-one-time-schedule` (`at`, `@reboot`).
- Relates to `add-multi-language-sidecar`: a later change can let language clients register handlers with the daemon over the same protocol.
