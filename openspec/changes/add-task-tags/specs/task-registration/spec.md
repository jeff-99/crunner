## ADDED Requirements

### Requirement: Task tags
The scheduler SHALL accept zero or more string tags per task at registration and store them, without duplicates, with the task. Registering a task with an empty tag SHALL fail with an error. Replacing a task SHALL replace its tags.

#### Scenario: Tags stored
- **WHEN** a task is registered with tags "tenant-a", "etl" and "etl"
- **THEN** the task has the tags "tenant-a" and "etl"

#### Scenario: Empty tag
- **WHEN** a task is registered with an empty tag
- **THEN** RegisterTask returns an error and the task is not registered

### Requirement: Removing tasks by tag
The scheduler SHALL remove every registered task that has at least one of the given tags, with the same effect per task as removing it by ID, and report how many tasks were removed. Removing by a tag that no task has SHALL remove nothing and not fail.

#### Scenario: Remove a group
- **WHEN** three tasks are registered, two tagged "tenant-a" and one tagged "tenant-b", and tasks are removed by tag "tenant-a"
- **THEN** two tasks are removed, their cancel callbacks are called, and only the "tenant-b" task keeps running

#### Scenario: Any of several tags
- **WHEN** tasks are removed by tags "tenant-a" and "tenant-b"
- **THEN** every task with either tag is removed

#### Scenario: Unknown tag
- **WHEN** tasks are removed by a tag that no task has
- **THEN** zero is reported and no task is affected
