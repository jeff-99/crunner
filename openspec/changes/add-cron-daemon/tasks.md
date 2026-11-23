# Tasks

## 1. Module scaffold

- [ ] 1.1 Create the nested module (`daemon/go.mod`) requiring the root module and a TOML parser, with empty `cmd/crunnerd` and `cmd/crunnerctl`, and verify `go build ./...` works in both modules and `go list -deps` of the root module shows no TOML dependency

## 2. Config loading and validation

- [ ] 2.1 Implement the config model, TOML decoding with unknown-key rejection, and the drop-in directory discovery, and verify unit tests for ordering, ignored files, multiple jobs per file and missing directory
- [ ] 2.2 Implement job validation (name, exactly one schedule, macros, time zone, durations, catch-up, handler fields, user requirement, template parsing) with errors naming file and job, and verify table-driven tests for each rule
- [ ] 2.3 Implement duplicate-name handling and per-job/per-file error isolation, and verify tests that one bad job or file leaves the others loaded
- [ ] 2.4 Implement the config trust checks (ownership, permissions, symlinks) when running as root, and verify tests using temporary files and a simulated owner
- [ ] 2.5 Implement `crunnerd --check` and verify tests for exit status with and without errors and that nothing is started

## 3. Reconciliation and reload

- [ ] 3.1 Normalise and hash job definitions, and verify tests that equivalent definitions hash equally and any field change alters the hash
- [ ] 3.2 Implement the reconciler against a fake scheduler (register, replace, remove, untouched), and verify tests for add, change, remove, disable and unchanged-while-running
- [ ] 3.3 Implement last-good retention and surfacing of reload errors, and verify tests for a bad edit keeping the previous version
- [ ] 3.4 Implement reload triggers (poll, SIGHUP, control request) with a configurable interval, and verify tests with a fake clock and a real temporary directory
- [ ] 3.5 Map jobs to scheduler options (`catch_up`, jitter, bounds, max runs, `@reboot`, `at`) once the dependent library changes are available, and verify with `testing/synctest` tests that windows, backfill and skips match the library's behaviour

## 4. Exec handler

- [ ] 4.1 Implement argv execution with templating (one value per argument, allow-listed functions) and the run variables as environment and template fields, and verify tests with values containing spaces, quotes and `;`
- [ ] 4.2 Implement the execution environment (clean env, `PATH`/`HOME`/`USER` from the user, working directory, umask, stdin from `/dev/null`, no inherited descriptors), and verify a test that prints its environment and open descriptors
- [ ] 4.3 Implement run-as identity (uid, gid, supplementary groups, `no_new_privs`) failing closed, and verify tests that run as root in a container (skipped otherwise) and tests for unknown user and failure to switch
- [ ] 4.4 Implement process groups and termination (SIGTERM, grace period, SIGKILL, return only when the group is gone) wired to the context, and verify tests with children that ignore SIGTERM
- [ ] 4.5 Implement exit status mapping and output capture with the byte cap and truncation notice, and verify tests for non-zero exit, signal death, start failure and output limit
- [ ] 4.6 Implement `on_success`, `on_error` and `on_skip` hooks with their own timeout and isolation, and verify tests that hook failures do not change outcomes or delay the next run

## 5. Operations

- [ ] 5.1 Implement the file-backed state store (atomic writes, safe names) and the instance lock, and verify tests for crash-simulated writes, restart continuity and a second instance exiting non-zero
- [ ] 5.2 Implement the daemon lifecycle (startup order, SIGTERM/SIGINT, exit codes) and `sd_notify`, and verify tests with a fake notification socket
- [ ] 5.3 Implement the control socket with peer-credential authorisation and the list, status, run and reload endpoints, and verify tests for allowed, forbidden and unknown-job requests
- [ ] 5.4 Implement `crunnerctl` (`list`, `status`, `run`, `reload`, offline `validate`, `--json`, exit codes), and verify tests against a running test daemon
- [ ] 5.5 Implement `max_concurrent_jobs`, structured logging and the start/end audit records, and verify tests for the limit and the logged fields
- [ ] 5.6 Run the whole daemon module under `go test -race ./...` and verify it passes

## 6. Packaging and documentation

- [ ] 6.1 Add the example systemd unit and example configuration, and verify `systemd-analyze verify` accepts the unit and `crunnerd --check` accepts the example config
- [ ] 6.2 Add an end-to-end test that starts the daemon against a temporary config and state directory, runs a job, edits the config, reloads, restarts, and verifies catch-up, and verify it passes in CI
- [ ] 6.3 Write the user documentation (config reference, variables, security model, migration from cron) and verify every example in it passes `crunnerd --check`

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied and the changes it depends on have landed.
- Follow-up changes (not in this one): crontab import, an `http` handler, handlers registered by language clients, metrics, deb/rpm packaging.
