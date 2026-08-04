package tui

import (
	"bytes"
	"os"
)

type capsLockMonitor struct {
	paths []string
}

func newCapsLockMonitor() *capsLockMonitor {
	return &capsLockMonitor{paths: capsLockLEDPaths()}
}

func (m *capsLockMonitor) on() bool {
	if m == nil {
		return false
	}
	for _, p := range m.paths {
		b, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if v := bytes.TrimSpace(b); len(v) > 0 && !bytes.Equal(v, []byte("0")) {
			return true
		}
	}
	return false
}
