//go:build !windows

package monitor

func detectDefaultGateway() string { return "" }
