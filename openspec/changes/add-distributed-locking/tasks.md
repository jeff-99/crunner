# Tasks

## 1. API

- [ ] 1.1 Add `lock.go` with `ErrLockHeld`, `Locker`, `Lock`, a no-op default locker and the key-format helper; verify with a unit test that keys are equal for equal (task ID, window end) regardless of time zone and differ for different window ends
- [ ] 1.2 Add `WithLocker` and default the scheduler to the no-op locker in `New`; verify all existing tests pass unchanged

## 2. Locking around the run

- [ ] 2.1 In the task goroutine, acquire the lock after the drift check and release it after state is saved; verify with a synctest test using a recording locker that Lock precedes the run and Unlock follows Save
- [ ] 2.2 On `ErrLockHeld` (checked with `errors.Is`), skip the run without callbacks or count changes and schedule the next window; verify with a synctest test that two schedulers sharing an in-test locker run each window exactly once in total
- [ ] 2.3 On other lock errors, skip execution, call the error callback, increment the error count and reschedule; verify with a synctest test using a failing locker
- [ ] 2.4 On Unlock error, log and reschedule; verify with a synctest test that the task keeps running

## 3. State interplay

- [ ] 3.1 While holding the lock, load state and skip the window if its recorded end is at or after the current end, then start the next window at the recorded end; verify with a synctest test where two interval-task replicas registered at different times converge on identical windows
- [ ] 3.2 Save state through the lock when it implements `Save`, otherwise through the state store; verify with a synctest test that a lock implementing `Save` receives the state and the store does not

## 4. Docs

- [ ] 4.1 Document multi-replica setup, the non-blocking `ErrLockHeld` contract, the key format, the TTL caveat and a sketch of a transactional Postgres backend in the README/package doc; verify `go vet ./...` passes and `go doc` shows the new symbols
