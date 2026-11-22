package crunner

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	// ErrInvalidTask is returned (wrapped) by RegisterTask when the task function is invalid.
	ErrInvalidTask = errors.New("invalid task")
	// ErrInvalidSchedule is returned (wrapped) by RegisterTask when the schedule is invalid.
	ErrInvalidSchedule = errors.New("invalid schedule")
	// ErrInvalidOption is returned (wrapped) by RegisterTask when a scheduling option is invalid.
	ErrInvalidOption = errors.New("invalid option")
)

// validatable is implemented by schedules that can check their own configuration.
type validatable interface {
	validate() error
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Pointer, reflect.Func, reflect.Map, reflect.Slice, reflect.Chan, reflect.Interface:
		return rv.IsNil()
	}
	return false
}

// validateRegistration reports every problem with the registration arguments, joined together.
func validateRegistration(task Task, schedule Schedule, opts SchedulingOptions) error {
	var errs []error

	if task == nil {
		errs = append(errs, fmt.Errorf("%w: task must not be nil", ErrInvalidTask))
	}

	if isNil(schedule) {
		errs = append(errs, fmt.Errorf("%w: schedule must not be nil", ErrInvalidSchedule))
	} else if v, ok := schedule.(validatable); ok {
		if err := v.validate(); err != nil {
			errs = append(errs, fmt.Errorf("%w: %w", ErrInvalidSchedule, err))
		}
	}

	if opts.MaxDrift < 0 {
		errs = append(errs, fmt.Errorf("%w: maxDrift must be >= 0, got %s", ErrInvalidOption, opts.MaxDrift))
	}
	if opts.NotBefore != nil && opts.NotAfter != nil && !opts.NotAfter.After(*opts.NotBefore) {
		errs = append(errs, fmt.Errorf("%w: notAfter (%s) must be after notBefore (%s)", ErrInvalidOption, opts.NotAfter, opts.NotBefore))
	}
	if opts.Backfill && opts.NotBefore == nil {
		errs = append(errs, fmt.Errorf("%w: backfill requires notBefore", ErrInvalidOption))
	}
	if opts.OnSuccess != nil && *opts.OnSuccess == nil {
		errs = append(errs, fmt.Errorf("%w: success callback must not be nil", ErrInvalidOption))
	}
	if opts.OnError != nil && *opts.OnError == nil {
		errs = append(errs, fmt.Errorf("%w: error callback must not be nil", ErrInvalidOption))
	}
	if opts.OnSkip != nil && *opts.OnSkip == nil {
		errs = append(errs, fmt.Errorf("%w: skip callback must not be nil", ErrInvalidOption))
	}
	if opts.OnRemoved != nil && *opts.OnRemoved == nil {
		errs = append(errs, fmt.Errorf("%w: removed callback must not be nil", ErrInvalidOption))
	}

	return errors.Join(errs...)
}
