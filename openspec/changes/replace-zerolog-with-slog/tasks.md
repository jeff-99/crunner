## 1. Logger option and level

- [ ] 1.1 Add exported `LevelTrace = slog.Level(-8)`, change `WithLogger` to accept `*slog.Logger` (nil → discard) and default the field to `slog.New(slog.DiscardHandler)` in `New`; verify `go build ./...` succeeds
- [ ] 1.2 Add a private trace helper guarded by `Logger.Enabled`; verify with a unit test that a handler at debug level receives no trace records

## 2. Migrate log calls

- [ ] 2.1 Rewrite every zerolog call in `scheduler.go` to slog with the same message, level and attribute keys; verify `grep -n zerolog *.go` returns nothing
- [ ] 2.2 Update any test that builds a zerolog logger to use slog; verify `go test -race ./...` passes
- [ ] 2.3 Add tests (using `testing/synctest` to drive time) with a recording handler: default/nil logger emits nothing, a custom logger receives records with a `task` attribute, trace records appear only when the handler level is `LevelTrace`; verify they pass

## 3. Dependencies and docs

- [ ] 3.1 Run `go mod tidy` and verify `github.com/rs/zerolog` no longer appears in `go.mod` or `go.sum`
- [ ] 3.2 Update README/package docs to show `WithLogger(slog.New(...))` and `LevelTrace` usage, and note the breaking change; verify the documented snippet compiles (e.g. as an `Example` test)
