package main

import "strings"

func findCustomTargetIndex(targets []string, host string) int {
	host = strings.TrimSpace(host)
	if host == "" {
		return -1
	}
	for index, target := range targets {
		if strings.EqualFold(strings.TrimSpace(target), host) {
			return index
		}
	}
	return -1
}

func removeCustomTargetValue(targets []string, host string) ([]string, bool) {
	index := findCustomTargetIndex(targets, host)
	if index < 0 {
		return targets, false
	}
	result := make([]string, 0, len(targets)-1)
	result = append(result, targets[:index]...)
	result = append(result, targets[index+1:]...)
	return result, true
}
