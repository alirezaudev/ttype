package langcache

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func newTestCache(t *testing.T, handler http.HandlerFunc) *Cache {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	c := New(t.TempDir())
	c.apiBase = server.URL
	c.rawBase = server.URL
	return c.WithHTTPClient(server.Client())
}

func TestListSavesTheManifest(t *testing.T) {
	t.Parallel()

	c := newTestCache(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{"name":"spanish.json"},{"name":"french.json"},{"name":"README.md"}]`))
	})

	ids, err := c.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	want := []string{"french", "spanish"}
	if strings.Join(ids, ",") != strings.Join(want, ",") {
		t.Fatalf("ids = %v, want %v", ids, want)
	}

	if _, err := os.Stat(c.manifestPath()); err != nil {
		t.Fatalf("manifest was not saved: %v", err)
	}
}

func TestListFallsBackToTheSavedManifest(t *testing.T) {
	t.Parallel()

	calls := 0
	c := newTestCache(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Write([]byte(`[{"name":"spanish.json"}]`))
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
	})

	if _, err := c.List(); err != nil {
		t.Fatalf("first List: %v", err)
	}

	stale := time.Now().Add(-30 * 24 * time.Hour)
	if err := os.Chtimes(c.manifestPath(), stale, stale); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	ids, err := c.List()
	if err != nil {
		t.Fatalf("second List: %v", err)
	}
	if len(ids) != 1 || ids[0] != "spanish" {
		t.Fatalf("ids = %v, want [spanish]", ids)
	}
}

func TestWordsDownloadsOnceAndThenReadsFromDisk(t *testing.T) {
	t.Parallel()

	downloads := 0
	c := newTestCache(t, func(w http.ResponseWriter, r *http.Request) {
		downloads++
		w.Write([]byte(`{"words":["uno","dos","tres"]}`))
	})

	if c.Cached("spanish") {
		t.Fatal("spanish should not be cached yet")
	}

	words, err := c.Words("spanish")
	if err != nil {
		t.Fatalf("Words: %v", err)
	}
	if len(words) != 3 {
		t.Fatalf("words = %v, want 3 entries", words)
	}

	if !c.Cached("spanish") {
		t.Fatal("spanish should be cached after the download")
	}

	if _, err := c.Words("spanish"); err != nil {
		t.Fatalf("second Words: %v", err)
	}
	if downloads != 1 {
		t.Fatalf("downloads = %d, want 1", downloads)
	}
}

func TestWordsRejectsAnEmptyWordList(t *testing.T) {
	t.Parallel()

	c := newTestCache(t, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"words":[]}`))
	})

	if _, err := c.Words("spanish"); err == nil {
		t.Fatal("Words should fail on an empty word list")
	}
	if c.Cached("spanish") {
		t.Fatal("a rejected download should not leave a cache entry")
	}
}

func TestDisplayName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		id   string
		want string
	}{
		{id: "", want: "english (built-in)"},
		{id: "spanish", want: "spanish"},
		{id: "code_python", want: "code python"},
	}

	for _, test := range tests {
		if got := DisplayName(test.id); got != test.want {
			t.Errorf("DisplayName(%q) = %q, want %q", test.id, got, test.want)
		}
	}
}
