## ADDED Requirements

### Requirement: Daily schedule
The library SHALL provide a daily schedule that triggers at one or more given times of day, every day, in a given time zone.

#### Scenario: Two times per day
- **WHEN** a daily schedule at 06:00 and 18:00 is evaluated at 12:00
- **THEN** the next run is 18:00 the same day, and the one after that is 06:00 the next day

#### Scenario: Exactly at a scheduled time
- **WHEN** a daily schedule at 06:00 is evaluated with a last run at exactly 06:00
- **THEN** the next run is 06:00 the following day

### Requirement: Weekly schedule
The library SHALL provide a weekly schedule that triggers at one or more given times of day on one or more given weekdays, in a given time zone.

#### Scenario: Selected weekdays
- **WHEN** a weekly schedule for Monday and Thursday at 09:30 is evaluated on a Tuesday
- **THEN** the next run is Thursday at 09:30

### Requirement: Monthly schedule
The library SHALL provide a monthly schedule that triggers at one or more given times of day on one or more given days of the month (1-31), in a given time zone. Months that lack a given day SHALL be skipped for that day.

#### Scenario: First of the month
- **WHEN** a monthly schedule for day 1 at 00:00 is evaluated on 15 March
- **THEN** the next run is 1 April at 00:00

#### Scenario: Day missing in month
- **WHEN** a monthly schedule for day 31 is evaluated on 1 April
- **THEN** the next run is 31 May

### Requirement: Calendar schedule time zones
Calendar schedules SHALL evaluate times of day in the given location, defaulting to the local time zone when none is given. A time that does not exist due to a DST gap SHALL run at the equivalent instant after the gap; a time that occurs twice SHALL run only at its first occurrence.

#### Scenario: Location honoured
- **WHEN** a daily schedule at 09:00 in Europe/Amsterdam is evaluated
- **THEN** the next run is 09:00 Amsterdam time, regardless of the process time zone

#### Scenario: DST gap
- **WHEN** a daily schedule at 02:30 in Europe/Amsterdam is evaluated on the day clocks jump from 02:00 to 03:00
- **THEN** that day's run happens at 03:30 local time and the schedule continues at 02:30 on following days

#### Scenario: DST overlap
- **WHEN** a daily schedule at 02:30 in Europe/Amsterdam is evaluated on the day clocks fall back from 03:00 to 02:00
- **THEN** it runs once that day, at the first 02:30

### Requirement: Calendar schedule validation
Calendar schedule constructors SHALL return an error when no time of day is given, when a time of day is out of range, when no weekday or day of month is given, or when a day of month is outside 1-31.

#### Scenario: Invalid time of day
- **WHEN** a daily schedule is created with hour 24
- **THEN** an error is returned

#### Scenario: Invalid day of month
- **WHEN** a monthly schedule is created with day 0 or 32
- **THEN** an error is returned
