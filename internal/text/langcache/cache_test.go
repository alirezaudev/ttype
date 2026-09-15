package langcache

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// Restarting a test asks for the words again; that must not re-read the file.
func TestWordsKeepsTheListInMemory(t *testing.T) {
	t.Parallel()

	list := `{"words":["uno","dos","tres"]}`
	downloads := 0
	c := newTestCache(t, func(w http.ResponseWriter, r *http.Request) {
		downloads++
		w.Write([]byte(list))
	})

	if _, err := c.Words("spanish"); err != nil {
		t.Fatalf("Words: %v", err)
	}
	if err := os.Remove(c.languagePath("spanish")); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	words, err := c.Words("spanish")
	if err != nil || len(words) != 3 {
		t.Fatalf("second Words = %v, %v; want the list from memory", words, err)
	}
	if downloads != 1 {
		t.Fatalf("downloads = %d, want 1", downloads)
	}

	// A fresh download replaces what is in memory.
	list = `{"words":["cuatro"]}`
	if err := c.download("spanish"); err != nil {
		t.Fatalf("download: %v", err)
	}
	if words, _ := c.Words("spanish"); len(words) != 1 || words[0] != "cuatro" {
		t.Fatalf("words after download = %v, want [cuatro]", words)
	}
}

func TestInstalledListsOnlyDownloadedLanguages(t *testing.T) {
	t.Parallel()

	c := New(t.TempDir())
	if err := os.MkdirAll(c.Dir(), 0o700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	for _, name := range []string{"spanish.json", manifestFile, "french.json.tmp"} {
		if err := os.WriteFile(filepath.Join(c.Dir(), name), []byte("{}"), 0o600); err != nil {
			t.Fatalf("WriteFile: %v", err)
		}
	}

	if got := c.Installed(); len(got) != 1 || got[0] != "spanish" {
		t.Fatalf("Installed = %v, want [spanish]", got)
	}
	if got := New(t.TempDir()).Installed(); len(got) != 0 {
		t.Fatalf("Installed with no cache dir = %v, want none", got)
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

func TestDownloadAllFetchesEverythingThatIsNotCached(t *testing.T) {
	t.Parallel()

	c := newTestCache(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/repos/") {
			w.Write([]byte(`[{"name":"spanish.json","size":100},{"name":"french.json","size":200},{"name":"README.md","size":9999}]`))
			return
		}
		w.Write([]byte(`{"words":["uno","dos"]}`))
	})

	plan, err := c.PlanDownloadAll()
	if err != nil {
		t.Fatalf("PlanDownloadAll: %v", err)
	}
	if plan.TotalRemote != 2 || len(plan.Entries) != 2 || plan.AlreadyCached != 0 {
		t.Fatalf("plan = %+v, want 2 remote / 2 to download / 0 cached", plan)
	}
	if plan.TotalBytes != 300 {
		t.Fatalf("TotalBytes = %d, want 300", plan.TotalBytes)
	}

	done := 0
	if failures := c.DownloadAll(plan, 2, func(int, int, string, error) { done++ }); len(failures) != 0 {
		t.Fatalf("DownloadAll: %v", failures)
	}
	if done != 2 {
		t.Fatalf("progress calls = %d, want 2", done)
	}

	replan, err := c.PlanDownloadAll()
	if err != nil {
		t.Fatalf("second PlanDownloadAll: %v", err)
	}
	if len(replan.Entries) != 0 || replan.AlreadyCached != 2 {
		t.Fatalf("replan = %+v, want nothing left to download", replan)
	}
}

func TestFormatSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		bytes int64
		want  string
	}{
		{bytes: 512, want: "512 B"},
		{bytes: 2 * 1024, want: "2 KB"},
		{bytes: 400 * 1024 * 1024, want: "400 MB"},
		{bytes: 3 * 1024 * 1024 * 1024, want: "3.0 GB"},
	}

	for _, test := range tests {
		if got := FormatSize(test.bytes); got != test.want {
			t.Errorf("FormatSize(%d) = %q, want %q", test.bytes, got, test.want)
		}
	}
}
