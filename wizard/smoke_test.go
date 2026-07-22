package wizard

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesystemInstallSmoke(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "setup.bin")
	destination := filepath.Join(root, "installed", "NetWatcher.exe")
	if err := os.WriteFile(source, []byte("netwatcher-payload"), 0644); err != nil {
		t.Fatal(err)
	}
	var done bool
	err := RunPlan(context.Background(), []Step{
		{Name: "prepare", Progress: 10, Fatal: true, Run: func(context.Context) error {
			return os.MkdirAll(filepath.Dir(destination), 0755)
		}},
		{Name: "copy", Progress: 50, Fatal: true, Run: func(context.Context) error {
			data, err := os.ReadFile(source)
			if err != nil {
				return err
			}
			return os.WriteFile(destination, data, 0755)
		}},
		{Name: "metadata", Progress: 90, Fatal: true, Run: func(context.Context) error {
			return os.WriteFile(filepath.Join(filepath.Dir(destination), "install.json"), []byte(`{"version":"0.7.0"}`), 0644)
		}},
	}, func(event Event) {
		if event.Done {
			done = true
		}
	})
	if err != nil {
		t.Fatalf("smoke install failed: %v", err)
	}
	if !done {
		t.Fatal("install did not emit completion")
	}
	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "netwatcher-payload" {
		t.Fatalf("unexpected installed payload: %q", got)
	}
}

func TestRepeatedBackNextNavigation(t *testing.T) {
	page := Welcome
	for cycle := 0; cycle < 500; cycle++ {
		next, action, err := Next(page, true)
		if err != nil || action != NoAction || next != Options {
			t.Fatalf("cycle %d welcome->options failed", cycle)
		}
		page = Back(next)
		if page != Welcome {
			t.Fatalf("cycle %d options->welcome failed", cycle)
		}
	}
}
