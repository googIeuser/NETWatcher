package main

// autoStartArgument selects the command-line mode used by the Windows Run key.
// The startup mode creates the application window without showing it and keeps
// the application reachable through its notification-area icon.
func autoStartArgument(startMinimized bool) string {
	if startMinimized {
		return "--startup"
	}
	return "--app"
}

// monitorButtonState returns the enabled state of the Start and Stop buttons.
// Exactly one command is available for each stable monitoring state.
func monitorButtonState(monitoring bool) (startEnabled bool, stopEnabled bool) {
	return !monitoring, monitoring
}

// Windows sends SIZE_MINIMIZED (value 1) in WM_SIZE when the user clicks
// the title-bar minimize button. Keeping this decision in a small pure
// function makes the behavior testable outside Windows.
func shouldMoveToTrayOnSize(sizeType uintptr) bool {
	return sizeType == 1
}
