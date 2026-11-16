## Purpose

Bounds how many task runs execute at the same time across a scheduler, and defines what happens to runs that become due while the bound is reached.

## ADDED Requirements

### Requirement: Concurrency limit
When configured with a limit of N, the scheduler SHALL never have more than N runs executing at the same time across all tasks. Without a configured limit, the number of concurrent runs SHALL be unbounded.

#### Scenario: Limit respected
- **WHEN** a scheduler with limit 2 has five tasks due at the same time, each running for 10 seconds
- **THEN** at most two runs execute at any moment

#### Scenario: No limit
- **WHEN** a scheduler without a limit has five tasks due at the same time
- **THEN** all five run concurrently

### Requirement: Wait limit mode
In wait mode, a run that becomes due while all slots are busy SHALL be queued and started, in the order the runs became due, when a slot frees. The MaxDrift check SHALL be applied when the run starts, measured from its expected trigger time.

#### Scenario: Queued run starts when a slot frees
- **WHEN** limit is 1 in wait mode, task A runs from 0 to 5 seconds and task B becomes due at 1 second with MaxDrift 30 seconds
- **THEN** B starts at 5 seconds

#### Scenario: Queued too long
- **WHEN** limit is 1 in wait mode, task A runs from 0 to 60 seconds and task B becomes due at 1 second with MaxDrift 30 seconds
- **THEN** B's run is skipped for drift when the slot frees, its skip callback is called, and its next window is scheduled

### Requirement: Skip limit mode
In skip mode, a run that becomes due while all slots are busy SHALL NOT run; its skip callback SHALL be called and the task's next window SHALL be scheduled.

#### Scenario: Run skipped at limit
- **WHEN** limit is 1 in skip mode and task B becomes due while task A is running
- **THEN** B is not run for that window, its skip callback is called, and B's next window is scheduled

### Requirement: Queued runs are dropped on stop and removal
Runs waiting for a slot SHALL be discarded without running when the scheduler stops, or when their task is removed or replaced.

#### Scenario: Stop with queued runs
- **WHEN** the scheduler is stopped while runs are waiting for a slot
- **THEN** those runs never execute and Stop returns after the executing runs return

#### Scenario: Removed while queued
- **WHEN** a task is removed while its run is waiting for a slot
- **THEN** that run never executes
