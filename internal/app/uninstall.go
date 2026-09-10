package app

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/alirezaudev/ttype/internal/storage"
)

type uninstallTargets struct {
	binary   string
	manPages []string
	dirs     [][2]string
}

// RunUninstall removes the running binary and its man page. Settings and
// history stay unless purge is set: uninstalling should not lose your records
// unless you say so.
func RunUninstall(out io.Writer, in io.Reader, purge, yes bool) error {
	targets, err := resolveUninstallTargets()
	if err != nil {
		return err
	}
	return uninstall(targets, out, in, purge, yes)
}

func resolveUninstallTargets() (uninstallTargets, error) {
	self, err := os.Executable()
	if err != nil {
		return uninstallTargets{}, fmt.Errorf("locate binary: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(self); err == nil {
		self = resolved
	}

	targets := uninstallTargets{binary: self}

	if runtime.GOOS != "windows" {
		for _, prefix := range []string{"/usr/local", "/usr"} {
			targets.manPages = append(targets.manPages,
				filepath.Join(prefix, "share", "man", "man1", "ttype.1"))
		}
		if home, err := os.UserHomeDir(); err == nil {
			targets.manPages = append(targets.manPages,
				filepath.Join(home, ".local", "share", "man", "man1", "ttype.1"))
		}
	}

	dirs, err := storage.DefaultDirs()
	if err != nil {
		return uninstallTargets{}, err
	}
	targets.dirs = [][2]string{{"config", dirs.Config}, {"data", dirs.Data}}

	return targets, nil
}

func uninstall(targets uninstallTargets, out io.Writer, in io.Reader, purge, yes bool) error {
	fmt.Fprintln(out, "This will remove:")
	fmt.Fprintf(out, "  binary  %s\n", targets.binary)
	for _, man := range existing(targets.manPages) {
		fmt.Fprintf(out, "  man     %s\n", man)
	}

	if purge {
		for _, dir := range targets.dirs {
			if _, err := os.Stat(dir[1]); err == nil {
				fmt.Fprintf(out, "  %-7s %s\n", dir[0], dir[1])
			}
		}
	} else {
		fmt.Fprintln(out, "Your settings and history are kept; pass --purge to delete them too.")
	}

	if !yes {
		confirmed, err := confirm(out, in, "Continue? [y/N] ")
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(out, "Nothing was removed.")
			return nil
		}
	}

	paths := append([]string{targets.binary}, existing(targets.manPages)...)
	if purge {
		for _, dir := range targets.dirs {
			paths = append(paths, dir[1])
		}
	}

	var failures []string
	for _, path := range paths {
		gone, err := removePath(path)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %s", path, err))
			continue
		}
		if gone {
			fmt.Fprintf(out, "  removed %s\n", path)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("could not remove %d path(s):\n  %s",
			len(failures), joinLines(failures))
	}
	return nil
}

func existing(paths []string) []string {
	var out []string
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			out = append(out, path)
		}
	}
	return out
}

func joinLines(items []string) string {
	out := ""
	for i, item := range items {
		if i > 0 {
			out += "\n  "
		}
		out += item
	}
	return out
}
