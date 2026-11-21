## Context

`ID()` is called many times per task (map Store/Load in `RegisterTask` ~229/261, the main loop's identity checks ~528/545, logging). With `options.ID == ""` each call returns a new random value, so `tasks.Store` and the later `tasks.Load` in the main loop disagree and the unit is dropped as "removed". The default random ID currently comes from `WithDefaultOptions()` (~57).

## Goals / Non-Goals

**Goals:**
- One ID per registration, assigned in exactly one place.
- `WithID("")` behaves like no `WithID`.

**Non-Goals:**
- Returning the generated ID from `RegisterTask` (signature change belongs to `tidy-public-api` / `add-task-inspection`).
- Rejecting or validating other ID values (`validate-task-registration`).
- Making registration atomic (`make-task-registration-atomic`).

## Decisions

- **Generate in `RegisterTask` after applying options**: `if opts.ID == "" { opts.ID = "task-" + rand.Text() }`, before building the `schedulingUnit`. Alternative: keep generation in `WithDefaultOptions` and only patch empty IDs — rejected, two generation sites for the same concern.
- **`ID()` returns `su.options.ID` unconditionally.** It could be replaced by a field, but keeping the method minimises churn for other in-flight changes.
- **Empty ID means "generate" rather than error.** Alternative: return an error — rejected here because `RegisterTask` currently never errors and erroring is the scope of `validate-task-registration`; generating is the least surprising fix.

## Risks / Trade-offs

- [Callers relying on `WithDefaultOptions().ID` being random] → Called out as BREAKING in the proposal; the function is effectively internal.
- [Generated ID not discoverable by the caller] → Unchanged from today; use `WithID` for tasks that must be removed later.
