//go:build linux

package tui

import "path/filepath"

func capsLockLEDPaths() []string {
	paths, _ := filepath.Glob("/sys/class/leds/*::capslock/brightness")
	return paths
}
