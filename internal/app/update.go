package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
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
	updateTimeout  = 30 * time.Second
	maxDownloadMiB = 64
)

var updateHTTPClient = &http.Client{Timeout: updateTimeout}

// The redirect is the answer, so it must not be followed.
var checkHTTPClient = &http.Client{
	Timeout: checkTimeout,
	CheckRedirect: func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

// CheckVersion compares the running binary against the latest release, asking
// the network at most once a day. It is best effort: a failure just means no
// update notice.
func CheckVersion(local string) domain.VersionInfo {
	dirs, err := storage.DefaultDirs()
	if err != nil {
		return domain.VersionInfo{Local: local}
	}
	fetch := func() (string, error) { return latestRelease(checkHTTPClient, releasesURL) }
	return checkVersion(local, updateStatePath(dirs.Data), time.Now(), fetch)
}

func checkVersion(local, statePath string, now time.Time, fetch func() (string, error)) domain.VersionInfo {
	state := loadUpdateState(statePath)
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

func RunUpdate(out io.Writer, local string) error {
	latest, err := latestRelease(checkHTTPClient, releasesURL)
	if err != nil {
		return err
	}

	if !newerVersion(latest, local) {
		fmt.Fprintf(out, "ttype %s is already the latest version.\n", local)
		return nil
	}

	name := assetName(latest)
	fmt.Fprintf(out, "Downloading ttype %s...\n", latest)
	archive, err := download(releaseAssetURL(releasesURL, latest, name))
	if err != nil {
		return fmt.Errorf("release %s for %s/%s: %w", latest, runtime.GOOS, runtime.GOARCH, err)
	}

	if err := verifyChecksum(archive, name, releaseAssetURL(releasesURL, latest, "checksums.txt")); err != nil {
		return err
	}

	binary, err := extractBinary(archive)
	if err != nil {
		return err
	}
	if err := replaceRunningBinary(binary); err != nil {
		return err
	}

	fmt.Fprintf(out, "Updated to ttype %s.\n", latest)
	return nil
}

func download(url string) ([]byte, error) {
	resp, err := updateHTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download: %s", resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxDownloadMiB<<20))
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

// replaceRunningBinary writes next to the current binary and renames over it,
// so a failed write never leaves a half-written executable behind.
func replaceRunningBinary(binary []byte) error {
	current, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locate binary: %w", err)
	}
	current, err = filepath.EvalSymlinks(current)
	if err != nil {
		return fmt.Errorf("resolve binary: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(current), ".ttype-update-")
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
	if err := os.Rename(tmp.Name(), current); err != nil {
		return fmt.Errorf("replace binary: %w", err)
	}
	return nil
}
