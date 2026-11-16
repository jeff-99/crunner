package crunner

import "context"

type Task func(ctx context.Context, meta TriggerMetadata) error

func SimpleTask(f func(ctx context.Context) error) Task {
	return func(ctx context.Context, meta TriggerMetadata) error {
		return f(ctx)
	}
}
