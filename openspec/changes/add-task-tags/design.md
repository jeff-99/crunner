## Context

Tasks live in `s.tasks` (`sync.Map` keyed by ID). `RemoveTask` deletes, stops the timer under `stateMu`, and calls the cancel callback in a goroutine (`scheduler.go` ~649-674).

## Goals / Non-Goals

**Goals:**
- Simple tag storage and group removal with identical per-task semantics to `RemoveTask`.

**Non-Goals:**
- Tag indexes, tag queries/listing (`add-task-inspection`), tag-based run-now or pausing.
- "All tags must match" semantics.

## Decisions

- **Match "any of the given tags"**, like gocron's `RemoveByTags`. Removing a group by a single tag is the dominant use; "all" can be added later as a separate function without breaking this one.
- **Linear scan over `s.tasks`**, then for each match call the same internal removal helper used by `RemoveTask`, using compare-and-delete so a task replaced concurrently (new unit, maybe different tags) is not removed. Task counts are small; an index would add synchronisation cost for no benefit.
- **Return `int`, no error.** Nothing can fail per the spec; an unknown tag removing nothing is not an error, unlike `RemoveTask` with an unknown ID, because tags denote groups that may legitimately be empty.
- **Tags stored as a deduplicated `[]string` copy** in `SchedulingOptions`, keeping registration order; copied so callers cannot mutate them later.

## Risks / Trade-offs

- [Interface addition breaks mocks] → documented BREAKING; acceptable pre-v1.
- [A task registered concurrently with a tag removal may or may not be removed] → documented as unspecified; callers needing ordering must sequence calls.
