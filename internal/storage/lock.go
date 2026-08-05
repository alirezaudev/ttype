package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

type fileLock struct {
	f *os.File
}

func acquireLock(path string) (*fileLock, error) {
	f, err := os.OpenFile(filepath.Join(path, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}
	if err := lockFile(f); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("acquire lock file: %w", err)
	}

	return &fileLock{f: f}, nil
}

func (l *fileLock) release() {
	unlockFile(l.f)
	_ = l.f.Close()
}
