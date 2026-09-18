package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
)

// installRetry spaces out background installs that failed, so a machine that
// can't download doesn't try on every launch.
const installRetry = time.Hour

var errTriedRecently = errors.New("tried to install less than an hour ago")

// UpdateModeFor picks what to do about new releases: TTYPE_UPDATE wins over the
// saved setting, and anything unset or unknown means auto.
func UpdateModeFor(setting domain.UpdateMode, env string) domain.UpdateMode {
	if mode, err := domain.ParseUpdateMode(env); err == nil {
		return mode
	}
	if mode, err := domain.ParseUpdateMode(string(setting)); err == nil {
		return mode
	}
	return domain.UpdateAuto
}

// installPlan decides whether ttype replaces itself with info.Latest, and if
// not, what to tell the user to run instead.
func installPlan(info domain.VersionInfo, inst install, exe, goos string, mode domain.UpdateMode, writable func(dir string) bool) domain.VersionInfo {
	switch {
	case !inst.owned():
		info.Command = inst.upgrade
	case goos == "windows":
		info.Command = "ttype update"
	case !sameMajor(info.Latest, info.Local):
		info.Command = "ttype update"
	case !writable(filepath.Dir(exe)):
		info.Command = "sudo ttype update"
	default:
		info.CanInstall = true
		info.AutoInstall = mode == domain.UpdateAuto
	}
	return info
}

func sameMajor(a, b string) bool {
	pa, pb := versionParts(a), versionParts(b)
	return pa != nil && pb != nil && pa[0] == pb[0]
}

func dirWritable(dir string) bool {
	f, err := os.CreateTemp(dir, tempPrefix)
	if err != nil {
		return false
	}
	f.Close()
	_ = os.Remove(f.Name())
	return true
}

// InstallUpdate replaces the running binary with the given release.
func InstallUpdate(build Build, latest string, asked bool) error {
	dirs, err := storage.DefaultDirs()
	if err != nil {
		return err
	}
	inst, exe, err := currentInstall(build)
	if err != nil {
		return err
	}
	return installUpdate(build.Version, latest, asked, inst, newUpdater(exe), updateStatePath(dirs.Data), time.Now(), dirWritable)
}

func installUpdate(local, latest string, asked bool, inst install, u updater, statePath string, now time.Time, writable func(string) bool) error {
	info := domain.VersionInfo{Local: local, Latest: latest, UpdateAvailable: newerVersion(latest, local)}
	if !info.UpdateAvailable {
		return nil
	}
	if plan := installPlan(info, inst, u.exe, u.goos, domain.UpdateAuto, writable); !plan.CanInstall {
		return errors.New("ttype can't update itself here; run " + plan.Command)
	}

	state := loadUpdateState(statePath)
	if state.Installed == latest {
		return nil
	}
	if !asked && now.Sub(state.LastInstall) < installRetry && !now.Before(state.LastInstall) {
		return errTriedRecently
	}

	err := u.install(latest)

	// Recorded after the attempt, so a killed download retries next launch.
	state = loadUpdateState(statePath)
	state.LastInstall = now
	if err == nil {
		state.Installed = latest
	}
	if saveErr := saveUpdateState(statePath, state); err == nil {
		err = saveErr
	}
	return err
}

// How long quitting waits for a running install.
const installWaitCap = 30 * time.Second

// pendingInstall lets quitting wait for a background install.
type pendingInstall struct {
	mu      sync.Mutex
	version string
	done    chan struct{}
}

func (p *pendingInstall) run(version string, install func() error) error {
	done := make(chan struct{})
	p.mu.Lock()
	p.version, p.done = version, done
	p.mu.Unlock()
	defer close(done)
	return install()
}

func (p *pendingInstall) wait(out io.Writer, limit time.Duration) {
	p.mu.Lock()
	version, done := p.version, p.done
	p.mu.Unlock()
	if done == nil {
		return
	}
	select {
	case <-done:
		return
	default:
	}
	fmt.Fprintf(out, "Finishing the update to ttype %s...\n", version)
	select {
	case <-done:
	case <-time.After(limit):
		fmt.Fprintln(out, "Update is taking too long; it will try again next time.")
	}
}
