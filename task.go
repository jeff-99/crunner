package crunner

import (
	"context"
	"fmt"
)

type Task func(ctx context.Context, meta TriggerMetadata) error

func SimpleTask(f func(ctx context.Context) error) Task {
	return func(ctx context.Context, meta TriggerMetadata) error {
		return f(ctx)
	}
}

// PanicError is the error reported for a task run that panicked.
type PanicError struct {
	// Value is the value passed to panic.
	Value any
	// Stack is the stack trace captured at the point of the panic.
	Stack []byte
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("task panicked: %v", p.Value)
}

// Unwrap returns the panic value when it is an error, so errors.Is and errors.As see through the panic.
func (p *PanicError) Unwrap() error {
	err, _ := p.Value.(error)
	return err
}
