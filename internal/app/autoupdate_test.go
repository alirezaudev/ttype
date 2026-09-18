package app

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestUpdateModeFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setting domain.UpdateMode
		env     string
		want    domain.UpdateMode
	}{
		{"", "", domain.UpdateAuto},
		{domain.UpdateNotify, "", domain.UpdateNotify},
		{domain.UpdateNotify, "off", domain.UpdateOff},
		{domain.UpdateOff, "auto", domain.UpdateAuto},
		{domain.UpdateOff, "sometimes", domain.UpdateOff},
		{"garbage", "", domain.UpdateAuto},
	}
	for _, test := range tests {
		if got := UpdateModeFor(test.setting, test.env); got != test.want {
			t.Errorf("UpdateModeFor(%q, %q) = %q, want %q", test.setting, test.env, got, test.want)
		}
	}
}

func TestInstallPlan(t *testing.T) {
	t.Parallel()

	owned := install{method: installScript}
	homebrew := detectInstall("/opt/homebrew/Cellar/ttype/1.0.0/bin/ttype", Build{Version: "1.0.0", Release: true}, nil, "darwin")
	system := detectInstall("/usr/bin/ttype", Build{Version: "1.0.0", Release: true}, nil, "linux")
	writable := func(string) bool { return true }
	readOnly := func(string) bool { return false }

	tests := []struct {
		name     string
		latest   string
		inst     install
		goos     string
		mode     domain.UpdateMode
		writable func(string) bool
		can      bool
		auto     bool
		command  string
	}{
		{"owned, auto", "1.2.0", owned, "linux", domain.UpdateAuto, writable, true, true, ""},
		{"owned, notify", "1.2.0", owned, "darwin", domain.UpdateNotify, writable, true, false, ""},
		{"homebrew", "1.2.0", homebrew, "darwin", domain.UpdateAuto, writable, false, false, "brew upgrade ttype"},
		{"system package", "1.2.0", system, "linux", domain.UpdateAuto, writable, false, false, ""},
		{"windows", "1.2.0", owned, "windows", domain.UpdateAuto, writable, false, false, "ttype update"},
		{"new major version", "2.0.0", owned, "linux", domain.UpdateAuto, writable, false, false, "ttype update"},
		{"folder not writable", "1.2.0", owned, "linux", domain.UpdateAuto, readOnly, false, false, "sudo ttype update"},
	}
	for _, test := range tests {
		info := domain.VersionInfo{Local: "1.1.0", Latest: test.latest, UpdateAvailable: true}
		got := installPlan(info, test.inst, "/home/me/.local/bin/ttype", test.goos, test.mode, test.writable)
		if got.CanInstall != test.can || got.AutoInstall != test.auto || got.Command != test.command {
			t.Errorf("%s: can=%v auto=%v command=%q; want can=%v auto=%v command=%q",
				test.name, got.CanInstall, got.AutoInstall, got.Command, test.can, test.auto, test.command)
		}
	}
}

func writableDir(string) bool { return true }

func oldBinary(t *testing.T) string {
	t.Helper()
	exe := filepath.Join(t.TempDir(), "ttype")
	if err := os.WriteFile(exe, []byte("old"), 0o755); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return exe
}

func TestInstallUpdateReplacesTheBinaryInTheBackground(t *testing.T) {
	t.Parallel()

	server, _ := releaseServer(t, "1.2.0", []byte("new"), 0)
	exe := oldBinary(t)
	statePath := filepath.Join(t.TempDir(), "update.json")
	now := time.Date(2026, 9, 17, 15, 0, 0, 0, time.UTC)

	var verified string
	u := testUpdater(server.URL, exe)
	u.verify = func(path, version string) error {
		verified = version
		if got, _ := os.ReadFile(path); string(got) != "new" {
			t.Errorf("verified %q, want the new binary", got)
		}
		return nil
	}

	if err := installUpdate("1.1.0", "1.2.0", false, install{method: installScript}, u, statePath, now, writableDir); err != nil {
		t.Fatalf("installUpdate: %v", err)
	}
	if got, _ := os.ReadFile(exe); string(got) != "new" {
		t.Fatalf("binary = %q, want new", got)
	}
	if verified != "1.2.0" {
		t.Errorf("verified version %q, want the new binary run before the swap", verified)
	}
	if state := loadUpdateState(statePath); state.Installed != "1.2.0" || !state.LastInstall.Equal(now) {
		t.Errorf("state = %+v, want 1.2.0 installed at %v", state, now)
	}
	entries, _ := os.ReadDir(filepath.Dir(exe))
	if len(entries) != 1 {
		t.Errorf("%d files next to the binary, want no temp files left", len(entries))
	}
}

func TestInstallUpdateLeavesTheBinaryWhenTheNewOneDoesNotRun(t *testing.T) {
	t.Parallel()

	server, _ := releaseServer(t, "1.2.0", []byte("new"), 0)
	exe := oldBinary(t)
	u := testUpdater(server.URL, exe)
	u.verify = func(string, string) error { return errors.New("exec format error") }

	err := installUpdate("1.1.0", "1.2.0", false, install{method: installScript}, u, filepath.Join(t.TempDir(), "update.json"), time.Now(), writableDir)
	if err == nil {
		t.Fatal("a binary that doesn't run was installed")
	}
	if got, _ := os.ReadFile(exe); string(got) != "old" {
		t.Fatalf("binary = %q, want the old one kept", got)
	}
	if entries, _ := os.ReadDir(filepath.Dir(exe)); len(entries) != 1 {
		t.Errorf("%d files next to the binary, want the temp file gone", len(entries))
	}
}

// Offline, every launch would otherwise try the download again.
func TestInstallUpdateWaitsAnHourAfterAFailure(t *testing.T) {
	t.Parallel()

	server, requests := releaseServer(t, "1.2.0", []byte("new"), 0)
	exe := oldBinary(t)
	statePath := filepath.Join(t.TempDir(), "update.json")
	now := time.Date(2026, 9, 17, 15, 0, 0, 0, time.UTC)
	inst := install{method: installScript}

	// 1.3.0 isn't on the server, so the download fails.
	if err := installUpdate("1.1.0", "1.3.0", false, inst, testUpdater(server.URL, exe), statePath, now, writableDir); err == nil {
		t.Fatal("installing a missing release succeeded")
	}
	before := requests.Load()

	err := installUpdate("1.1.0", "1.3.0", false, inst, testUpdater(server.URL, exe), statePath, now.Add(10*time.Minute), writableDir)
	if !errors.Is(err, errTriedRecently) || requests.Load() != before {
		t.Fatalf("second try: err = %v, %d new requests; want it skipped", err, requests.Load()-before)
	}

	// Pressing u doesn't wait.
	_ = installUpdate("1.1.0", "1.3.0", true, inst, testUpdater(server.URL, exe), statePath, now.Add(10*time.Minute), writableDir)
	if requests.Load() == before {
		t.Error("an install the user asked for was held back")
	}

	if err := installUpdate("1.1.0", "1.2.0", false, inst, testUpdater(server.URL, exe), statePath, now.Add(2*time.Hour), writableDir); err != nil {
		t.Fatalf("an hour later: %v", err)
	}
}

func TestInstallUpdateRefusesWhatItCanNotOwn(t *testing.T) {
	t.Parallel()

	server, requests := releaseServer(t, "2.0.0", []byte("new"), 0)
	exe := oldBinary(t)
	statePath := filepath.Join(t.TempDir(), "update.json")

	homebrew := detectInstall("/opt/homebrew/Cellar/ttype/1.1.0/bin/ttype", Build{Version: "1.1.0", Release: true}, nil, "darwin")
	for name, inst := range map[string]install{"homebrew": homebrew, "new major": {method: installScript}} {
		if err := installUpdate("1.1.0", "2.0.0", true, inst, testUpdater(server.URL, exe), statePath, time.Now(), writableDir); err == nil {
			t.Errorf("%s: installed anyway", name)
		}
	}
	if requests.Load() != 0 {
		t.Errorf("made %d requests for updates it must not install", requests.Load())
	}
}

func TestCheckVersionAnnouncesAnUpdateOnce(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "update.json")
	now := time.Date(2026, 9, 17, 15, 0, 0, 0, time.UTC)
	offline := func() (string, error) { return "", errors.New("offline") }
	if err := saveUpdateState(statePath, updateState{LastCheck: now, Latest: "1.2.0", Installed: "1.2.0"}); err != nil {
		t.Fatalf("saveUpdateState: %v", err)
	}

	// The old binary is still the one running: nothing to announce yet.
	if info := checkVersion("1.1.0", statePath, now, offline); info.JustUpdated {
		t.Fatal("announced before the new version ran")
	}
	if info := checkVersion("1.2.0", statePath, now, offline); !info.JustUpdated || info.UpdateAvailable {
		t.Fatalf("first launch of 1.2.0: %+v, want it announced and nothing newer", info)
	}
	if info := checkVersion("1.2.0", statePath, now, offline); info.JustUpdated {
		t.Fatal("announced twice")
	}
}

func TestRunsAsVersion(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell script as the binary")
	}

	dir := t.TempDir()
	script := func(name, body string) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
		return path
	}

	good := script("good", "#!/bin/sh\necho 'ttype version 1.2.0'\n")
	if err := runsAsVersion(good, "1.2.0"); err != nil {
		t.Errorf("a binary reporting the right version failed: %v", err)
	}
	if err := runsAsVersion(good, "1.3.0"); err == nil {
		t.Error("a binary reporting another version passed")
	}
	if err := runsAsVersion(script("garbage", "\x7fELF not really"), "1.2.0"); err == nil {
		t.Error("a binary that can't run passed")
	}
}

func TestReplaceBinaryClearsOldTempFiles(t *testing.T) {
	t.Parallel()

	exe := oldBinary(t)
	dir := filepath.Dir(exe)
	stale := filepath.Join(dir, tempPrefix+"stale")
	fresh := filepath.Join(dir, tempPrefix+"fresh")
	for _, path := range []string{stale, fresh} {
		if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(stale, old, old); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	if err := replaceBinary(exe, []byte("new"), runtime.GOOS, nil); err != nil {
		t.Fatalf("replaceBinary: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Error("a temp file left by a killed update survived")
	}
	if _, err := os.Stat(fresh); err != nil {
		t.Error("removed the temp file of an update that may still be running")
	}
}

// A download killed by quitting must not hold back the next launch.
func TestInstallUpdateRecordsTheAttemptAfterIt(t *testing.T) {
	t.Parallel()

	server, _ := releaseServer(t, "1.2.0", []byte("new"), 0)
	exe := oldBinary(t)
	statePath := filepath.Join(t.TempDir(), "update.json")
	u := testUpdater(server.URL, exe)
	u.verify = func(string, string) error {
		if state := loadUpdateState(statePath); !state.LastInstall.IsZero() {
			t.Errorf("attempt recorded while it was still running: %+v", state)
		}
		return nil
	}

	if err := installUpdate("1.1.0", "1.2.0", false, install{method: installScript}, u, statePath, time.Now(), writableDir); err != nil {
		t.Fatalf("installUpdate: %v", err)
	}
}

func TestQuittingWaitsForAnInstall(t *testing.T) {
	t.Parallel()

	var p pendingInstall
	var out strings.Builder
	p.wait(&out, time.Second)
	if out.Len() != 0 {
		t.Fatalf("waited with nothing running: %q", out.String())
	}

	release := make(chan struct{})
	finished := make(chan struct{})
	go func() {
		_ = p.run("1.2.0", func() error { <-release; return nil })
		close(finished)
	}()
	for {
		p.mu.Lock()
		started := p.done != nil
		p.mu.Unlock()
		if started {
			break
		}
		runtime.Gosched()
	}

	go func() { time.Sleep(50 * time.Millisecond); close(release) }()
	p.wait(&out, 5*time.Second)
	select {
	case <-finished:
	default:
		t.Fatal("wait returned before the install finished")
	}
	if !strings.Contains(out.String(), "1.2.0") {
		t.Fatalf("out = %q, want it to say what it's waiting for", out.String())
	}
}

func TestQuittingGivesUpOnASlowInstall(t *testing.T) {
	t.Parallel()

	var p pendingInstall
	block := make(chan struct{})
	defer close(block)
	go func() { _ = p.run("1.2.0", func() error { <-block; return nil }) }()
	for {
		p.mu.Lock()
		started := p.done != nil
		p.mu.Unlock()
		if started {
			break
		}
		runtime.Gosched()
	}

	var out strings.Builder
	start := time.Now()
	p.wait(&out, 50*time.Millisecond)
	if time.Since(start) > time.Second || !strings.Contains(out.String(), "next time") {
		t.Fatalf("waited %v, out = %q", time.Since(start), out.String())
	}
}
