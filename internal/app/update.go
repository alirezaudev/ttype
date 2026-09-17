package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
)

const (
	releasesURL    = "https://github.com/alirezaudev/ttype/releases"
	checkInterval  = 24 * time.Hour
	checkTimeout   = 10 * time.Second
	maxDownloadMiB = 64
	// A slow link can take minutes over the archive; this only stops a
	// download that trickles forever.
	maxDownloadTime = 15 * time.Minute
)

// stallTimeout abandons a download that stops moving.
var stallTimeout = 30 * time.Second

var updateHTTPClient = &http.Client{Timeout: maxDownloadTime}

// The redirect is the answer, so it must not be followed.
var checkHTTPClient = &http.Client{
	Timeout: checkTimeout,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// CheckVersion compares the running binary against the latest release, asking
// the network at most once a day, and says whether ttype can install it
// itself. It is best effort: a failure just means no update notice.
func CheckVersion(build Build, mode domain.UpdateMode) domain.VersionInfo {
	dirs, err := storage.DefaultDirs()
	if err != nil {
		return domain.VersionInfo{Local: build.Version}
	}
	fetch := func() (string, error) { return latestRelease(checkHTTPClient, releasesURL) }
	info := checkVersion(build.Version, updateStatePath(dirs.Data), time.Now(), fetch)
	if !info.UpdateAvailable {
		return info
	}
	inst, exe, err := currentInstall(build)
	if err != nil {
		return info
	}
	return installPlan(info, inst, exe, runtime.GOOS, mode, dirWritable)
}

func checkVersion(local, statePath string, now time.Time, fetch func() (string, error)) domain.VersionInfo {
	state := loadUpdateState(statePath)

	// A background install is announced on the first launch that runs it.
	justUpdated := false
	if state.Installed != "" && !newerVersion(state.Installed, local) {
		justUpdated = state.Installed == local
		state.Installed = ""
		_ = saveUpdateState(statePath, state)
	}

	// A clock set back must not postpone the next check forever.
	if now.Sub(state.LastCheck) >= checkInterval || now.Before(state.LastCheck) {
		// Record the attempt before asking, so being offline still waits a day.
		state.LastCheck = now
		_ = saveUpdateState(statePath, state)
		if latest, err := fetch(); err == nil {
			state.Latest = latest
			_ = saveUpdateState(statePath, state)
		}
	}
	return domain.VersionInfo{
		Local:           local,
		Latest:          state.Latest,
		UpdateAvailable: newerVersion(state.Latest, local),
		JustUpdated:     justUpdated,
	}
}

// latestRelease reads the newest tag from the releases/latest redirect, which,
// unlike the API, has no per-IP rate limit.
func latestRelease(client *http.Client, base string) (string, error) {
	req, err := http.NewRequest(http.MethodHead, base+"/latest", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "ttype")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("check releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 || resp.StatusCode > 399 {
		return "", fmt.Errorf("check releases: %s", resp.Status)
	}
	_, tag, _ := strings.Cut(resp.Header.Get("Location"), "/releases/tag/")
	version := strings.TrimPrefix(tag, "v")
	if versionParts(version) == nil {
		return "", fmt.Errorf("no release published yet")
	}
	return version, nil
}

func releaseAssetURL(base, version, name string) string {
	return fmt.Sprintf("%s/download/v%s/%s", base, version, name)
}

// newerVersion compares dotted versions numerically, so 1.10.0 beats 1.9.0.
// A local build that is not a version ("dev") always counts as older.
func newerVersion(latest, local string) bool {
	if latest == "" {
		return false
	}
	remote := versionParts(latest)
	if remote == nil {
		return false
	}
	current := versionParts(local)
	if current == nil {
		return true
	}

	for i := 0; i < len(remote) && i < len(current); i++ {
		if remote[i] != current[i] {
			return remote[i] > current[i]
		}
	}
	return len(remote) > len(current)
}

func versionParts(v string) []int {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	if v == "" {
		return nil
	}

	var parts []int
	for _, field := range strings.Split(v, ".") {
		n, err := strconv.Atoi(field)
		if err != nil {
			return nil
		}
		parts = append(parts, n)
	}
	return parts
}

func assetName(version string) string {
	return fmt.Sprintf("ttype_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)
}

func RunUpdate(out io.Writer, build Build) error {
	inst, exe, err := currentInstall(build)
	if err != nil {
		return err
	}
	return runUpdate(out, build.Version, inst, newUpdater(exe))
}

func runUpdate(out io.Writer, local string, inst install, u updater) error {
	// A package manager would lose track of a binary replaced behind its back.
	if !inst.owned() {
		return inst.updateRefusal()
	}

	latest, err := latestRelease(checkHTTPClient, u.base)
	if err != nil {
		return err
	}

	if !newerVersion(latest, local) {
		fmt.Fprintf(out, "ttype %s is already the latest version.\n", local)
		return nil
	}

	fmt.Fprintf(out, "Downloading ttype %s...\n", latest)
	if err := u.install(latest); err != nil {
		return err
	}
	fmt.Fprintf(out, "Updated to ttype %s.\n", latest)
	return nil
}

// updater installs a release over one binary.
type updater struct {
	base string
	exe  string
	goos string
	// verify runs the new binary before it replaces the old one.
	verify func(path, version string) error
}

func newUpdater(exe string) updater {
	return updater{base: releasesURL, exe: exe, goos: runtime.GOOS, verify: runsAsVersion}
}

func (u updater) install(version string) error {
	name := assetName(version)
	archive, err := download(releaseAssetURL(u.base, version, name))
	if err != nil {
		return fmt.Errorf("release %s for %s/%s: %w", version, runtime.GOOS, runtime.GOARCH, err)
	}
	if err := verifyChecksum(archive, name, releaseAssetURL(u.base, version, "checksums.txt")); err != nil {
		return err
	}
	binary, err := extractBinary(archive)
	if err != nil {
		return err
	}

	var check func(string) error
	if u.verify != nil {
		check = func(path string) error { return u.verify(path, version) }
	}
	return replaceBinary(u.exe, binary, u.goos, check)
}

// runsAsVersion starts the new binary with --version, so a build that can't
// run here (a wrong architecture, a noexec mount) never replaces one that can.
func runsAsVersion(path, version string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, path, "--version").Output()
	if err != nil {
		return fmt.Errorf("new binary did not run: %w", err)
	}
	if !strings.Contains(string(out), version) {
		return fmt.Errorf("new binary reports %q, want %s", strings.TrimSpace(string(out)), version)
	}
	return nil
}

func download(url string) ([]byte, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stalled := time.AfterFunc(stallTimeout, cancel)
	defer stalled.Stop()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	resp, err := updateHTTPClient.Do(req)
	if err != nil {
		return nil, downloadError(ctx, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download: %s", resp.Status)
	}

	body := progressReader{
		r:        io.LimitReader(resp.Body, maxDownloadMiB<<20),
		progress: func() { stalled.Reset(stallTimeout) },
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, downloadError(ctx, err)
	}
	return data, nil
}

func downloadError(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return fmt.Errorf("download: nothing received for %s", stallTimeout)
	}
	return fmt.Errorf("download: %w", err)
}

type progressReader struct {
	r        io.Reader
	progress func()
}

func (p progressReader) Read(b []byte) (int, error) {
	n, err := p.r.Read(b)
	if n > 0 {
		p.progress()
	}
	return n, err
}

func verifyChecksum(archive []byte, name, checksumURL string) error {
	list, err := download(checksumURL)
	if err != nil {
		return err
	}
	return matchChecksum(archive, name, list)
}

func matchChecksum(archive []byte, name string, list []byte) error {
	want := ""
	for _, line := range strings.Split(string(list), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			want = fields[0]
			break
		}
	}
	if want == "" {
		return fmt.Errorf("no checksum published for %s", name)
	}

	sum := sha256.Sum256(archive)
	if got := hex.EncodeToString(sum[:]); got != want {
		return fmt.Errorf("checksum mismatch for %s", name)
	}
	return nil
}

func extractBinary(archive []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("open archive: %w", err)
	}
	defer gz.Close()

	reader := tar.NewReader(gz)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg || filepath.Base(header.Name) != binaryName() {
			continue
		}
		return io.ReadAll(io.LimitReader(reader, maxDownloadMiB<<20))
	}
	return nil, fmt.Errorf("archive has no ttype binary")
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "ttype.exe"
	}
	return "ttype"
}

const tempPrefix = ".ttype-update-"

// replaceBinary writes next to the current binary and renames over it, so a
// failed write never leaves a half-written executable behind.
func replaceBinary(current string, binary []byte, goos string, check func(string) error) error {
	removeStaleTemps(filepath.Dir(current))

	// Windows only runs a file that ends in .exe
	tmp, err := os.CreateTemp(filepath.Dir(current), tempPrefix+"*"+filepath.Ext(current))
	if err != nil {
		return fmt.Errorf("write update: %w (is the install dir writable?)", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(binary); err != nil {
		tmp.Close()
		return fmt.Errorf("write update: %w", err)
	}
	// A short write surfaces on Close, and a truncated binary must not pass.
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("write update: %w", err)
	}
	if err := os.Chmod(tmp.Name(), 0o755); err != nil {
		return fmt.Errorf("write update: %w", err)
	}
	if check != nil {
		if err := check(tmp.Name()); err != nil {
			return err
		}
	}

	if goos == "windows" {
		return swapRunningExe(tmp.Name(), current)
	}
	if err := os.Rename(tmp.Name(), current); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}

// A run killed mid-update can leave its temp file behind. Only old ones go, so
// another ttype updating right now keeps its own.
func removeStaleTemps(dir string) {
	matches, _ := filepath.Glob(filepath.Join(dir, tempPrefix+"*"))
	for _, path := range matches {
		if info, err := os.Stat(path); err == nil && time.Since(info.ModTime()) > time.Hour {
			_ = os.Remove(path)
		}
	}
}

// Windows refuses to overwrite a running .exe but lets it be renamed, so the
// old one steps aside and is deleted on the next start.
func swapRunningExe(next, current string) error {
	old := oldBinaryPath(current)
	_ = os.Remove(old)
	if err := os.Rename(current, old); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	if err := os.Rename(next, current); err != nil {
		_ = os.Rename(old, current)
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}

func oldBinaryPath(current string) string {
	return current + ".old"
}

// RemoveOldBinary deletes the copy a Windows update left behind. It is still
// running during the update, so this waits for the next start.
func RemoveOldBinary() {
	if exe, err := os.Executable(); err == nil {
		_ = os.Remove(oldBinaryPath(exe))
	}
}
