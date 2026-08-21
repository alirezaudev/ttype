package tui

import (
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func TestFooterLinksAreClickable(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Theme: ThemeDefault}
	ver := domain.VersionInfo{Local: "1.2.0"}
	l := footerLayoutFor(cfg, ver, 80)

	col := l.offset
	for _, link := range footerLinks() {
		url, ok := footerURLAt(col, 21, 80, 24, cfg, ver)
		if !ok || url != link.url {
			t.Fatalf("click at %d = %q/%v, want %q", col, url, ok, link.url)
		}
		col += len(link.label) + 2
	}

	if _, ok := footerURLAt(0, 10, 80, 24, cfg, ver); ok {
		t.Fatal("a click outside the footer row hit a link")
	}
}

func TestFooterVersionLinkGoesToReleases(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Theme: ThemeDefault}
	ver := domain.VersionInfo{Local: "1.2.0", UpdateAvailable: true}
	l := footerLayoutFor(cfg, ver, 80)

	x := l.offset + l.leftWidth + l.gap + len(cfg.Theme) + 2
	url, ok := footerURLAt(x, 21, 80, 24, cfg, ver)
	if !ok || url != GitHubURL+"/releases" {
		t.Fatalf("version click = %q/%v, want the releases page", url, ok)
	}
}

func TestZenHidesTheFooter(t *testing.T) {
	t.Parallel()

	cfg := domain.TestConfig{Theme: ThemeDefault, Zen: true}
	if got := renderFooterSection(defaultTheme(), cfg, domain.VersionInfo{}, 80); got != "" {
		t.Fatalf("zen footer = %q, want empty", got)
	}
	if _, ok := footerURLAt(0, 21, 80, 24, cfg, domain.VersionInfo{}); ok {
		t.Fatal("zen footer is clickable")
	}
}

func TestFooterMarksAnAvailableUpdate(t *testing.T) {
	t.Parallel()

	plain := versionLabel(domain.VersionInfo{Local: "1.0.0"})
	updated := versionLabel(domain.VersionInfo{Local: "1.0.0", UpdateAvailable: true})

	if strings.Contains(plain, "new") {
		t.Fatalf("version label = %q, want no update marker", plain)
	}
	if !strings.Contains(updated, "new") {
		t.Fatalf("version label = %q, want an update marker", updated)
	}
}
