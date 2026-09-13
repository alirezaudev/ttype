package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestNewerVersionComparesNumerically(t *testing.T) {
	t.Parallel()

	tests := []struct {
		latest, local string
		want          bool
	}{
		{"1.2.0", "1.1.9", true},
		{"1.10.0", "1.9.0", true},
		{"1.2.0", "1.2.0", false},
		{"1.2.0", "1.3.0", false},
		{"1.2.1", "1.2", true},
		{"1.0.0", "dev", true},
		{"", "1.0.0", false},
	}

	for _, test := range tests {
		if got := newerVersion(test.latest, test.local); got != test.want {
			t.Errorf("newerVersion(%q, %q) = %v, want %v", test.latest, test.local, got, test.want)
		}
	}
}

func TestAssetNameMatchesTheReleaseLayout(t *testing.T) {
	t.Parallel()

	want := "ttype_1.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	if got := assetName("1.0.0"); got != want {
		t.Fatalf("assetName = %q, want %q", got, want)
	}
}

func tarball(t *testing.T, name string, content []byte) []byte {
	t.Helper()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o755, Size: int64(len(content))}); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar Close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gzip Close: %v", err)
	}
	return buf.Bytes()
}

func TestExtractBinaryFindsTheExecutable(t *testing.T) {
	t.Parallel()

	archive := tarball(t, "ttype_1.0.0/"+binaryName(), []byte("binary"))
	got, err := extractBinary(archive)
	if err != nil {
		t.Fatalf("extractBinary: %v", err)
	}
	if string(got) != "binary" {
		t.Fatalf("extracted %q", got)
	}

	if _, err := extractBinary(tarball(t, "README.md", []byte("hi"))); err == nil {
		t.Fatal("an archive without the binary should be rejected")
	}
}

func TestMatchChecksum(t *testing.T) {
	t.Parallel()

	archive := []byte("payload")
	sum := sha256.Sum256(archive)
	name := "ttype_1.0.0_linux_amd64.tar.gz"
	list := []byte("0000  other.tar.gz\n" + hex.EncodeToString(sum[:]) + "  " + name + "\n")

	if err := matchChecksum(archive, name, list); err != nil {
		t.Fatalf("matchChecksum: %v", err)
	}
	if err := matchChecksum([]byte("tampered"), name, list); err == nil {
		t.Fatal("a tampered archive should be rejected")
	}
	if err := matchChecksum(archive, "missing.tar.gz", list); err == nil {
		t.Fatal("an unlisted asset should be rejected")
	}
}

func TestLatestReleaseReadsTheRedirect(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		location string
		want     string
		wantErr  bool
	}{
		{"tagged release", http.StatusFound, "https://github.com/alirezaudev/ttype/releases/tag/v1.4.0", "1.4.0", false},
		{"no release yet", http.StatusFound, "https://github.com/alirezaudev/ttype/releases", "", true},
		{"tag that is not a version", http.StatusFound, "https://github.com/alirezaudev/ttype/releases/tag/nightly", "", true},
		{"not a redirect", http.StatusOK, "", "", true},
	}

	for _, test := range tests {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodHead || r.URL.Path != "/latest" {
				http.NotFound(w, r)
				return
			}
			if test.location != "" {
				w.Header().Set("Location", test.location)
			}
			w.WriteHeader(test.status)
		}))
		got, err := latestRelease(checkHTTPClient, server.URL)
		server.Close()

		if got != test.want || (err != nil) != test.wantErr {
			t.Errorf("%s: latestRelease = %q, %v; want %q, error %v", test.name, got, err, test.want, test.wantErr)
		}
	}
}

func TestCheckVersionAsksAtMostOnceADay(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "update.json")
	start := time.Date(2026, 9, 14, 20, 0, 0, 0, time.UTC)
	fetches := 0
	fetch := func() (string, error) {
		fetches++
		return "1.1.0", nil
	}

	if info := checkVersion("1.0.0", path, start, fetch); !info.UpdateAvailable || fetches != 1 {
		t.Fatalf("first check = %+v after %d fetches, want an update from one fetch", info, fetches)
	}
	if info := checkVersion("1.0.0", path, start.Add(23*time.Hour), fetch); !info.UpdateAvailable || fetches != 1 {
		t.Fatalf("same day = %+v after %d fetches, want the cached update and no new fetch", info, fetches)
	}
	checkVersion("1.0.0", path, start.Add(25*time.Hour), fetch)
	if fetches != 2 {
		t.Fatalf("fetches = %d the next day, want 2", fetches)
	}
}

func TestCheckVersionOfflineStillWaitsADay(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "update.json")
	start := time.Date(2026, 9, 14, 20, 0, 0, 0, time.UTC)
	fetches := 0
	offline := func() (string, error) {
		fetches++
		return "", errors.New("dial tcp: network is unreachable")
	}

	if info := checkVersion("1.0.0", path, start, offline); info.UpdateAvailable {
		t.Fatalf("offline check = %+v, want no update", info)
	}
	checkVersion("1.0.0", path, start.Add(time.Hour), offline)
	if fetches != 1 {
		t.Fatalf("fetches = %d, want 1: a failed check should still wait a day", fetches)
	}
}

func TestCheckVersionRecordsTheAttemptBeforeAsking(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "update.json")
	now := time.Date(2026, 9, 14, 20, 0, 0, 0, time.UTC)

	checkVersion("1.0.0", path, now, func() (string, error) {
		if got := loadUpdateState(path).LastCheck; !got.Equal(now) {
			t.Errorf("state during the request = %v, want the attempt at %v already saved", got, now)
		}
		return "", errors.New("quit before the request finished")
	})
}

func TestCheckVersionRecoversFromABrokenStateOrClock(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 14, 20, 0, 0, 0, time.UTC)
	fetch := func() (string, error) { return "1.1.0", nil }

	broken := filepath.Join(t.TempDir(), "update.json")
	if err := os.WriteFile(broken, []byte("not json"), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if info := checkVersion("1.0.0", broken, now, fetch); !info.UpdateAvailable {
		t.Fatalf("a broken state file blocked the check: %+v", info)
	}

	future := filepath.Join(t.TempDir(), "update.json")
	if err := saveUpdateState(future, updateState{LastCheck: now.Add(48 * time.Hour)}); err != nil {
		t.Fatalf("saveUpdateState: %v", err)
	}
	if info := checkVersion("1.0.0", future, now, fetch); !info.UpdateAvailable {
		t.Fatalf("a check stamped in the future blocked the check: %+v", info)
	}
}
