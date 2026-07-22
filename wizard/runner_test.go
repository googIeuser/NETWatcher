package wizard

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestRunPlanCompletesAndEmitsDone(t *testing.T) {
	var ran []string
	var events []Event
	steps := []Step{
		{Name: "prepare", Progress: 10, Fatal: true, Run: func(context.Context) error { ran = append(ran, "prepare"); return nil }},
		{Name: "copy", Progress: 40, Fatal: true, Run: func(context.Context) error { ran = append(ran, "copy"); return nil }},
		{Name: "shortcuts", Progress: 70, Fatal: false, Run: func(context.Context) error { ran = append(ran, "shortcuts"); return errors.New("optional failure") }},
		{Name: "registry", Progress: 90, Fatal: true, Run: func(context.Context) error { ran = append(ran, "registry"); return nil }},
	}
	err := RunPlan(context.Background(), steps, func(event Event) { events = append(events, event) })
	if err != nil {
		t.Fatalf("RunPlan returned error: %v", err)
	}
	if !reflect.DeepEqual(ran, []string{"prepare", "copy", "shortcuts", "registry"}) {
		t.Fatalf("unexpected run order: %v", ran)
	}
	if len(events) == 0 || !events[len(events)-1].Done || events[len(events)-1].Progress != 100 {
		t.Fatalf("missing done event: %#v", events)
	}
}

func TestRunPlanStopsOnFatalError(t *testing.T) {
	ranLast := false
	err := RunPlan(context.Background(), []Step{
		{Name: "copy", Progress: 40, Fatal: true, Run: func(context.Context) error { return errors.New("disk full") }},
		{Name: "registry", Progress: 90, Fatal: true, Run: func(context.Context) error { ranLast = true; return nil }},
	}, nil)
	if err == nil {
		t.Fatal("expected fatal error")
	}
	if ranLast {
		t.Fatal("step after fatal error was executed")
	}
}

func TestRunPlanHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ran := false
	err := RunPlan(ctx, []Step{{Name: "prepare", Fatal: true, Run: func(context.Context) error { ran = true; return nil }}}, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected canceled, got %v", err)
	}
	if ran {
		t.Fatal("step ran after cancellation")
	}
}
