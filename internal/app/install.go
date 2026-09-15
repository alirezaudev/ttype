package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
)

// Build describes the running binary.
type Build struct {
	Version string
	// Release is set only when the release pipeline stamped the version.
	Release bool
}

type installMethod int

const (
	installScript installMethod = iota
	installHomebrew
	installScoop
	installSystem
	installGo
	installSource
)

// install says who put the binary there, and so who gets to replace it.
type install struct {
	method installMethod
	by     string
	// upgrade and remove are the commands to use instead of ttype's own.
	upgrade string
	remove  string
}

// owned reports whether ttype may overwrite the binary itself.
func (i install) owned() bool {
	return i.method == installScript
}

// managed reports whether a package manager keeps track of the files.
func (i install) managed() bool {
	switch i.method {
	case installHomebrew, installScoop, installSystem:
		return true
	}
	return false
}

func (i install) updateRefusal() error {
	switch i.method {
	case installSystem:
		return fmt.Errorf("ttype was installed by your system's package manager; update it there")
	case installSource:
		return fmt.Errorf("ttype was built from source; pull and rebuild it to update")
	}
	return fmt.Errorf("ttype was installed with %s; update it with:\n  %s", i.by, i.upgrade)
}

func (i install) removeNote() string {
	if i.remove == "" {
		return fmt.Sprintf("ttype was installed by %s, so its files are left alone; remove it there.", i.by)
	}
	return fmt.Sprintf("ttype was installed with %s, so its files are left alone; remove it with:\n  %s", i.by, i.remove)
}

func currentInstall(build Build) (install, string, error) {
	exe, err := os.Executable()
	if err != nil {
		return install{}, "", fmt.Errorf("locate binary: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	info, _ := debug.ReadBuildInfo()
	return detectInstall(exe, build, info, runtime.GOOS), exe, nil
}

func detectInstall(exe string, build Build, info *debug.BuildInfo, goos string) install {
	path := filepath.ToSlash(exe)
	if goos == "windows" {
		path = strings.ToLower(strings.ReplaceAll(exe, `\`, "/"))
	}

	switch {
	case strings.Contains(path, "/Cellar/"):
		return install{method: installHomebrew, by: "Homebrew", upgrade: "brew upgrade ttype", remove: "brew uninstall ttype"}
	case strings.Contains(path, "/scoop/apps/"):
		return install{method: installScoop, by: "Scoop", upgrade: "scoop update ttype", remove: "scoop uninstall ttype"}
	case goos != "windows" && underSystemDir(path):
		return install{method: installSystem, by: "your system's package manager"}
	}

	if build.Release {
		return install{method: installScript}
	}
	if builtByGoInstall(info) {
		return install{
			method:  installGo,
			by:      "go install",
			upgrade: "go install github.com/alirezaudev/ttype/cmd/ttype@latest",
		}
	}
	return install{method: installSource}
}

func underSystemDir(path string) bool {
	for _, dir := range []string{"/usr/bin/", "/usr/sbin/", "/bin/", "/sbin/", "/nix/store/", "/snap/"} {
		if strings.HasPrefix(path, dir) {
			return true
		}
	}
	return false
}

// go install @version builds from the module cache, which carries a module
// version but no VCS details; a build inside a checkout has both.
func builtByGoInstall(info *debug.BuildInfo) bool {
	if info == nil || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return false
	}
	for _, setting := range info.Settings {
		if setting.Key == "vcs.revision" {
			return false
		}
	}
	return true
}
