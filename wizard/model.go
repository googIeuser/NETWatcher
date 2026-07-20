package wizard

import "fmt"

type Page int

const (
	Welcome Page = iota
	Options
	Ready
	Installing
	Finished
)

const (
	CtrlLanguageCombo = 6001
	CtrlPathEdit      = 6002
	CtrlBrowse        = 6003
	CtrlDesktop       = 6004
	CtrlStartMenu     = 6005
	CtrlLaunch        = 6006
	CtrlBack          = 6007
	CtrlNext          = 6008
	CtrlCancel        = 6009
	CtrlSummary       = 6012
	CtrlProgress      = 6013
	CtrlDetails       = 6014
	CtrlAutoStart     = 6019
)

type Rect struct {
	X int32
	Y int32
	W int32
	H int32
}

func (r Rect) Right() int32  { return r.X + r.W }
func (r Rect) Bottom() int32 { return r.Y + r.H }
func (r Rect) Valid() bool   { return r.W > 0 && r.H > 0 }

func (r Rect) Overlaps(other Rect) bool {
	return r.X < other.Right() && r.Right() > other.X && r.Y < other.Bottom() && r.Bottom() > other.Y
}

type Layout struct {
	Width       int32
	Height      int32
	Header      Rect
	Title       Rect
	Subtitle    Rect
	Card        Rect
	Footer      Rect
	Controls    map[int]Rect
	TextRegions map[string]Rect
}

func ComputeLayout(width, height int32, page Page) Layout {
	if width < 940 {
		width = 940
	}
	if height < 680 {
		height = 680
	}
	footerH := int32(76)
	footerY := height - footerH
	layout := Layout{
		Width:       width,
		Height:      height,
		Header:      Rect{0, 0, width, 140},
		Title:       Rect{54, 158, width - 108, 50},
		Subtitle:    Rect{55, 210, width - 110, 42},
		Card:        Rect{54, 274, width - 108, footerY - 294},
		Footer:      Rect{0, footerY, width, footerH},
		Controls:    map[int]Rect{},
		TextRegions: map[string]Rect{},
	}

	buttonY := footerY + 18
	layout.Controls[CtrlBack] = Rect{width - 414, buttonY, 122, 40}
	layout.Controls[CtrlNext] = Rect{width - 282, buttonY, 122, 40}
	layout.Controls[CtrlCancel] = Rect{width - 150, buttonY, 122, 40}

	switch page {
	case Welcome:
		layout.Card = Rect{54, 276, width - 108, footerY - 310}
		layout.TextRegions["welcomeBody"] = Rect{90, 304, width - 180, 86}
		layout.TextRegions["languageLabel"] = Rect{90, 398, 300, 28}
		layout.Controls[CtrlLanguageCombo] = Rect{90, 438, 300, 100}
	case Options:
		layout.TextRegions["pathLabel"] = Rect{112, 302, 76, 32}
		layout.Controls[CtrlPathEdit] = Rect{198, 292, width - 410, 40}
		layout.Controls[CtrlBrowse] = Rect{width - 194, 290, 132, 42}
		layout.TextRegions["pathNote"] = Rect{92, 348, width - 184, 52}
		layout.Controls[CtrlDesktop] = Rect{88, 414, width - 176, 40}
		layout.Controls[CtrlStartMenu] = Rect{88, 466, width - 176, 40}
		layout.Controls[CtrlAutoStart] = Rect{88, 518, width - 176, 40}
	case Ready:
		layout.Controls[CtrlSummary] = Rect{74, 294, width - 148, footerY - 324}
	case Installing:
		layout.Controls[CtrlProgress] = Rect{74, 302, width - 148, 24}
		layout.TextRegions["detailsLabel"] = Rect{74, 346, width - 148, 26}
		layout.Controls[CtrlDetails] = Rect{74, 378, width - 148, footerY - 404}
	case Finished:
		layout.TextRegions["finishBody"] = Rect{86, 306, width - 172, 150}
		layout.Controls[CtrlLaunch] = Rect{84, 486, width - 168, 42}
	}
	return layout
}

func VisibleControls(page Page) map[int]bool {
	visible := map[int]bool{
		CtrlBack:   true,
		CtrlNext:   true,
		CtrlCancel: true,
	}
	switch page {
	case Welcome:
		visible[CtrlLanguageCombo] = true
	case Options:
		visible[CtrlPathEdit] = true
		visible[CtrlBrowse] = true
		visible[CtrlDesktop] = true
		visible[CtrlStartMenu] = true
		visible[CtrlAutoStart] = true
	case Ready:
		visible[CtrlSummary] = true
	case Installing:
		visible[CtrlProgress] = true
		visible[CtrlDetails] = true
	case Finished:
		visible[CtrlLaunch] = true
	}
	return visible
}

type Action int

const (
	NoAction Action = iota
	StartInstall
	FinishWizard
)

func Next(page Page, validPath bool) (Page, Action, error) {
	switch page {
	case Welcome:
		return Options, NoAction, nil
	case Options:
		if !validPath {
			return Options, NoAction, fmt.Errorf("invalid install path")
		}
		return Ready, NoAction, nil
	case Ready:
		return Installing, StartInstall, nil
	case Installing:
		return Installing, NoAction, nil
	case Finished:
		return Finished, FinishWizard, nil
	default:
		return page, NoAction, fmt.Errorf("unknown page: %d", page)
	}
}

func Back(page Page) Page {
	switch page {
	case Options:
		return Welcome
	case Ready:
		return Options
	default:
		return page
	}
}

func ValidateLayout(layout Layout, page Page) error {
	bounds := Rect{0, 0, layout.Width, layout.Height}
	for id, r := range layout.Controls {
		if !r.Valid() {
			return fmt.Errorf("control %d has invalid rectangle: %+v", id, r)
		}
		if r.X < 0 || r.Y < 0 || r.Right() > bounds.Right() || r.Bottom() > bounds.Bottom() {
			return fmt.Errorf("control %d is outside window: %+v", id, r)
		}
	}
	for name, r := range layout.TextRegions {
		if !r.Valid() {
			return fmt.Errorf("text region %s has invalid rectangle: %+v", name, r)
		}
		if r.X < 0 || r.Y < 0 || r.Right() > bounds.Right() || r.Bottom() > bounds.Bottom() {
			return fmt.Errorf("text region %s is outside window: %+v", name, r)
		}
	}

	visible := VisibleControls(page)
	ids := make([]int, 0, len(visible))
	for id := range visible {
		if _, ok := layout.Controls[id]; ok {
			ids = append(ids, id)
		}
	}
	for a := 0; a < len(ids); a++ {
		for b := a + 1; b < len(ids); b++ {
			ra, rb := layout.Controls[ids[a]], layout.Controls[ids[b]]
			if ra.Overlaps(rb) {
				return fmt.Errorf("visible controls %d and %d overlap: %+v / %+v", ids[a], ids[b], ra, rb)
			}
		}
	}
	for name, textRect := range layout.TextRegions {
		for _, id := range ids {
			controlRect := layout.Controls[id]
			if textRect.Overlaps(controlRect) {
				return fmt.Errorf("text region %s overlaps control %d: %+v / %+v", name, id, textRect, controlRect)
			}
		}
	}
	return nil
}
