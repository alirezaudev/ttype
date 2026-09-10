package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func uninstallFixture(t *testing.T) uninstallTargets {
	t.Helper()

	dir := t.TempDir()
	binary := filepath.Join(dir, "ttype")
	man := filepath.Join(dir, "ttype.1")
	config := filepath.Join(dir, "config")
	data := filepath.Join(dir, "data")

	for _, path := range []string{binary, man} {
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}
	for _, path := range []string{config, data} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
	}

	return uninstallTargets{
		binary:   binary,
		manPages: []string{man},
		dirs:     [][2]string{{"config", config}, {"data", data}},
	}
}

func TestUninstallKeepsDataWithoutPurge(t *testing.T) {
	t.Parallel()

	targets := uninstallFixture(t)
	var out bytes.Buffer
	if err := uninstall(targets, &out, strings.NewReader(""), false, true); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	if _, err := os.Stat(targets.binary); !os.IsNotExist(err) {
		t.Error("the binary survived")
	}
	if _, err := os.Stat(targets.manPages[0]); !os.IsNotExist(err) {
		t.Error("the man page survived")
	}
	for _, dir := range targets.dirs {
		if _, err := os.Stat(dir[1]); err != nil {
			t.Errorf("%s was deleted without --purge", dir[0])
		}
	}
}

func TestUninstallPurgeRemovesEverything(t *testing.T) {
	t.Parallel()

	targets := uninstallFixture(t)
	var out bytes.Buffer
	if err := uninstall(targets, &out, strings.NewReader(""), true, true); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	for _, dir := range targets.dirs {
		if _, err := os.Stat(dir[1]); !os.IsNotExist(err) {
			t.Errorf("%s survived --purge", dir[0])
		}
	}
}

func TestUninstallStopsWhenDeclined(t *testing.T) {
	t.Parallel()

	targets := uninstallFixture(t)
	var out bytes.Buffer
	if err := uninstall(targets, &out, strings.NewReader("n\n"), true, false); err != nil {
		t.Fatalf("uninstall: %v", err)
	}

	if _, err := os.Stat(targets.binary); err != nil {
		t.Error("the binary was removed despite the answer being no")
	}
	if !strings.Contains(out.String(), "Nothing was removed") {
		t.Errorf("output = %q", out.String())
	}
}
