package wizard

import (
	"context"
	"fmt"
)

type Step struct {
	Name     string
	Progress int
	Fatal    bool
	Run      func(context.Context) error
}

type Event struct {
	Name     string
	Progress int
	Err      error
	Done     bool
}

// RunPlan executes installer steps sequentially. It is UI-toolkit independent,
// so the exact completion/error behavior can be verified on Linux before the
// Windows adapter is built.
func RunPlan(ctx context.Context, steps []Step, emit func(Event)) error {
	if emit == nil {
		emit = func(Event) {}
	}
	for index, step := range steps {
		if err := ctx.Err(); err != nil {
			return err
		}
		if step.Run == nil {
			return fmt.Errorf("step %d (%s) has no runner", index, step.Name)
		}
		emit(Event{Name: step.Name, Progress: step.Progress})
		err := step.Run(ctx)
		if err != nil {
			emit(Event{Name: step.Name, Progress: step.Progress, Err: err})
			if step.Fatal {
				return fmt.Errorf("%s: %w", step.Name, err)
			}
		}
	}
	emit(Event{Progress: 100, Done: true})
	return nil
}
