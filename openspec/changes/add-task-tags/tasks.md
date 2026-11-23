## 1. Tags

- [ ] 1.1 Add `Tags` to `SchedulingOptions` and `WithTags` (deduplicate, copy); reject empty tags in `RegisterTask`; verify with unit tests for deduplication, empty-tag error and tag replacement on re-registration

## 2. Removal by tag

- [ ] 2.1 Extract the per-task removal logic from `RemoveTask` into a helper using compare-and-delete; verify existing `RemoveTask` tests still pass
- [ ] 2.2 Add `RemoveTasksByTag(tags ...string) int` to `Scheduler` and implement it with the helper; verify with synctest tests for group removal (cancel callbacks called, other tasks keep running), any-of-several tags, and unknown tag returning 0, under `go test -race ./...`
- [ ] 2.3 Document `WithTags` and `RemoveTasksByTag` in doc comments; verify `go doc` output
