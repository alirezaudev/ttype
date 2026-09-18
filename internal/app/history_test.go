package app

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/storage"
)

func TestPrintHistoryTableEmpty(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	printHistoryTable(&out, nil)

	if got := out.String(); got != "No test history yet.\n" {
		t.Fatalf("out = %q", got)
	}
}

func TestPrintHistoryTablePrintsNewestFirst(t *testing.T) {
	t.Parallel()

	newest := domain.Result{
		Timestamp: time.Date(2026, 8, 9, 14, 5, 0, 0, time.Local),
		Config:    domain.TestConfig{Kind: domain.TestKindWords, WordCount: 25},
		WPM:       81.4,
		RawWPM:    88.2,
		Accuracy:  94.6,
		Incorrect: 7,
	}
	oldest := domain.Result{
		Timestamp: time.Date(2026, 8, 8, 9, 30, 0, 0, time.Local),
		Config:    domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60},
		WPM:       60,
	}

	var out bytes.Buffer
	printHistoryTable(&out, []domain.Result{newest, oldest})

	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want header + 2 rows:\n%s", len(lines), out.String())
	}

	if !strings.HasPrefix(lines[0], "Date") {
		t.Errorf("header = %q", lines[0])
	}
	if !strings.Contains(lines[1], "Aug 9 14:05") || !strings.Contains(lines[1], "25w") {
		t.Errorf("newest row should come first and show its word count: %q", lines[1])
	}
	if !strings.Contains(lines[1], "81") || !strings.Contains(lines[1], "95%") {
		t.Errorf("newest row should carry its stats (accuracy rounds to 95%%): %q", lines[1])
	}
	if !strings.Contains(lines[2], "Aug 8 09:30") || !strings.Contains(lines[2], "60s") {
		t.Errorf("oldest row should come last and show its duration: %q", lines[2])
	}
}

func TestHistoryLanguageFitsTheColumn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		want string
	}{
		{id: "", want: "english"},
		{id: "spanish", want: "spanish"},
		{id: "english_1k", want: "english 1k"},
		{id: "portuguese_brazilian", want: "portugues…"},
	}

	for _, test := range tests {
		got := historyLanguage(test.id)
		if got != test.want {
			t.Errorf("historyLanguage(%q) = %q, want %q", test.id, got, test.want)
		}
		if n := len([]rune(got)); n > languageColumnWidth {
			t.Errorf("historyLanguage(%q) is %d columns wide, want <= %d", test.id, n, languageColumnWidth)
		}
	}
}

type listStore struct {
	storage.Store
	results []domain.Result
}

func (s listStore) ListResults(limit int) ([]domain.Result, error) {
	if limit > 0 && limit < len(s.results) {
		return s.results[:limit], nil
	}
	return s.results, nil
}

// The limit counts tagged runs, not all of them.
func TestHistoryTagFiltersBeforeTheLimit(t *testing.T) {
	t.Parallel()

	tagged := func(tag string) domain.Result { return domain.Result{Config: domain.TestConfig{Tag: tag}} }
	store := listStore{results: []domain.Result{tagged(""), tagged("english"), tagged(""), tagged("english"), tagged("english")}}

	got, err := listResults(store, 2, "english")
	if err != nil {
		t.Fatalf("listResults: %v", err)
	}
	if len(got) != 2 || got[0].Config.Tag != "english" || got[1].Config.Tag != "english" {
		t.Fatalf("got %+v, want two english runs", got)
	}
}
