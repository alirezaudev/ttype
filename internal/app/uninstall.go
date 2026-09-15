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
	// binary is empty when a package manager owns it.
	binary   string
	manPages []string
	dirs     [][2]string
	// managedBy explains why the binary is left alone.
	managedBy string
}

// RunUninstall removes the running binary and its man page. Settings and
// history stay unless purge is set: uninstalling should not lose your records
// unless you say so.
func RunUninstall(out io.Writer, in io.Reader, build Build, purge, yes bool) error {
	inst, exe, err := currentInstall(build)
	if err != nil {
		return err
	}
	targets, err := resolveUninstallTargets(inst, exe)
	if err != nil {
		return err
	}
	return uninstall(targets, out, in, purge, yes)
}

func resolveUninstallTargets(inst install, exe string) (uninstallTargets, error) {
	dirs, err := storage.DefaultDirs()
	if err != nil {
		return uninstallTargets{}, err
	}
	targets := uninstallTargets{dirs: [][2]string{{"config", dirs.Config}, {"data", dirs.Data}}}

	// Removing files a package manager tracks would leave it out of sync.
	if inst.managed() {
		targets.managedBy = inst.removeNote()
		return targets, nil
	}

	targets.binary = exe
	if runtime.GOOS != "windows" {
		targets.manPages = append(targets.manPages, "/usr/local/share/man/man1/ttype.1")
		if home, err := os.UserHomeDir(); err == nil {
			targets.manPages = append(targets.manPages,
				filepath.Join(home, ".local", "share", "man", "man1", "ttype.1"))
		}
	}
	return targets, nil
}

func uninstall(targets uninstallTargets, out io.Writer, in io.Reader, purge, yes bool) error {
	if targets.managedBy != "" {
		fmt.Fprintln(out, targets.managedBy)
	}
	if !purge {
		if targets.binary == "" {
			fmt.Fprintln(out, "Your settings and history are kept; pass --purge to delete them.")
			return nil
		}
		fmt.Fprintln(out, "Your settings and history are kept; pass --purge to delete them too.")
	}

	var dirs [][2]string
	if purge {
		for _, dir := range targets.dirs {
			if _, err := os.Stat(dir[1]); err == nil {
				dirs = append(dirs, dir)
			}
		}
	}
	manPages := existing(targets.manPages)
	if targets.binary == "" && len(manPages) == 0 && len(dirs) == 0 {
		fmt.Fprintln(out, "Nothing to remove.")
		return nil
	}

	fmt.Fprintln(out, "This will remove:")
	if targets.binary != "" {
		fmt.Fprintf(out, "  binary  %s\n", targets.binary)
	}
	for _, man := range manPages {
		fmt.Fprintf(out, "  man     %s\n", man)
	}
	for _, dir := range dirs {
		fmt.Fprintf(out, "  %-7s %s\n", dir[0], dir[1])
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

	var paths []string
	if targets.binary != "" {
		paths = append(paths, targets.binary)
	}
	paths = append(paths, manPages...)
	for _, dir := range dirs {
		paths = append(paths, dir[1])
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
