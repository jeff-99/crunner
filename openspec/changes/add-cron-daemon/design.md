# Design

## Context

The `crunner` library schedules Go functions and has no I/O of its own, no config format, no persistence by default (see `add-state-store`) and no process handling. A cron replacement therefore is a new layer on top: a config loader and reconciler, an executor for external commands, and operational plumbing. See proposal.md for motivation and the list of changes this one depends on. The config's rule that the root package prefers the standard library shapes where the new code lives.

## Goals / Non-Goals

**Goals:**
- Replace cron/anacron for single-host use with declarative, reloadable jobs that can use their trigger window.
- Safe by default when running jobs as other users.
- Zero change to the library; reuse its scheduling, drift, backfill and state semantics unchanged.

**Non-Goals:**
- Importing existing crontabs (a later change); a `last` catch-up mode (can be expressed with max drift for now).
- Handlers other than `exec` (an `http` handler and handlers registered by language clients over the sidecar protocol are later changes).
- Multi-host coordination (see `add-distributed-locking`), run history storage, retries, metrics endpoint.
- Sandboxing beyond uid/gid, umask, environment, `no_new_privs` and the process group (cgroups, namespaces are left to systemd or later).

## Decisions

### Nested module for the daemon
`cmd/crunnerd`, `cmd/crunnerctl` and their packages live in a nested Go module (`daemon/`) that requires `crunner` and a TOML parser. The root module keeps its standard-library-only dependency set. Alternative: same module with the dependency, which would force every library user to download a TOML parser.

### TOML for configuration
No indentation pitfalls, comments, typed durations as strings, well suited for `/etc`. YAML was rejected for its implicit typing and ambiguity; JSON for lacking comments. Unknown keys are errors, so typos do not silently disable options.

### Change detection by polling
The standard library has no inotify. Polling file metadata and content hash every few seconds is portable, dependency-free and good enough for configuration. SIGHUP and `crunnerctl reload` give immediate reload.

### Reconciler compares normalised definitions
Each job is parsed to a normalised struct and hashed. A reload registers jobs whose hash is new, replaces jobs whose hash changed (the library keeps state when the schedule fingerprint is unchanged), and removes jobs that vanished. Unchanged jobs are never re-registered, so timers and running processes are untouched. Per-file and per-job error isolation falls out of keeping the last good definition per job name.

### Job name is the task ID
Task IDs are stable across restarts, which persisted state needs. Renaming a job intentionally starts it fresh.

### Exec: no shell, one argument per template
Templates are rendered after parsing the argument array, once per argument; the result is passed as a single argv element. This removes the injection path of putting `{{.LastError}}` inside a shell string. Shell scripts are allowed only without templates, and run variables reach them through the environment.
Templates use `text/template` with an allow-listed function map and no access to the environment; templates are parsed at validation, so mistakes surface at load, not at run time.

### Dropping privileges
The daemon runs as root and starts the command with `SysProcAttr.Credential` (uid, gid, supplementary groups from the user database) and `Setpgid`, so the process group can be signalled. If any step fails, the run fails; there is no fallback to the daemon's identity. Jobs must name their user when the daemon is root, so nothing runs as root by accident. `no_new_privs` and the umask are applied in the child before exec. The user database is read at run time, so group changes take effect without a reload.

### Termination
On timeout, stop or removal the library cancels the run's context; the handler sends SIGTERM to the process group, waits for the grace period and then sends SIGKILL, and returns only when the group is gone, so the library's non-overlap guarantee holds for real processes, not only for the goroutine.

### Output capture
Pipes feed line-oriented readers that log through the daemon's logger with a per-run byte cap. Pipes (rather than files) avoid the permission problems of log files owned by other users.

### Hooks run inside the task wrapper
The library's success and error callbacks do not receive trigger metadata, but the task function does. The daemon's task wrapper therefore runs the main command, then the matching hook with the same metadata and error, with its own timeout; a failing hook is only logged. `on_skip` uses the library's skip callback, which receives the skipped window.

### State store
The library ships only an in-memory store, so the daemon provides a file-backed `StateStore` (one JSON file per job, written to a temp file, fsynced and renamed). The job-name-to-file mapping is safe because job names are restricted to a small character set. A lock file in the state directory (`flock`) enforces a single instance.

### Control API
HTTP with JSON bodies over a Unix socket, so `curl --unix-socket` works for debugging and `crunnerctl` stays small. Peer credentials come from `SO_PEERCRED` on the accepted connection. Authorisation is by uid: root for mutating requests, plus per-job `allow_run_users` for manual runs.

### systemd integration
`sd_notify` is a datagram to `NOTIFY_SOCKET`, implemented with the standard library. The unit file uses `Type=notify`, `ExecReload=` sending SIGHUP and conservative hardening options.

## Risks / Trade-offs

- [Compromised config directory means arbitrary command execution as any user] → Trust checks on ownership and permissions, no symlinks, documented as the security boundary.
- [Root daemon is a high-value target] → Minimal network surface (no TCP), control socket restricted by peer credentials, no shell by default, clean environment, hardened unit file.
- [Polling delays config changes by up to the interval] → Interval is configurable; reload is immediate on SIGHUP.
- [Crash between a run finishing and its state being saved re-runs the last window] → At-least-once, inherited from the state store; jobs should be idempotent per window.
- [A command ignoring SIGTERM delays the stop] → The kill grace period bounds it, then SIGKILL.
- [Time zone data missing on minimal hosts] → Validation reports the unknown zone at load.
- [Blocking on Linux-specific syscalls] → The daemon targets Linux only; non-Linux builds are not supported in this change.
- [Dependent changes land late] → The tasks are ordered so that parts needing them (catch-up, `run`, jitter, limits) come last.

## Migration Plan

Additive. Install the binaries and unit file, translate crontab entries by hand into job files (an import tool is a later change), run `crunnerd --check`, then enable the unit and disable cron. Rollback is stopping the unit and re-enabling cron.

## Open Questions

- Default location of the control socket when running as a non-root user (for example under `$XDG_RUNTIME_DIR`); can be settled during implementation.
- Packaging formats (deb/rpm) are left to a later change.
