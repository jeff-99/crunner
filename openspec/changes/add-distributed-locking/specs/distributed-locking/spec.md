# Spec Delta

## Purpose

Lets several scheduler instances (for example replicas of one application) share tasks so that each trigger window of a task is executed by at most one instance at a time, using a user-supplied locking backend.

## ADDED Requirements

### Requirement: Pluggable locker
The library SHALL accept a user-supplied locker through a scheduler option. A locker acquires a lock for a string key, returning either a lock that can be released, a dedicated "lock held" error when another holder owns the key, or another error.

#### Scenario: Custom locker is used
- **WHEN** a scheduler is created with a custom locker and a task becomes due
- **THEN** the scheduler acquires a lock through the custom locker before running the task

### Requirement: No-op default locker
Without a configured locker the scheduler SHALL behave as if every lock is acquired and released successfully.

#### Scenario: No locker configured
- **WHEN** a scheduler is created without a locker
- **THEN** tasks run exactly as without distributed locking

### Requirement: Lock per task window
The scheduler SHALL derive the lock key deterministically from the task ID and the trigger window end only, so that instances computing the same window use the same key.

#### Scenario: Same window on two instances
- **WHEN** two schedulers with a shared locker run a task with the same ID and the same window end
- **THEN** both request a lock with the same key

#### Scenario: Different windows
- **WHEN** the same task becomes due for two different windows
- **THEN** the lock keys differ

### Requirement: Lock held around the run
The scheduler SHALL acquire the lock after the drift check passes and before the task executes, and SHALL release it after the run has returned and its window state has been saved.

#### Scenario: Release after run
- **WHEN** a run that holds the lock returns
- **THEN** the window state is saved first and the lock is released afterwards

#### Scenario: Release failure
- **WHEN** releasing a lock fails
- **THEN** the error is logged and the next window is scheduled as usual

### Requirement: Contended window is skipped
When the locker reports that the lock is held elsewhere, the scheduler SHALL NOT execute the run, SHALL NOT invoke success, error or cancel callbacks, SHALL NOT change run counts, and SHALL schedule the next window.

#### Scenario: Another instance holds the lock
- **WHEN** a task becomes due and its window lock is held by another instance
- **THEN** the task is not executed on this instance and its next window is scheduled

### Requirement: Already-handled window is skipped
While holding the lock, the scheduler SHALL load the task's persisted state; if the recorded window end is at or after the current window end, it SHALL skip the run as for a contended window and start its next window at the recorded window end.

#### Scenario: Window handled by another instance
- **WHEN** an instance acquires the lock for window ending E and the persisted state records a window end at or after E
- **THEN** the run is skipped, the lock is released, and the next window starts at the recorded end

### Requirement: Lock acquisition error
When acquiring the lock fails with an error other than "lock held", the scheduler SHALL NOT execute the run, SHALL call the task's error callback with the error, count the run as errored, and schedule the next window.

#### Scenario: Locker unavailable
- **WHEN** the locker returns a connection error
- **THEN** the task is not executed and the error callback receives that error

### Requirement: Transactional state save through the lock
When the acquired lock can itself save task state, the scheduler SHALL save the window state through the lock instead of through the configured state store.

#### Scenario: Backend implements both
- **WHEN** the lock returned by the locker supports saving state
- **THEN** the window state after the run is saved through the lock and not through the state store
