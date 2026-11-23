## MODIFIED Requirements

### Requirement: Scheduler construction
The library SHALL provide a constructor that returns an idle scheduler, configurable through functional options for the logger and the event hooks.

#### Scenario: Default construction
- **WHEN** a scheduler is created without options
- **THEN** it is idle, logs nothing, and has no event hooks

## REMOVED Requirements

### Requirement: Periodic task summaries
**Reason**: Reporting format and cadence are an application concern; the library now exposes event hooks instead.
**Migration**: Remove `WithSummaryInterval`. Count runs, successes and errors in `Hooks.AfterRun` and log or export them on your own schedule.
