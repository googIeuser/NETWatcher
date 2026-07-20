//go:build windows

package main

import "syscall"

func mainButtonWindowStyle() uint32 {
	return WS_CHILD | WS_VISIBLE | WS_TABSTOP | BS_OWNERDRAW
}

func isMainButtonID(id int) bool {
	switch id {
	case ctrlStart, ctrlStop, ctrlReport, ctrlLogs, ctrlSettings, ctrlStats,
		ctrlExport, ctrlHistory, ctrlEvidence, ctrlAccess, ctrlAdd, ctrlRemove, ctrlTargets:
		return true
	default:
		return false
	}
}

func isPrimaryMainButton(id int) bool {
	return id == ctrlStart || id == ctrlAccess
}

func isAccentMainButton(id int) bool {
	return id == ctrlAdd || id == ctrlReport || id == ctrlStats || id == ctrlEvidence
}

func isDangerMainButton(id int) bool {
	return id == ctrlStop || id == ctrlRemove
}

func (a *App) drawMainButton(dis *DRAWITEMSTRUCT) {
	if a == nil || dis == nil {
		return
	}

	hdc := dis.HDC
	rc := dis.RcItem
	id := int(dis.CtlID)
	text := getText(dis.HwndItem)
	dark := isDarkTheme(a.config.Theme)
	pressed := dis.ItemState&ODS_SELECTED != 0
	disabled := dis.ItemState&ODS_DISABLED != 0
	focused := dis.ItemState&ODS_FOCUS != 0

	fill := rgb(245, 247, 250)
	border := rgb(190, 198, 209)
	textColor := rgb(32, 38, 48)
	focusColor := rgb(73, 138, 255)

	if dark {
		fill = rgb(35, 40, 49)
		border = rgb(67, 75, 88)
		textColor = rgb(235, 239, 245)
		focusColor = rgb(105, 166, 255)
	}

	switch {
	case isPrimaryMainButton(id):
		fill = rgb(35, 105, 230)
		border = rgb(50, 127, 255)
		textColor = rgb(255, 255, 255)
	case isDangerMainButton(id):
		if dark {
			fill = rgb(49, 34, 39)
			border = rgb(190, 74, 88)
			textColor = rgb(255, 175, 183)
		} else {
			fill = rgb(255, 247, 248)
			border = rgb(213, 74, 91)
			textColor = rgb(173, 38, 55)
		}
	case isAccentMainButton(id):
		if dark {
			fill = rgb(30, 45, 65)
			border = rgb(62, 126, 218)
			textColor = rgb(176, 210, 255)
		} else {
			fill = rgb(244, 249, 255)
			border = rgb(81, 139, 222)
			textColor = rgb(31, 93, 180)
		}
	}

	if pressed {
		switch {
		case isPrimaryMainButton(id):
			fill = rgb(25, 82, 183)
			border = rgb(39, 105, 218)
		case isDangerMainButton(id):
			if dark {
				fill = rgb(71, 38, 45)
			} else {
				fill = rgb(249, 226, 230)
			}
		case isAccentMainButton(id):
			if dark {
				fill = rgb(37, 58, 83)
			} else {
				fill = rgb(228, 240, 255)
			}
		default:
			if dark {
				fill = rgb(47, 53, 64)
			} else {
				fill = rgb(229, 234, 241)
			}
		}
	}

	if disabled {
		if dark {
			fill = rgb(31, 35, 42)
			border = rgb(52, 58, 68)
			textColor = rgb(112, 120, 133)
		} else {
			fill = rgb(239, 242, 246)
			border = rgb(214, 219, 226)
			textColor = rgb(150, 157, 168)
		}
	}

	drawRoundedBox(hdc, RECT{1, 1, rc.Right - 1, rc.Bottom - 1}, 10, fill, border, 1)
	font := a.buttonFont
	if font == 0 {
		stock, _, _ := procGetStockObject.Call(DEFAULT_GUI_FONT)
		font = syscall.Handle(stock)
	}
	drawInstallerTextRect(hdc, rc, text, textColor, font, DT_CENTER|DT_VCENTER|DT_SINGLELINE)

	if focused && !disabled {
		inner := RECT{4, 4, rc.Right - 4, rc.Bottom - 4}
		drawRoundedBox(hdc, inner, 8, fill, focusColor, 1)
		drawInstallerTextRect(hdc, rc, text, textColor, font, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	}
}
