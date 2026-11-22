package crunner

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"runtime/debug"
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

// WithTaskTimeout sets the default timeout of every run (default 30 days). A task's own WithTimeout
// overrides it. A value <= 0 disables the timeout: runs are then only cancelled when the scheduler stops.
func WithTaskTimeout(d time.Duration) SchedulerOption {
	return func(o *internalScheduler) {
		o.taskTimeout = d
	}
}

// WithScheduleTimeout bounds each Schedule.Next call (default 1 minute). A value <= 0 disables the
// timeout: the call's context is then only cancelled when the scheduler stops.
func WithScheduleTimeout(d time.Duration) SchedulerOption {
	return func(o *internalScheduler) {
		o.scheduleTimeout = d
	}
}

// WithQueueSize sets the buffer size of the scheduler's internal queues (default 100). Values below 1
// are clamped to 1.
func WithQueueSize(n int) SchedulerOption {
	return func(o *internalScheduler) {
		o.queueSize = n
	}
}

type (
	SuccessCallback func(ctx context.Context)
	ErrorCallback   func(ctx context.Context, err error)
	// SkipCallback is called when a run is skipped for drift. meta describes the skipped run: its trigger
	// window, its expected trigger time and, as ActualTriggerTime, the moment the skip was decided.
	SkipCallback func(ctx context.Context, meta TriggerMetadata)
	// RemovedCallback is called when a task is removed or replaced and will therefore never run again.
	RemovedCallback func(ctx context.Context)
)

type SchedulingOptions struct {
	MaxDrift  time.Duration
	NotAfter  *time.Time
	NotBefore *time.Time
	Backfill  bool

	ID        string
	OnSuccess *SuccessCallback
	OnError   *ErrorCallback
	OnSkip    *SkipCallback
	OnRemoved *RemovedCallback

	// Timeout overrides the scheduler's task timeout for this task; nil inherits it.
	Timeout *time.Duration
}

type SchedulingOption func(*SchedulingOptions)

func WithDefaultOptions() SchedulingOptions {
	return SchedulingOptions{
		MaxDrift: 30 * time.Second,
		Backfill: false,
	}
}

// WithMaxDrift sets how late a run may start relative to its expected trigger time, the end of its
// window (default 30s). A later run is skipped: it is not executed, the skip callback
// (WithSkipCallback) is called once, and its window is merged into the next run's window, which fires at
// the schedule's next point after now. Drift includes time spent waiting for the task's own previous
// run, so a task whose runs overrun its interval by more than MaxDrift is skipped now and then; raise
// MaxDrift for tasks that may overrun. For long intervals (e.g. a daily cron) a skipped run's work waits
// up to one interval.
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

// WithTimeout sets this task's run timeout, overriding the scheduler's WithTaskTimeout.
// A value <= 0 disables the timeout for this task.
func WithTimeout(d time.Duration) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.Timeout = &d
	}
}

// WithSkipCallback sets a callback called once when a run is skipped because it started more than
// MaxDrift late. The task is not gone: the skipped window is merged into its next run. The callback runs
// asynchronously, never blocks scheduling, and Stop waits for it. It is never called for backfilling
// tasks, which are not skipped.
func WithSkipCallback(f SkipCallback) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.OnSkip = &f
	}
}

// WithRemovedCallback sets a callback called when the task is removed (RemoveTask) or replaced by a new
// registration with the same ID, that is, when it will never run again. It is not called for drift
// skips; see WithSkipCallback.
func WithRemovedCallback(f RemovedCallback) SchedulingOption {
	return func(o *SchedulingOptions) {
		o.OnRemoved = &f
	}
}

type Scheduler interface {
	RegisterTask(task Task, schedule Schedule, options ...SchedulingOption) error
	RemoveTask(string) error
	ScheduledTasks() []Task
	Start(ctx context.Context) error
	Stop() error
}

// TriggerMetadata describes a run. Its trigger window [TriggerWindowStart, TriggerWindowEnd) is
// contiguous with the task's previous window and lies on the schedule's grid. A window normally spans
// one step of the schedule, but after a run was cancelled for drift it spans several steps, so do not
// assume TriggerWindowEnd - TriggerWindowStart equals the interval.
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
	doneCh    chan runResult

	// closed by mainLoop on exit; Stop waits on it before waiting for running tasks
	loopDone chan struct{}

	summaryInterval time.Duration
	taskTimeout     time.Duration
	scheduleTimeout time.Duration
	queueSize       int

	state atomic.Int32
}

// runResult is sent by a task goroutine to the main loop, the sole owner of
// per-task mutable state, when a run has finished.
type runResult struct {
	unit  *schedulingUnit
	start time.Time
	err   error
}

type schedulingUnit struct {
	task     Task
	schedule Schedule

	meta    *TriggerMetadata
	options SchedulingOptions

	// coveredUntil is the end of the last executed window; every new window starts here.
	// It is zero until the unit is first scheduled while the scheduler runs, which sets the window origin.
	coveredUntil time.Time
	// hasRun reports whether any window has been executed, i.e. whether coveredUntil is a previous window end.
	hasRun bool

	timer *time.Timer

	backfilling bool

	onError   *ErrorCallback
	onSuccess *SuccessCallback
	onSkip    *SkipCallback
	onRemoved *RemovedCallback

	// snapshot of counts at last summary report (read/written only from mainLoop goroutine)
	lastReportedCount        int
	lastReportedSuccessCount int
	lastReportedErrorCount   int
}

func New(opts ...SchedulerOption) Scheduler {
	scheduler := &internalScheduler{
		wg:              &sync.WaitGroup{},
		tasks:           sync.Map{},
		logger:          zerolog.Nop(), // Initialize with Nop logger by default
		summaryInterval: 30 * time.Second,
		taskTimeout:     30 * 24 * time.Hour,
		scheduleTimeout: time.Minute,
		queueSize:       100,
	}
	scheduler.state.Store(stateIdle) // Initialize state

	for _, opt := range opts {
		opt(scheduler)
	}

	// Channels are created after the options so that their buffer size can be configured.
	scheduler.queueSize = max(scheduler.queueSize, 1)
	scheduler.newTaskCh = make(chan *schedulingUnit, scheduler.queueSize)
	scheduler.tasksIn = make(chan *schedulingUnit, scheduler.queueSize)
	scheduler.doneCh = make(chan runResult, scheduler.queueSize)

	return scheduler
}

// ID returns the task's ID, fixed at registration.
func (su *schedulingUnit) ID() string {
	return su.options.ID
}

// RegisterTask stores a task under its ID, replacing any task already registered with that ID.
// It returns an error, wrapping ErrInvalidTask, ErrInvalidSchedule and/or ErrInvalidOption, and registers
// nothing (an existing task with the same ID is left untouched) when: the task or schedule is nil, an
// interval is not positive, MaxDrift is negative, NotAfter is not after NotBefore, backfill is enabled
// without NotBefore, or a callback is nil.
func (s *internalScheduler) RegisterTask(task Task, schedule Schedule, options ...SchedulingOption) error {
	opts := WithDefaultOptions()
	for _, opt := range options {
		opt(&opts)
	}

	// The ID is fixed here, once: an empty ID means "generate one".
	if opts.ID == "" {
		opts.ID = "task-" + rand.Text()
	}

	// Validate before anything has side effects, so a rejected registration leaves an existing task untouched.
	if err := validateRegistration(task, schedule, opts); err != nil {
		return err
	}

	newSchedulingUnit := &schedulingUnit{
		task:     task,
		schedule: schedule,
		options:  opts,
		meta:     &TriggerMetadata{},

		onSuccess: opts.OnSuccess,
		onError:   opts.OnError,
		onSkip:    opts.OnSkip,
		onRemoved: opts.OnRemoved,
	}

	// Swap the new unit in atomically; exactly the unit it displaced (if any) is cancelled.
	id := newSchedulingUnit.ID()
	if previous, loaded := s.tasks.Swap(id, newSchedulingUnit); loaded {
		s.logger.Trace().Str("task", id).Msg("Overwriting existing task registration")
		// The main loop stops the overwritten unit's timer when it sees the new unit;
		// a stale timer that fires first is dropped by the main loop's identity check.
		if existingUnit, ok := previous.(*schedulingUnit); ok && existingUnit.onRemoved != nil {
			callbackCtx := context.Background()
			if s.schedulerCtx != nil {
				callbackCtx = s.schedulerCtx
			}
			s.logger.Trace().Str("task", id).Msg("Calling removed callback for overwritten task")
			f := *existingUnit.onRemoved
			go f(callbackCtx) // Run in a goroutine to avoid blocking registration
		}
	}

	// Send the new unit to the main loop for scheduling its *first* run (or rescheduling if Start was already called).
	// Only send if the scheduler is actually running.
	if s.state.Load() == stateRunning {
		select {
		case <-s.schedulerCtx.Done():
			s.logger.Warn().Str("task", id).Msg("Scheduler context done, cannot schedule newly registered task")
		case s.newTaskCh <- newSchedulingUnit:
			s.logger.Trace().Str("task", id).Msg("New/overwritten task sent for scheduling")
		}
	} else {
		s.logger.Trace().Str("task", id).Msg("Scheduler not running, task registered but not scheduled yet")
	}

	return nil
}

func (s *internalScheduler) scheduleTask(su *schedulingUnit) {
	s.scheduleWindow(su, nil)
}

// scheduleWindow computes the task's next window and sets its timer. Every window starts at
// coveredUntil. Normally the window ends at the schedule's next point after that start. When
// skippedEnd is set, the previous run was skipped for drift: the new window then also covers
// it and ends at the first point of the schedule after now.
func (s *internalScheduler) scheduleWindow(su *schedulingUnit, skippedEnd *time.Time) {
	if s.state.Load() != stateRunning {
		s.logger.Trace().Str("task", su.ID()).Msg("Scheduler not running, skipping task scheduling")
		return
	}

	s.logger.Trace().Str("task", su.ID()).Msg("Scheduling task")
	ctx, cancel := withOptionalTimeout(s.schedulerCtx, s.scheduleTimeout)
	defer cancel()

	isBackfill := su.options.Backfill && su.options.NotBefore != nil
	now := time.Now()

	if su.coveredUntil.IsZero() {
		// Window origin: NotBefore when backfilling, otherwise the moment the task is first scheduled.
		if isBackfill {
			su.coveredUntil = *su.options.NotBefore
		} else {
			su.coveredUntil = now
		}
	}
	triggerWindowStart := su.coveredUntil

	var triggerWindowEnd time.Time
	var err error
	if skippedEnd == nil {
		var ref *time.Time
		if su.hasRun {
			ref = &triggerWindowStart
		}
		triggerWindowEnd, err = su.schedule.Next(ctx, triggerWindowStart, ref)
	} else {
		triggerWindowEnd = *skippedEnd
		for err == nil && !triggerWindowEnd.After(now) {
			var next time.Time
			next, err = su.schedule.Next(ctx, triggerWindowStart, &triggerWindowEnd)
			if err == nil && !next.After(triggerWindowEnd) {
				err = fmt.Errorf("schedule did not advance past %s", triggerWindowEnd)
			}
			triggerWindowEnd = next
		}
	}
	if err != nil {
		s.logger.Error().Err(err).Str("task", su.ID()).Time("windowStart", triggerWindowStart).Msg("Failed to get next run time from schedule")
		return
	}

	var expectedTriggerTime time.Time
	if isBackfill && !triggerWindowEnd.After(now) {
		expectedTriggerTime = now
		su.backfilling = true
		s.logger.Trace().Str("task", su.ID()).Time("windowStart", triggerWindowStart).Time("windowEnd", triggerWindowEnd).Msg("Task is backfilling")
	} else {
		// Drift is measured from the window end, so a run delayed by its own previous run counts as late.
		expectedTriggerTime = triggerWindowEnd
		su.backfilling = false
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
		s.logger.Warn().Str("task", su.ID()).Time("expectedTriggerTime", su.meta.ExpectedTriggerTime).Time("actualTime", now).Dur("drift", drift).Dur("maxDrift", su.options.MaxDrift).Time("windowStart", su.meta.TriggerWindowStart).Time("windowEnd", su.meta.TriggerWindowEnd).Msg("Task run skipped due to excessive drift; its window is merged into the next run")

		// Copy the skipped run's metadata before scheduleWindow overwrites it with the merged window.
		skippedMeta := *su.meta
		skippedMeta.ActualTriggerTime = now
		if su.onSkip != nil {
			f := *su.onSkip
			s.wg.Add(1)
			go func() {
				defer s.wg.Done()
				defer cancel()
				f(ctx, skippedMeta)
			}()
		} else {
			cancel()
		}
		skippedEnd := su.meta.TriggerWindowEnd
		s.scheduleWindow(su, &skippedEnd)
		return
	}

	su.meta.ActualTriggerTime = now
	metaCopy := *su.meta

	s.logger.Trace().Str("task", su.ID()).Msg("Task running")

	s.wg.Add(1)
	go func(taskUnit *schedulingUnit) {
		defer s.wg.Done()
		defer cancel()
		var err error
		func() {
			timeout := s.taskTimeout
			if taskUnit.options.Timeout != nil {
				timeout = *taskUnit.options.Timeout
			}
			taskCtx, taskCancel := withOptionalTimeout(ctx, timeout)
			defer taskCancel()
			// Recover here, not around the whole goroutine, so the result is still
			// recorded and the task rescheduled. runtime.Goexit is not a panic and is not handled.
			defer func() {
				if r := recover(); r != nil {
					err = &PanicError{Value: r, Stack: debug.Stack()}
				}
			}()
			err = taskUnit.task(taskCtx, metaCopy)
		}()

		finished := time.Now()

		if err != nil {
			logEvent := s.logger.Error().Err(err).Str("task", taskUnit.ID())
			var panicErr *PanicError
			if errors.As(err, &panicErr) {
				logEvent = logEvent.Bytes("stack", panicErr.Stack)
			}
			logEvent.Msg("Task execution failed")
			if taskUnit.onError != nil {
				f := *taskUnit.onError
				f(ctx, err)
			}
		} else {
			s.logger.Debug().Str("task", taskUnit.ID()).Str("finished_time", finished.Format(time.RFC3339)).Str("start_time", now.Format(time.RFC3339)).Str("duration", finished.Sub(now).String()).Msgf("Task finished successfully")
			if taskUnit.onSuccess != nil {
				f := *taskUnit.onSuccess
				f(ctx)
			}
		}

		select {
		case <-s.schedulerCtx.Done():
			s.logger.Trace().Str("task", taskUnit.ID()).Msg("Scheduler context done, not reporting run result.")
		case s.doneCh <- runResult{unit: taskUnit, start: now, err: err}:
			s.logger.Trace().Str("task", taskUnit.ID()).Msg("Run result sent to main loop.")
		}
	}(su)
}

// handleRunResult records a finished run on the main loop and reschedules the task if it is still registered.
func (s *internalScheduler) handleRunResult(r runResult) {
	su := r.unit
	now := r.start

	su.meta.LastRunTime = &now
	su.meta.Count++
	su.coveredUntil = su.meta.TriggerWindowEnd
	su.hasRun = true
	if r.err != nil {
		su.meta.LastErroredRunTime = &now
		su.meta.ErrorCount++
		su.meta.LastError = r.err
	} else {
		su.meta.LastSuccessfulRunTime = &now
		su.meta.SuccessfulCount++
		su.meta.LastError = nil
	}

	if !s.isRegistered(su) {
		s.logger.Trace().Str("task", su.ID()).Msg("Mainloop: Run result for overwritten or removed task, not rescheduling")
		return
	}
	s.scheduleTask(su)
}

func (s *internalScheduler) isRegistered(su *schedulingUnit) bool {
	currentVal, ok := s.tasks.Load(su.ID())
	if !ok {
		return false
	}
	currentUnit, isUnit := currentVal.(*schedulingUnit)
	return isUnit && currentUnit == su
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

	// Units whose timers this loop owns, by task ID. Only the main loop touches
	// unit timers; on exit it stops every one of them.
	owned := make(map[string]*schedulingUnit)
	defer func() {
		for _, u := range owned {
			if u.timer != nil {
				u.timer.Stop()
			}
		}
		close(s.loopDone)
	}()
	release := func(u *schedulingUnit) {
		if u.timer != nil {
			u.timer.Stop()
		}
		if owned[u.ID()] == u {
			delete(owned, u.ID())
		}
	}

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
					if prev := owned[receivedUnit.ID()]; prev != nil && prev != receivedUnit {
						release(prev) // superseded by an overwrite
					}
					owned[receivedUnit.ID()] = receivedUnit
					s.scheduleTask(receivedUnit) // Calculate next run time and set timer
				} else {
					// Task was overwritten before this schedule request was processed.
					s.logger.Trace().Str("task", receivedUnit.ID()).Msg("Mainloop: Received schedule request for overwritten task, ignoring")
				}
			} else {
				// Task was removed before this schedule request was processed.
				s.logger.Trace().Str("task", receivedUnit.ID()).Msg("Mainloop: Received schedule request for removed task, ignoring")
			}

		// Case 3b: A task run finished
		case r := <-s.doneCh:
			s.handleRunResult(r)
			if !s.isRegistered(r.unit) {
				release(r.unit)
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
					release(su)
				}
			} else {
				// The task was removed between timer firing and processing. Ignore.
				s.logger.Trace().Str("task", su.ID()).Msg("Mainloop: Received ready task for removed task, ignoring")
				release(su)
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

	s.logger.Trace().Msg("Starting scheduler")

	context, cancel := context.WithCancel(ctx)
	s.cancelCtx = cancel
	s.schedulerCtx = context

	s.loopDone = make(chan struct{})

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

	// mainLoop stops all timers on exit; once it is done nothing adds to wg any more.
	<-s.loopDone

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
	taskValue, loaded := s.tasks.LoadAndDelete(id)
	if !loaded {
		return fmt.Errorf("task with id %s not found", id)
	}
	su := taskValue.(*schedulingUnit)

	s.logger.Trace().Str("task", id).Msg("Removed task from map")

	// The unit's timer, if still pending, fires once and is dropped by the main loop.

	callbackCtx := context.Background()
	if s.schedulerCtx != nil {
		callbackCtx = s.schedulerCtx
	}
	if su.onRemoved != nil {
		s.logger.Trace().Str("task", id).Msg("Calling removed callback for removed task")
		f := *su.onRemoved
		go f(callbackCtx)
	}

	return nil
}

// withOptionalTimeout derives a context with the given timeout, or without one when timeout <= 0.
func withOptionalTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, timeout)
}
