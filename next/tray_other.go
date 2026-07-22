//go:build !windows

package main

func startTray(*App) {}
func stopTray()      {}
func syncTrayMenu()  {}

func showTrayNotification(string, string, string) bool { return false }
