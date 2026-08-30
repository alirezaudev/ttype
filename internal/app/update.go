package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
)

const (
	releasesAPI    = "https://api.github.com/repos/alirezaudev/ttype/releases/latest"
	updateTimeout  = 30 * time.Second
	maxDownloadMiB = 64
)

var updateHTTPClient = &http.Client{Timeout: updateTimeout}

type release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// LatestRelease resolves the newest published version.
func LatestRelease() (string, error) {
	rel, err := fetchLatestRelease()
	if err != nil {
		return "", err
	}
	return strings.TrimPrefix(rel.TagName, "v"), nil
}

// CheckVersion compares the running binary against the latest release. It is
// best effort: a network failure just means no update notice.
func CheckVersion(local string) domain.VersionInfo {
	info := domain.VersionInfo{Local: local}
	latest, err := LatestRelease()
	if err != nil {
		return info
	}
	info.Latest = latest
	info.UpdateAvailable = newerVersion(latest, local)
	return info
}

func fetchLatestRelease() (release, error) {
	resp, err := updateHTTPClient.Get(releasesAPI)
	if err != nil {
		return release{}, fmt.Errorf("check releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return release{}, fmt.Errorf("check releases: %s", resp.Status)
	}

	var rel release
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return release{}, fmt.Errorf("parse release: %w", err)
	}
	if rel.TagName == "" {
		return release{}, fmt.Errorf("no release published yet")
	}
	return rel, nil
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
	rel, err := fetchLatestRelease()
	if err != nil {
		return err
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	if !newerVersion(latest, local) {
		fmt.Fprintf(out, "ttype %s is already the latest version.\n", local)
		return nil
	}

	name := assetName(latest)
	var assetURL, checksumURL string
	for _, asset := range rel.Assets {
		switch asset.Name {
		case name:
			assetURL = asset.URL
		case "checksums.txt":
			checksumURL = asset.URL
		}
	}
	if assetURL == "" {
		return fmt.Errorf("release %s has no build for %s/%s", latest, runtime.GOOS, runtime.GOARCH)
	}

	fmt.Fprintf(out, "Downloading ttype %s...\n", latest)
	archive, err := download(assetURL)
	if err != nil {
		return err
	}

	if checksumURL != "" {
		if err := verifyChecksum(archive, name, checksumURL); err != nil {
			return err
		}
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
