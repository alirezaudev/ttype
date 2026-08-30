package app

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"runtime"
	"testing"
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
