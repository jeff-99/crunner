## Why

Many tasks on the same schedule (e.g. `0 * * * *`) across tasks or replicas all fire at the same instant, causing load spikes on shared resources. A per-task random start delay spreads them out without changing what each run covers.

## What Changes

- Add per-task option `WithJitter(max time.Duration)`: each run starts after a uniformly random delay in `[0, max)` after its trigger time.
- Jitter only shifts the actual start; the trigger window and expected trigger time are unchanged.
- The drift allowance for a jittered run becomes MaxDrift + max jitter, so jitter never causes drift skips.
- Backfill catch-up runs are not jittered.
- Negative jitter is rejected by `RegisterTask`.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `task-execution`: "Maximum drift" includes the jitter allowance; a new requirement defines jittered starts.

## Impact

- Code: `scheduler.go` (`SchedulingOptions.Jitter`, option, delay applied when arming the timer in `scheduleTask` ~354-371, drift check in `processTask` ~398-399); tests.
- API: additive; no breaking change. Uses `math/rand/v2` (standard library).
- Dependencies: none required. Interacts textually with `split-cancel-callback` and `make-trigger-windows-contiguous`, which also modify "Maximum drift"; whichever lands later must rebase its delta.
