# Spec Delta

## MODIFIED Requirements

### Requirement: Cron schedule
The library SHALL provide a cron schedule built from a standard 5-field cron expression or, optionally, a 6-field expression with seconds, including `@` descriptors, evaluated in a given time zone.

#### Scenario: Valid expression
- **WHEN** a cron schedule is created from `*/5 * * * *`
- **THEN** the task runs at every minute divisible by five

#### Scenario: Seconds field
- **WHEN** a cron schedule is created with seconds enabled from `* * * * * *`
- **THEN** the task runs every second

#### Scenario: Invalid expression
- **WHEN** a cron schedule is created from an invalid expression
- **THEN** an error is returned

#### Scenario: No third-party parser
- **WHEN** the module's dependencies are listed
- **THEN** no third-party cron library is among them

## ADDED Requirements

### Requirement: Cron field count
Without seconds a cron expression SHALL have exactly 5 whitespace-separated fields (minute, hour, day of month, month, day of week); with seconds enabled exactly 6, the first being seconds. Any other count, or an empty expression, SHALL be an error.

#### Scenario: Six fields without seconds
- **WHEN** `* * * * * *` is parsed with seconds disabled
- **THEN** an error is returned

#### Scenario: Five fields with seconds
- **WHEN** `* * * * *` is parsed with seconds enabled
- **THEN** an error is returned

#### Scenario: Extra whitespace
- **WHEN** `0  12 *\t* *` is parsed
- **THEN** it is accepted as 5 fields

### Requirement: Cron field values
Each field SHALL be a comma-separated list of items; empty items are ignored but at least one item is required. An item is `*` or `?`, a value, a range `a-b`, each optionally followed by `/step`. Values are non-negative decimal integers (a leading `+` is allowed).

#### Scenario: Empty list item ignored
- **WHEN** `1,,2 * * * *` is parsed
- **THEN** it matches minutes 1 and 2

#### Scenario: List without values
- **WHEN** `, * * * *` is parsed
- **THEN** an error is returned

#### Scenario: Question mark
- **WHEN** `0 0 ? * ?` is parsed
- **THEN** it is equivalent to `0 0 * * *`

### Requirement: Cron field bounds
Allowed values SHALL be seconds 0-59, minutes 0-59, hours 0-23, day of month 1-31, month 1-12 and day of week 0-6 (0 is Sunday). A value outside its field's bounds, or a range whose start exceeds its end, SHALL be an error.

#### Scenario: Day of week 7
- **WHEN** `* * * * 7` is parsed
- **THEN** an error is returned

#### Scenario: Reversed range
- **WHEN** `* 10-2 * * *` is parsed
- **THEN** an error is returned

### Requirement: Cron names
Month names `jan`-`dec` and day names `sun`-`sat` SHALL be accepted, case-insensitively, wherever a month or day-of-week value is allowed, including as range ends.

#### Scenario: Named weekday range
- **WHEN** `0 9 * * MON-fri` is parsed
- **THEN** it matches 09:00 Monday through Friday

#### Scenario: Name in wrong field
- **WHEN** `0 9 * mon *` is parsed
- **THEN** an error is returned

### Requirement: Cron steps
A step SHALL be a positive integer. `*/n` selects every n-th value from the field minimum, `a-b/n` every n-th value from a to b, and `a/n` every n-th value from a to the field maximum. A step of 0 or more than one `/` or `-` SHALL be an error.

#### Scenario: Start with step
- **WHEN** `5/20 * * * *` is parsed
- **THEN** it matches minutes 5, 25 and 45

#### Scenario: Zero step
- **WHEN** `*/0 * * * *` is parsed
- **THEN** an error is returned

### Requirement: Cron day matching
If the day-of-month or day-of-week field is an unrestricted wildcard (`*` or `?`, without a step greater than 1), a day SHALL match when both fields match; otherwise a day SHALL match when either field matches.

#### Scenario: Both restricted
- **WHEN** `0 0 1 * */2` is evaluated
- **THEN** it fires on the 1st of every month and on every Sunday, Tuesday, Thursday and Saturday

#### Scenario: Day of week only
- **WHEN** `0 0 * * 1` is evaluated
- **THEN** it fires only on Mondays

### Requirement: Cron descriptors
The descriptors `@yearly`, `@annually`, `@monthly`, `@weekly`, `@daily`, `@midnight` and `@hourly` SHALL be accepted (lower case, exact) as 0:00:00 on 1 January, on the 1st of the month, on Sunday, every day, and at minute 0 of every hour respectively, regardless of whether seconds are enabled. Any other `@` word SHALL be an error.

#### Scenario: Weekly
- **WHEN** `@weekly` is evaluated
- **THEN** it fires every Sunday at 00:00:00

#### Scenario: Upper-case descriptor
- **WHEN** `@Daily` is parsed
- **THEN** an error is returned

### Requirement: Every descriptor
`@every <duration>` SHALL be accepted, where duration uses Go duration syntax. The duration is truncated to whole seconds with a minimum of one second, and the next time is the previous time plus the duration, truncated to the second.

#### Scenario: Sub-second duration
- **WHEN** `@every 500ms` is evaluated from 12:00:00
- **THEN** the next time is 12:00:01

#### Scenario: Invalid duration
- **WHEN** `@every soon` is parsed
- **THEN** an error is returned

### Requirement: Cron time-zone prefix
An expression or descriptor MAY be prefixed with `TZ=<name>` or `CRON_TZ=<name>` followed by whitespace, where name is an IANA time-zone name; the expression is then evaluated in that zone, overriding the schedule's location. An unknown zone, or a prefix not followed by an expression, SHALL be an error.

#### Scenario: Prefix overrides location
- **WHEN** a cron schedule is created from `CRON_TZ=UTC 0 0 * * *` with location Europe/Amsterdam
- **THEN** it fires at 00:00 UTC

#### Scenario: Prefix without expression
- **WHEN** `TZ=UTC` is parsed
- **THEN** an error is returned and the parser does not panic

#### Scenario: Unknown zone
- **WHEN** `TZ=Mars/Olympus 0 0 * * *` is parsed
- **THEN** an error is returned

### Requirement: Cron daylight-saving behaviour
By default, matching SHALL use wall-clock time in the evaluation zone. A wall-clock time that does not exist on a day because of a forward DST transition SHALL NOT fire that day; a wall-clock time that occurs twice because of a backward transition SHALL fire at both occurrences.

#### Scenario: Spring forward
- **WHEN** `30 2 * * *` is evaluated in Europe/Amsterdam across 29 March 2026
- **THEN** it fires on 28 March and 30 March at 02:30 but not on 29 March

#### Scenario: Fall back
- **WHEN** `30 2 * * *` is evaluated in Europe/Amsterdam on 25 October 2026
- **THEN** it fires at 02:30 CEST and again at 02:30 CET

### Requirement: DST-safe cron option
The cron schedule SHALL offer an opt-in DST-safe mode. In this mode a wall-clock time that does not exist on a day because of a forward DST transition SHALL fire once on that day at the equivalent instant after the gap, and a wall-clock time that occurs twice because of a backward transition SHALL fire only at its first occurrence. Expressions whose fields do not fix a time inside the affected hour (wildcards and steps over hours or minutes) SHALL behave as in the default mode.

#### Scenario: Gap runs after the transition
- **WHEN** a DST-safe `30 2 * * *` is evaluated in Europe/Amsterdam across 29 March 2026
- **THEN** it fires on 28 March at 02:30, on 29 March at 03:30 CEST, and on 30 March at 02:30

#### Scenario: Overlap fires once
- **WHEN** a DST-safe `30 2 * * *` is evaluated in Europe/Amsterdam on 25 October 2026
- **THEN** it fires once, at 02:30 CEST, and the next run is 26 October at 02:30 CET

#### Scenario: Gap at local midnight
- **WHEN** a DST-safe `30 0 * * *` is evaluated in America/Sao_Paulo on a day whose midnight does not exist
- **THEN** it fires once that day at the equivalent instant after the gap, not on the previous calendar day

#### Scenario: Wildcards unaffected
- **WHEN** a DST-safe `*/10 * * * *` is evaluated across a transition
- **THEN** it fires at the same instants as without the option

#### Scenario: Default unchanged
- **WHEN** the option is not given
- **THEN** the default scenarios of "Cron daylight-saving behaviour" apply

### Requirement: Cron search limit
When an expression has no matching time within five years of the reference time, computing the next run of a cron schedule SHALL return an error.

#### Scenario: Impossible date
- **WHEN** the next run of `0 0 30 2 *` is computed
- **THEN** an error is returned and the task is not scheduled again
