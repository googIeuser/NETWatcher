package main

import (
	"os"
	"strings"
	"testing"
)

func sourceSection(t *testing.T, source, start, end string) string {
	t.Helper()
	startAt := strings.Index(source, start)
	if startAt < 0 {
		t.Fatalf("source section %q was not found", start)
	}
	endAt := strings.Index(source[startAt+len(start):], end)
	if endAt < 0 {
		t.Fatalf("end of source section %q was not found", end)
	}
	return source[startAt : startAt+len(start)+endAt]
}

func TestDashboardTimingControlsAreSettingsOnly(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	controls := sourceSection(t, source, "func (a *App) buildControls()", "func (a *App) applyLanguage()")
	for _, forbidden := range []string{"ctrlTimeout", "staticTimeout", "ctrlInterval", "staticInterval"} {
		if strings.Contains(controls, forbidden) {
			t.Fatalf("the main dashboard must not create %s", forbidden)
		}
	}
	startCommand := sourceSection(t, source, "case ctrlStart:", "case ctrlStop:")
	for _, forbidden := range []string{"controls[ctrlTimeout]", "controls[ctrlInterval]"} {
		if strings.Contains(startCommand, forbidden) {
			t.Fatalf("the Start button must not read %s", forbidden)
		}
	}
	for _, required := range []string{"a.config.TimeoutMS", "a.config.Interval"} {
		if !strings.Contains(startCommand, required) {
			t.Fatalf("the Start button must use the Settings value %s", required)
		}
	}
}
