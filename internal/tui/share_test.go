package tui

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"
)

func TestWriteOSC52WrapsForMultiplexers(t *testing.T) {
	t.Parallel()

	payload := base64.StdEncoding.EncodeToString([]byte("ttype — 126 WPM"))
	tests := []struct {
		name   string
		env    map[string]string
		prefix string
	}{
		{"plain", nil, "\x1b]52;c;"},
		{"tmux", map[string]string{"TMUX": "/tmp/tmux-1000/default,1,0"}, "\x1bPtmux;\x1b\x1b]52;c;"},
		{"screen", map[string]string{"STY": "1234.pts-0.host"}, "\x1bP\x1b]52;c;"},
	}

	for _, test := range tests {
		var out bytes.Buffer
		getenv := func(key string) string { return test.env[key] }
		if err := writeOSC52(&out, "ttype — 126 WPM", getenv); err != nil {
			t.Fatalf("%s: writeOSC52: %v", test.name, err)
		}
		got := out.String()
		if !strings.HasPrefix(got, test.prefix) {
			t.Errorf("%s: sequence %q, want prefix %q", test.name, got, test.prefix)
		}
		if !strings.Contains(strings.ReplaceAll(got, "\x1b\\\x1bP", ""), payload) {
			t.Errorf("%s: sequence %q is missing the text", test.name, got)
		}
	}
}
