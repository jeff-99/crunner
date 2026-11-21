package crunner

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// Schedule computes the end of a task's next trigger window.
//
// schedulingStart is the start of the window being computed. lastRun is a previous point of the
// task's window sequence (normally the end of the previous window), or nil for the task's first
// window. It is never the time a run actually started, so windows stay on the schedule's grid.
// After a run is cancelled for drift, Next may be called several times in a row, each time with
// the previous result, until the returned time is in the future.
type Schedule interface {
	Next(ctx context.Context, schedulingStart time.Time, lastRun *time.Time) (time.Time, error)
}

type cronSchedule struct {
	Spec     string
	Location *time.Location
	schedule cron.Schedule
}

func ParseCron(spec string, includeSeconds bool) (cron.Schedule, error) {
	parseOpts := cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor
	if includeSeconds {
		parseOpts = cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor
	}

	parser := cron.NewParser(parseOpts)
	schedule, err := parser.Parse(spec)
	if err != nil {
		return nil, fmt.Errorf("invalid cron spec: %w", err)
	}

	return schedule, nil
}

func WithCronSchedule(spec string, location *time.Location, includeSeconds bool) (cronSchedule, error) {
	schedule, err := ParseCron(spec, includeSeconds)
	if err != nil {
		return cronSchedule{}, err
	}

	if location == nil {
		location = time.Local
	}

	return cronSchedule{
		Spec:     spec,
		Location: location,
		schedule: schedule,
	}, nil
}

func (c cronSchedule) Next(ctx context.Context, now time.Time, lastRun *time.Time) (time.Time, error) {
	location := c.Location
	if location == nil {
		location = time.Local
	}

	base := now
	if lastRun != nil {
		base = *lastRun
	}

	return c.schedule.Next(base.In(location)), nil
}

func (c cronSchedule) validate() error {
	if c.schedule == nil {
		return fmt.Errorf("cron schedule is not initialised, use WithCronSchedule")
	}
	return nil
}

type intervalSchedule struct {
	Interval time.Duration
}

func (i intervalSchedule) Next(ctx context.Context, schedulingStart time.Time, lastRun *time.Time) (time.Time, error) {
	if lastRun == nil {
		return schedulingStart.Add(i.Interval), nil
	}

	return (*lastRun).Add(i.Interval), nil
}

func (i intervalSchedule) validate() error {
	if i.Interval <= 0 {
		return fmt.Errorf("interval must be positive, got %s", i.Interval)
	}
	return nil
}

// WithIntervalSchedule returns a schedule whose windows are interval long: the next window ends at
// the reference time (the previous window end) plus interval, or at the scheduling start plus
// interval for the first window.
func WithIntervalSchedule(interval time.Duration) intervalSchedule {
	return intervalSchedule{Interval: interval}
}
