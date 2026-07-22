//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

type settingsSurfaceLayout struct {
	header     RECT
	monitoring RECT
	startup    RECT
	updates    RECT
	footerInfo RECT
}

func settingsSurfaceGeometry(width, height int32) settingsSurfaceLayout {
	if width < 880 {
		width = 880
	}
	if height < 840 {
		height = 840
	}
	updatesBottom := height - 70
	if updatesBottom < 770 {
		updatesBottom = 770
	}
	return settingsSurfaceLayout{
		header:     RECT{0, 0, width, 96},
		monitoring: RECT{24, 108, width - 24, 354},
		startup:    RECT{24, 366, width - 24, 584},
		updates:    RECT{24, 596, width - 24, updatesBottom},
		footerInfo: RECT{32, height - 54, width - 350, height - 16},
	}
}

func (s *SettingsWindow) drawSettingsSurface(hdc syscall.Handle, client RECT) {
	if s == nil {
		return
	}
	dark := isDarkTheme(s.parent.config.Theme)
	background := rgb(246, 248, 252)
	cardFill := rgb(250, 250, 252)
	cardBorder := rgb(210, 218, 229)
	text := rgb(31, 38, 49)
	muted := rgb(92, 103, 119)
	if dark {
		background = rgb(24, 28, 35)
		cardFill = rgb(35, 39, 47)
		cardBorder = rgb(57, 66, 79)
		text = rgb(239, 243, 249)
		muted = rgb(172, 182, 196)
	}

	brush, _, _ := procCreateSolidBrush.Call(background)
	procFillRect.Call(uintptr(hdc), uintptr(unsafe.Pointer(&client)), brush)
	procDeleteObject.Call(brush)

	layout := settingsSurfaceGeometry(client.Right, client.Bottom)
	drawRoundedBox(hdc, layout.header, 0, rgb(24, 112, 238), rgb(24, 112, 238), 1)
	drawInstallerTextRect(hdc, RECT{32, 18, layout.header.Right - 32, 54}, "NetWatcher Settings", rgb(255, 255, 255), s.titleFont, DT_LEFT|DT_VCENTER|DT_SINGLELINE)
	drawInstallerTextRect(hdc, RECT{32, 52, layout.header.Right - 32, 86}, "Configure monitoring, startup behavior, notifications and local log storage.", rgb(224, 239, 255), s.subtitleFont, DT_LEFT|DT_VCENTER|DT_WORDBREAK)

	cards := []struct {
		rect  RECT
		title string
	}{
		{layout.monitoring, "Monitoring"},
		{layout.startup, "Startup and background behavior"},
		{layout.updates, "Updates, notifications and storage"},
	}
	for _, card := range cards {
		drawRoundedBox(hdc, card.rect, 14, cardFill, cardBorder, 1)
		drawInstallerTextRect(hdc, RECT{card.rect.Left + 20, card.rect.Top + 8, card.rect.Right - 20, card.rect.Top + 38}, card.title, text, s.sectionFont, DT_LEFT|DT_VCENTER|DT_SINGLELINE)
	}
	drawInstallerTextRect(hdc, layout.footerInfo, "All settings are stored locally on this computer.", muted, s.smallFont, DT_LEFT|DT_VCENTER|DT_WORDBREAK)
}

func (s *SettingsWindow) drawSettingsButton(dis *DRAWITEMSTRUCT) {
	if s == nil || dis == nil {
		return
	}
	id := int(dis.CtlID)
	dark := isDarkTheme(s.parent.config.Theme)
	pressed := dis.ItemState&ODS_SELECTED != 0
	disabled := dis.ItemState&ODS_DISABLED != 0
	focused := dis.ItemState&ODS_FOCUS != 0

	fill := rgb(242, 245, 249)
	border := rgb(180, 190, 203)
	textColor := rgb(39, 47, 59)
	if dark {
		fill = rgb(36, 42, 51)
		border = rgb(74, 84, 98)
		textColor = rgb(237, 241, 247)
	}
	if id == setSave {
		fill = rgb(28, 111, 235)
		border = rgb(63, 139, 255)
		textColor = rgb(255, 255, 255)
	}
	if pressed {
		if id == setSave {
			fill = rgb(22, 87, 186)
		} else if dark {
			fill = rgb(48, 55, 66)
		} else {
			fill = rgb(226, 232, 240)
		}
	}
	if disabled {
		fill = rgb(222, 226, 232)
		border = rgb(196, 202, 211)
		textColor = rgb(135, 143, 154)
		if dark {
			fill = rgb(31, 36, 43)
			border = rgb(52, 59, 69)
			textColor = rgb(112, 121, 134)
		}
	}

	rc := dis.RcItem
	drawRoundedBox(dis.HDC, RECT{1, 1, rc.Right - 1, rc.Bottom - 1}, 10, fill, border, 1)
	drawInstallerTextRect(dis.HDC, rc, getText(dis.HwndItem), textColor, s.buttonFont, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	if focused && !disabled {
		drawRoundedBox(dis.HDC, RECT{4, 4, rc.Right - 4, rc.Bottom - 4}, 8, fill, rgb(116, 177, 255), 1)
		drawInstallerTextRect(dis.HDC, rc, getText(dis.HwndItem), textColor, s.buttonFont, DT_CENTER|DT_VCENTER|DT_SINGLELINE)
	}
}

func (s *SettingsWindow) releaseSettingsFonts() {
	if s == nil {
		return
	}
	for _, font := range []syscall.Handle{s.titleFont, s.subtitleFont, s.sectionFont, s.bodyFont, s.smallFont, s.buttonFont} {
		if font != 0 {
			procDeleteObject.Call(uintptr(font))
		}
	}
	s.titleFont, s.subtitleFont, s.sectionFont, s.bodyFont, s.smallFont, s.buttonFont = 0, 0, 0, 0, 0, 0
}
