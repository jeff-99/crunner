package crunner

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

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

type intervalSchedule struct {
	Interval time.Duration
}

func (i intervalSchedule) Next(ctx context.Context, schedulingStart time.Time, lastRun *time.Time) (time.Time, error) {
	if lastRun == nil {
		return schedulingStart.Add(i.Interval), nil
	}

	return (*lastRun).Add(i.Interval), nil
}

func WithIntervalSchedule(interval time.Duration) intervalSchedule {
	return intervalSchedule{Interval: interval}
}
