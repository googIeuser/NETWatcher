//go:build !windows

package platform

import (
	"os/exec"
	"runtime"
)

func SetStartWithWindows(bool) error { return nil }
func OpenPath(path string) error {
	if runtime.GOOS == "darwin" {
		return exec.Command("open", path).Start()
	}
	return exec.Command("xdg-open", path).Start()
}
func OpenFile(path string) error  { return OpenPath(path) }
func OpenURL(path string) error   { return OpenPath(path) }
func Notify(string, string) error { return nil }
func ProcessID() string           { return "" }
