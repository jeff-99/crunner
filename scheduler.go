package crunner

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

const (
	stateIdle int32 = iota
	stateRunning
	stateStopping
)

type SchedulerOption func(*internalScheduler)

func WithLogger(logger zerolog.Logger) SchedulerOption {
	return func(o *internalScheduler) {
		o.logger = logger
	}
}

func WithSummaryInterval(d time.Duration) SchedulerOption {
	return func(o *internalScheduler) {
		o.summaryInterval = d
	}
}

type (
	SuccessCallback func(ctx context.Context)
	ErrorCallback   func(ctx context.Context, err error)
	CancelCallback  func(ctx context.Context)
)

type SchedulingOptions struct {
	MaxDrift  time.Duration
	NotAfter  *time.Time
	NotBefore *time.Time
	Backfill  bool

	ID        string
	OnSuccess *SuccessCallback
	OnError   *ErrorCallback
	OnCancel  *CancelCallback
}

type SchedulingOption func(*SchedulingOptions)

func WithDefaultOptions() SchedulingOptions {
	return SchedulingOptions{
		MaxDrift: 30 * time.Second,
		ID:       "task-" + rand.Text(),
		Backfill: false,
	}
}

func WithMaxDrift(d time.Duration) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.MaxDrift = d
	}
}

func WithNotAfter(t time.Time) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.NotAfter = &t
	}
}

func WithNotBefore(t time.Time) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.NotBefore = &t
	}
}

func WithBackfill() SchedulingOption {
	return func(so *SchedulingOptions) {
		so.Backfill = true
	}
}

func WithID(id string) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.ID = id
	}
}

func WithSuccessCallback(f SuccessCallback) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.OnSuccess = &f
	}
}

func WithErrorCallback(f ErrorCallback) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.OnError = &f
	}
}

func WithCancelCallback(f CancelCallback) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.OnCancel = &f
	}
}

type Scheduler interface {
	RegisterTask(task Task, schedule Schedule, options ...SchedulingOption) error
	RemoveTask(string) error
	ScheduledTasks() []Task
	Start(ctx context.Context) error
	Stop() error
}

type TriggerMetadata struct {
	ExpectedTriggerTime time.Time
	ActualTriggerTime   time.Time

	TriggerWindowStart time.Time
	TriggerWindowEnd   time.Time

	LastSuccessfulRunTime *time.Time
	LastErroredRunTime    *time.Time
	LastRunTime           *time.Time

	Count           int
	SuccessfulCount int
	ErrorCount      int

	LastError error
}

type internalScheduler struct {
	wg           *sync.WaitGroup
	cancelCtx    context.CancelFunc
	schedulerCtx context.Context
	logger       zerolog.Logger

	tasks sync.Map

	newTaskCh chan *schedulingUnit
	tasksIn   chan *schedulingUnit

	schedulingEpoch *time.Time

	summaryInterval time.Duration

	state   atomic.Int32
	stateMu sync.Mutex
}

type schedulingUnit struct {
	task     Task
	schedule Schedule

	meta      *TriggerMetadata
	options   SchedulingOptions
	createdAt time.Time

	timer *time.Timer

	backfilling bool

	onError   *ErrorCallback
	onSuccess *SuccessCallback
	onCancel  *CancelCallback

	// snapshot of counts at last summary report (read/written only from mainLoop goroutine)
	lastReportedCount        int
	lastReportedSuccessCount int
	lastReportedErrorCount   int
}

func New(opts ...SchedulerOption) Scheduler {
	// Use buffered channels to prevent blocking
	const channelBufferSize = 100

	scheduler := &internalScheduler{
		wg:              &sync.WaitGroup{},
		tasks:           sync.Map{},
		newTaskCh:       make(chan *schedulingUnit, channelBufferSize),
		tasksIn:         make(chan *schedulingUnit, channelBufferSize),
		logger:          zerolog.Nop(), // Initialize with Nop logger by default
		summaryInterval: 30 * time.Second,
	}
	scheduler.state.Store(stateIdle) // Initialize state

	// Apply options, potentially overwriting the logger
	for _, opt := range opts {
		opt(scheduler)
	}

	return scheduler
}

func (su *schedulingUnit) ID() string {
	id := su.options.ID
	if id == "" {
		return "task-" + rand.Text()
	} else {
		return id
	}
}

func (s *internalScheduler) RegisterTask(task Task, schedule Schedule, options ...SchedulingOption) error {
	opts := WithDefaultOptions()
	for _, opt := range options {
		opt(&opts)
	}

	newSchedulingUnit := &schedulingUnit{
		task:      task,
		schedule:  schedule,
		options:   opts,
		meta:      &TriggerMetadata{},
		createdAt: time.Now(),

		onSuccess: opts.OnSuccess,
		onError:   opts.OnError,
		onCancel:  opts.OnCancel,
	}

	// --- Overwrite Handling ---
	// Before storing the new unit, check if one already exists with the same ID.
	// If so, stop its timer and call its cancel callback.
	if existingValue, loaded := s.tasks.Load(newSchedulingUnit.ID()); loaded {
		s.logger.Trace().Str("task", newSchedulingUnit.ID()).Msg("Overwriting existing task registration")
		if existingUnit, ok := existingValue.(*schedulingUnit); ok {
			// Use stateMu to prevent race with Stop() iterating timers
			s.stateMu.Lock()
			// Check state again after acquiring lock
			currentState := s.state.Load()
			if currentState == stateRunning || currentState == stateStopping {
				if existingUnit.timer != nil {
					stopped := existingUnit.timer.Stop()
					s.logger.Trace().Str("task", newSchedulingUnit.ID()).Bool("timerStopped", stopped).Msg("Stopped timer for overwritten task")
				}
			}
			s.stateMu.Unlock()

			// Call the cancel callback of the *overwritten* task, if it exists
			if existingUnit.onCancel != nil {
				callbackCtx := context.Background() // Use background context for the callback
				if s.schedulerCtx != nil {
					// Prefer scheduler context if available, but don't block registration on it
					// Consider if a derived context with a short timeout is better here.
					callbackCtx = s.schedulerCtx
				}
				s.logger.Trace().Str("task", existingUnit.ID()).Msg("Calling cancel callback for overwritten task")
				f := *existingUnit.onCancel
				go f(callbackCtx) // Run in a goroutine to avoid blocking registration
			}
		}
	}
	// --- End Overwrite Handling ---

	// Store the new or updated scheduling unit.
	s.tasks.Store(newSchedulingUnit.ID(), newSchedulingUnit)

	// Send the new unit to the main loop for scheduling its *first* run (or rescheduling if Start was already called).
	// Only send if the scheduler is actually running.
	if s.state.Load() == stateRunning {
		select {
		case <-s.schedulerCtx.Done():
			s.logger.Warn().Str("task", newSchedulingUnit.ID()).Msg("Scheduler context done, cannot schedule newly registered task")
		case s.newTaskCh <- newSchedulingUnit:
			s.logger.Trace().Str("task", newSchedulingUnit.ID()).Msg("New/overwritten task sent for scheduling")
		}
	} else {
		s.logger.Trace().Str("task", newSchedulingUnit.ID()).Msg("Scheduler not running, task registered but not scheduled yet")
	}

	return nil
}

func (s *internalScheduler) scheduleTask(su *schedulingUnit) {
	if s.state.Load() != stateRunning {
		s.logger.Trace().Str("task", su.ID()).Msg("Scheduler not running, skipping task scheduling")
		return
	}

	s.logger.Trace().Str("task", su.ID()).Msg("Scheduling task")
	ctx, cancel := context.WithTimeout(s.schedulerCtx, 1*time.Minute)
	defer cancel()

	previousRunMeta := su.meta
	isBackfill := su.options.Backfill && su.options.NotBefore != nil
	hasCaughtUp := false
	now := time.Now()

	var triggerWindowStart, triggerWindowEnd, expectedTriggerTime time.Time
	var calculationBaseTime time.Time
	var calculationEpoch time.Time

	if !previousRunMeta.TriggerWindowEnd.IsZero() {
		// has already run once and we are not backfilling
		calculationBaseTime = previousRunMeta.TriggerWindowEnd // TODO this looks wrong, but breaks the tests if I change it to the schudulingEpoch (which seems right?)
	} else if isBackfill {
		calculationBaseTime = *su.options.NotBefore
	} else {
		// if s.schedulingEpoch == nil {
		// 	s.logger.Error().Str("task", su.ID()).Msg("BUG: schedulingEpoch is nil during scheduleTask. Scheduler might not have been started correctly.")
		// 	calculationBaseTime = now
		// } else {
		// 	calculationBaseTime = *s.schedulingEpoch
		// }
		calculationBaseTime = su.createdAt
	}
	calculationEpoch = calculationBaseTime

	// Determine the appropriate last run time to pass to schedule.Next
	var lastRunPtr *time.Time
	if isBackfill {
		// When backfilling, we want to continue from where we left off (TriggerWindowEnd)
		// but only if we've already run at least once
		if !previousRunMeta.TriggerWindowEnd.IsZero() {
			lastRunPtr = &previousRunMeta.TriggerWindowEnd
		}
	} else {
		// For normal scheduling, use the actual LastRunTime if available
		lastRunPtr = previousRunMeta.LastRunTime
	}

	nextWindowEnd, err := su.schedule.Next(ctx, calculationEpoch, lastRunPtr)
	if err != nil {
		s.logger.Error().Err(err).Str("task", su.ID()).Time("calculationBaseTime", calculationBaseTime).Msg("Failed to get next run time from schedule")
		return
	}

	if isBackfill && nextWindowEnd.After(now) {
		hasCaughtUp = true
	}

	triggerWindowStart = calculationBaseTime
	triggerWindowEnd = nextWindowEnd

	if isBackfill && !hasCaughtUp {
		expectedTriggerTime = now
		su.backfilling = true
		s.logger.Trace().Str("task", su.ID()).Time("windowStart", triggerWindowStart).Time("windowEnd", triggerWindowEnd).Msg("Task is backfilling")
	} else {
		expectedTriggerTime = triggerWindowEnd
		su.backfilling = false
		s.logger.Trace().Str("task", su.ID()).Time("windowStart", triggerWindowStart).Time("windowEnd", triggerWindowEnd).Msg("Task is normal or backfill caught up")
	}

	su.meta.TriggerWindowStart = triggerWindowStart
	su.meta.TriggerWindowEnd = triggerWindowEnd
	su.meta.ExpectedTriggerTime = expectedTriggerTime

	fireAfter := expectedTriggerTime.Sub(now)

	if su.backfilling && fireAfter < 0 {
		fireAfter = 0
	} else if !su.backfilling && fireAfter < 0 {
		s.logger.Warn().Str("task", su.ID()).Time("expectedTriggerTime", expectedTriggerTime).Dur("fireAfter", fireAfter).Msg("Normal task's expected trigger time is in the past. Firing immediately.")
		fireAfter = 0
	} else if fireAfter < 0 {
		s.logger.Error().Str("task", su.ID()).Time("expectedTriggerTime", expectedTriggerTime).Dur("fireAfter", fireAfter).Msg("Unexpected negative fireAfter duration. Setting to 0.")
		fireAfter = 0
	}

	s.logger.Debug().Str("task", su.ID()).Str("fireAfter", fireAfter.String()).Time("expectedTriggerTime", expectedTriggerTime).Time("now", now).Bool("backfilling", su.backfilling).Time("windowStart", triggerWindowStart).Time("windowEnd", triggerWindowEnd).Msg("Scheduling task timer")

	if su.timer != nil {
		su.timer.Stop()
	}
	taskTimer := time.AfterFunc(fireAfter, func() {
		fireTime := time.Now()
		if s.state.Load() != stateRunning {
			s.logger.Trace().Str("task", su.ID()).Msg("Timer fired but scheduler not running.")
			return
		}
		s.logger.Trace().Str("task", su.ID()).Time("fireTime", fireTime).Time("expectedTriggerTime", expectedTriggerTime).Msg("Task timer fired")

		select {
		case <-s.schedulerCtx.Done():
			s.logger.Trace().Str("task", su.ID()).Msg("Scheduler context done, not queueing task")
			return
		case s.tasksIn <- su:
			s.logger.Trace().Str("task", su.ID()).Msg("Task queued for execution")
		}
	})

	su.timer = taskTimer
}

func (s *internalScheduler) processTask(su *schedulingUnit) {
	ctx, cancel := context.WithCancel(s.schedulerCtx)
	// defer cancel() // this cancels to early as the task has a derived context in a goroutine

	s.logger.Trace().Str("task", su.ID()).Time("windowStart", su.meta.TriggerWindowStart).Time("windowEnd", su.meta.TriggerWindowEnd).Msg("Processing task")
	now := time.Now()

	drift := now.Sub(su.meta.ExpectedTriggerTime)
	mayExecute := drift <= su.options.MaxDrift

	if !mayExecute {
		s.logger.Warn().Str("task", su.ID()).Time("expectedTriggerTime", su.meta.ExpectedTriggerTime).Time("actualTime", now).Dur("drift", drift).Dur("maxDrift", su.options.MaxDrift).Msg("Task execution skipped due to excessive drift")

		// issue when a task is skipped due to excessive drift the task will be rescheduled, but that task will also be skipped. So the task will never run again.
		// so we need to set the last run time to now, so that the next run will be scheduled correctly.
		su.meta.LastRunTime = &now

		if su.onCancel != nil {
			f := *su.onCancel
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				defer cancel()
				f(ctx)
			}()
		} else {
			cancel()
		}
		s.scheduleTask(su)
		return
	}

	su.meta.ActualTriggerTime = now

	s.logger.Trace().Str("task", su.ID()).Msg("Task running")

	s.wg.Add(1)
	go func(taskUnit *schedulingUnit) {
		defer s.wg.Done()
		defer cancel()
		metaCopy := *taskUnit.meta
		var err error
		func() {
			taskCtx, taskCancel := context.WithTimeout(ctx, 30*24*time.Hour)
			defer taskCancel()
			err = taskUnit.task(taskCtx, metaCopy)
		}()

		finished := time.Now()

		taskUnit.meta.LastRunTime = &now
		taskUnit.meta.Count++

		if err != nil {
			s.logger.Error().Err(err).Str("task", taskUnit.ID()).Msg("Task execution failed")
			taskUnit.meta.LastErroredRunTime = &now
			taskUnit.meta.ErrorCount++
			taskUnit.meta.LastError = err
			if taskUnit.onError != nil {
				f := *taskUnit.onError
				f(ctx, err)
			}
		} else {
			s.logger.Debug().Str("task", taskUnit.ID()).Str("finished_time", finished.Format(time.RFC3339)).Str("start_time", now.Format(time.RFC3339)).Str("duration", finished.Sub(now).String()).Msgf("Task finished successfully")
			taskUnit.meta.LastSuccessfulRunTime = &now
			taskUnit.meta.SuccessfulCount++
			taskUnit.meta.LastError = nil
			if taskUnit.onSuccess != nil {
				f := *taskUnit.onSuccess
				f(ctx)
			}
		}

		select {
		case <-s.schedulerCtx.Done():
			s.logger.Trace().Str("task", taskUnit.ID()).Msg("Scheduler context done, not rescheduling task.")
		case s.newTaskCh <- taskUnit:
			s.logger.Trace().Str("task", taskUnit.ID()).Msg("Task sent for rescheduling.")
		}
	}(su)
}

func (s *internalScheduler) emitTaskSummaries() {
	s.tasks.Range(func(key, value interface{}) bool {
		su := value.(*schedulingUnit)
		meta := su.meta

		runsDelta := meta.Count - su.lastReportedCount
		successDelta := meta.SuccessfulCount - su.lastReportedSuccessCount
		errorDelta := meta.ErrorCount - su.lastReportedErrorCount

		// Only log if there was any activity in this period
		if runsDelta > 0 {
			logEvent := s.logger.Info().
				Str("task", su.ID()).
				Int("runs", runsDelta).
				Int("success", successDelta).
				Int("errors", errorDelta).
				Int("totalRuns", meta.Count).
				Int("totalSuccess", meta.SuccessfulCount).
				Int("totalErrors", meta.ErrorCount)

			if meta.LastError != nil {
				logEvent = logEvent.Str("lastError", meta.LastError.Error())
			}

			logEvent.Msg("Task summary")
		}

		su.lastReportedCount = meta.Count
		su.lastReportedSuccessCount = meta.SuccessfulCount
		su.lastReportedErrorCount = meta.ErrorCount

		return true
	})
}

func (s *internalScheduler) mainLoop() {
	s.logger.Trace().Msg("Scheduler main loop started")
	defer s.logger.Trace().Msg("Scheduler main loop stopped") // Log when loop exits

	summaryTicker := time.NewTicker(s.summaryInterval)
	defer summaryTicker.Stop()

	for {
		select {
		// Case 1: Context canceled (Stop() was called)
		case <-s.schedulerCtx.Done():
			s.logger.Trace().Msg("Mainloop: Context canceled, shutting down")
			return // Exit loop

		// Case 2: Periodic task summary
		case <-summaryTicker.C:
			s.emitTaskSummaries()

		// Case 3: New task registration or reschedule request (from RegisterTask or rescheduling after execution)
		case receivedUnit := <-s.newTaskCh:
			// Check if the received unit is still the one registered for its ID.
			if currentVal, ok := s.tasks.Load(receivedUnit.ID()); ok {
				if currentUnit, isUnit := currentVal.(*schedulingUnit); isUnit && currentUnit == receivedUnit {
					// Only schedule if the received unit is the currently registered one.
					s.logger.Trace().Str("task", receivedUnit.ID()).Msg("Mainloop: Received schedule/reschedule request")
					s.scheduleTask(receivedUnit) // Calculate next run time and set timer
				} else {
					// Task was overwritten before this schedule request was processed.
					s.logger.Trace().Str("task", receivedUnit.ID()).Msg("Mainloop: Received schedule request for overwritten task, ignoring")
				}
			} else {
				// Task was removed before this schedule request was processed.
				s.logger.Trace().Str("task", receivedUnit.ID()).Msg("Mainloop: Received schedule request for removed task, ignoring")
			}

		// Case 4: Task timer fired, ready for processing
		case su := <-s.tasksIn:
			// Check if the received unit 'su' is still the one registered for its ID.
			if currentVal, ok := s.tasks.Load(su.ID()); ok {
				if currentUnit, isUnit := currentVal.(*schedulingUnit); isUnit && currentUnit == su {
					// Only process if the received unit is the currently registered one.
					s.logger.Trace().Str("task", su.ID()).Msg("Mainloop: Received ready task")
					s.processTask(su) // Process the task (check drift, run in goroutine)
				} else {
					// The task was overwritten between timer firing and processing. Ignore.
					s.logger.Trace().Str("task", su.ID()).Msg("Mainloop: Received ready task for overwritten task, ignoring")
				}
			} else {
				// The task was removed between timer firing and processing. Ignore.
				s.logger.Trace().Str("task", su.ID()).Msg("Mainloop: Received ready task for removed task, ignoring")
			}
		}
	}
}

func (s *internalScheduler) Start(ctx context.Context) error {
	if !s.state.CompareAndSwap(stateIdle, stateRunning) {
		currentState := s.state.Load()
		if currentState == stateRunning {
			s.logger.Trace().Msg("Scheduler already running")
			return fmt.Errorf("scheduler already running")
		}
		s.logger.Warn().Int32("state", currentState).Msg("Scheduler cannot start from current state")
		return fmt.Errorf("scheduler cannot start from state %d", currentState)
	}

	now := time.Now()
	s.schedulingEpoch = &now
	s.logger.Trace().Msg("Starting scheduler")

	context, cancel := context.WithCancel(ctx)
	s.cancelCtx = cancel
	s.schedulerCtx = context

	//nolint:contextcheck
	go s.mainLoop()

	s.tasks.Range(func(key, value interface{}) bool {
		schedulingUnit := value.(*schedulingUnit)
		select {
		case <-s.schedulerCtx.Done():
			s.logger.Trace().Msg("Scheduler context done during initial task registration")
			return false
		case s.newTaskCh <- schedulingUnit:
			s.logger.Trace().Str("task", schedulingUnit.ID()).Msg("Existing task queued for scheduling on start")
		}
		return true
	})

	return nil
}

func (s *internalScheduler) Stop() error {
	if !s.state.CompareAndSwap(stateRunning, stateStopping) {
		currentState := s.state.Load()
		if currentState == stateIdle || currentState == stateStopping {
			s.logger.Trace().Int32("state", currentState).Msg("Scheduler already stopped or stopping")
			return nil
		}
		s.logger.Warn().Int32("state", currentState).Msg("Scheduler cannot stop from current state")
		return fmt.Errorf("scheduler cannot stop from state %d", currentState)
	}

	s.logger.Trace().Msg("Stopping scheduler")

	if s.cancelCtx != nil {
		s.cancelCtx()
	}

	s.stateMu.Lock()
	s.tasks.Range(func(key, value interface{}) bool {
		schedulingUnit := value.(*schedulingUnit)
		if schedulingUnit.timer != nil {
			schedulingUnit.timer.Stop()
		}
		return true
	})
	s.stateMu.Unlock()

	s.wg.Wait()

	s.state.Store(stateIdle)
	s.logger.Trace().Msg("Scheduler stopped")

	return nil
}

func (s *internalScheduler) ScheduledTasks() []Task {
	tasks := make([]Task, 0)
	s.tasks.Range(func(key, value interface{}) bool {
		tasks = append(tasks, value.(*schedulingUnit).task)
		return true
	})
	return tasks
}

func (s *internalScheduler) RemoveTask(id string) error {
	taskValue, loaded := s.tasks.Load(id)
	if !loaded {
		return fmt.Errorf("task with id %s not found", id)
	}
	su := taskValue.(*schedulingUnit)

	s.tasks.Delete(id)
	s.logger.Trace().Str("task", id).Msg("Removed task from map")

	s.stateMu.Lock()
	currentState := s.state.Load()
	if currentState == stateRunning || currentState == stateStopping {
		if su.timer != nil {
			stopped := su.timer.Stop()
			s.logger.Trace().Str("task", id).Bool("timerStoppedSuccessfully", stopped).Msg("Attempted to stop timer for removed task")
		}
	}
	s.stateMu.Unlock()

	callbackCtx := context.Background()
	if s.schedulerCtx != nil {
		callbackCtx = s.schedulerCtx
	}
	if su.onCancel != nil {
		s.logger.Trace().Str("task", id).Msg("Calling cancel callback for removed task")
		f := *su.onCancel
		go f(callbackCtx)
	}

	return nil
}
