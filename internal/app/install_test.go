package app

import (
	"runtime/debug"
	"testing"
)

func TestDetectInstall(t *testing.T) {
	t.Parallel()

	release := Build{Version: "1.0.1", Release: true}
	goInstall := &debug.BuildInfo{Main: debug.Module{Version: "v1.0.1"}}
	checkout := &debug.BuildInfo{
		Main:     debug.Module{Version: "v1.0.2-0.20260914000000-8fb52bd1a2b3+dirty"},
		Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "8fb52bd"}},
	}

	tests := []struct {
		name  string
		exe   string
		build Build
		info  *debug.BuildInfo
		goos  string
		want  installMethod
	}{
		{"install.sh", "/usr/local/bin/ttype", release, nil, "linux", installScript},
		{"install.sh in home", "/home/me/.local/bin/ttype", release, nil, "linux", installScript},
		{"homebrew on apple silicon", "/opt/homebrew/Cellar/ttype/1.0.1/bin/ttype", release, nil, "darwin", installHomebrew},
		{"linuxbrew", "/home/linuxbrew/.linuxbrew/Cellar/ttype/1.0.1/bin/ttype", release, nil, "linux", installHomebrew},
		{"scoop", `C:\Users\me\scoop\apps\ttype\1.0.1\ttype.exe`, release, nil, "windows", installScoop},
		{"windows download", `C:\Tools\ttype.exe`, release, nil, "windows", installScript},
		{"aur", "/usr/bin/ttype", release, nil, "linux", installSystem},
		{"nix", "/nix/store/abc-ttype-1.0.1/bin/ttype", release, nil, "linux", installSystem},
		{"go install", "/home/me/go/bin/ttype", Build{Version: "1.0.1"}, goInstall, "linux", installGo},
		{"go install at main", "/home/me/go/bin/ttype", Build{Version: "dev"}, &debug.BuildInfo{Main: debug.Module{Version: "v1.0.2-0.20260914000000-8fb52bd1a2b3"}}, "linux", installGo},
		{"make build", "/home/me/ttype/bin/ttype", Build{Version: "dev"}, checkout, "linux", installSource},
		{"go run", "/tmp/go-build1/exe/ttype", Build{Version: "dev"}, &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, "linux", installSource},
	}

	for _, test := range tests {
		if got := detectInstall(test.exe, test.build, test.info, test.goos); got.method != test.want {
			t.Errorf("%s: method = %d, want %d", test.name, got.method, test.want)
		}
	}
}
