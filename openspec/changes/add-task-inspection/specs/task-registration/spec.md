## MODIFIED Requirements

### Requirement: Listing tasks
The scheduler SHALL return a snapshot of every registered task, ordered by task ID.

#### Scenario: List registered tasks
- **WHEN** two tasks are registered and ScheduledTasks is called
- **THEN** two task snapshots are returned, carrying the IDs of the two tasks in ascending order
