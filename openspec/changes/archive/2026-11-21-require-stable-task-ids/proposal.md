## Why

`schedulingUnit.ID()` (scheduler.go ~199-206) returns a fresh random ID on every call when the stored ID is empty (`WithID("")`). Such a task is stored under one ID, looked up under another, and is therefore never scheduled and can never be removed. The default ID is also generated inside `WithDefaultOptions()` (~57), so identity is tied to an options constructor rather than to registration.

## What Changes

- A task's ID is fixed once, at registration: if no ID (or an empty ID) is given, a single random `task-` ID is generated in `RegisterTask` and stored on the task; every later lookup uses that stored value.
- `schedulingUnit.ID()` becomes a plain accessor; no random generation outside registration.
- `WithDefaultOptions()` no longer generates a random ID (its `ID` field is empty). **BREAKING** (minor): code reading `WithDefaultOptions().ID` gets `""` instead of a random ID.
- `WithID("")` is treated the same as omitting `WithID`: an ID is generated.

## Capabilities

### New Capabilities

### Modified Capabilities
- `task-registration`: Task identity requirement — ID is assigned once at registration and stays stable; empty ID means "generate".

## Impact

- Code: `scheduler.go` (`WithDefaultOptions`, `schedulingUnit.ID`, `RegisterTask`), tests.
- API: **BREAKING** — `WithDefaultOptions().ID` is now empty. Behaviour of `WithID("")` changes from "broken task" to "generated ID".
- Dependencies on other changes: none. Related: `validate-task-registration` may later choose to reject other invalid IDs; `tidy-public-api` may rename `WithDefaultOptions`; `add-task-inspection` gives callers a way to read a generated ID (until then `RegisterTask` still returns only `error`).
