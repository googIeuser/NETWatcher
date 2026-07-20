package wizard

import "testing"

func TestTransitions(t *testing.T) {
	page, action, err := Next(Welcome, true)
	if err != nil || page != Options || action != NoAction {
		t.Fatalf("welcome next: page=%v action=%v err=%v", page, action, err)
	}
	page, action, err = Next(Options, false)
	if err == nil || page != Options || action != NoAction {
		t.Fatalf("invalid options path should stay on page")
	}
	page, action, err = Next(Options, true)
	if err != nil || page != Ready || action != NoAction {
		t.Fatalf("options next: page=%v action=%v err=%v", page, action, err)
	}
	page, action, err = Next(Ready, true)
	if err != nil || page != Installing || action != StartInstall {
		t.Fatalf("ready next: page=%v action=%v err=%v", page, action, err)
	}
	page, action, err = Next(Finished, true)
	if err != nil || page != Finished || action != FinishWizard {
		t.Fatalf("finish next: page=%v action=%v err=%v", page, action, err)
	}
}

func TestBackTransitions(t *testing.T) {
	if got := Back(Options); got != Welcome {
		t.Fatalf("options back = %v", got)
	}
	if got := Back(Ready); got != Options {
		t.Fatalf("ready back = %v", got)
	}
	if got := Back(Installing); got != Installing {
		t.Fatalf("installing back must not move")
	}
}

func TestLayoutsHaveNoOverlaps(t *testing.T) {
	sizes := [][2]int32{{940, 680}, {1024, 720}, {760, 610}, {1280, 800}}
	for _, size := range sizes {
		for page := Welcome; page <= Finished; page++ {
			layout := ComputeLayout(size[0], size[1], page)
			if err := ValidateLayout(layout, page); err != nil {
				t.Fatalf("size=%v page=%v: %v", size, page, err)
			}
		}
	}
}

func TestPageControlsAreIsolated(t *testing.T) {
	for page := Welcome; page <= Finished; page++ {
		visible := VisibleControls(page)
		for other := Welcome; other <= Finished; other++ {
			if other == page {
				continue
			}
			for id := range VisibleControls(other) {
				if id == CtrlBack || id == CtrlNext || id == CtrlCancel {
					continue
				}
				if visible[id] {
					t.Fatalf("page %v unexpectedly contains page %v control %d", page, other, id)
				}
			}
		}
	}
}
