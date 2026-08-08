package langcache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultRepo    = "alirezaudev/monkeytype"
	defaultBranch  = "master"
	languagesPath  = "frontend/static/languages"
	manifestFile   = "_manifest.json"
	manifestMaxAge = 7 * 24 * time.Hour
	fetchTimeout   = 60 * time.Second

	defaultDownloadWorkers = 8
)

type Entry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size,omitempty"`
}

type DownloadPlan struct {
	Entries       []Entry
	TotalBytes    int64
	AlreadyCached int
	TotalRemote   int
}

type Cache struct {
	dir        string
	repo       string
	branch     string
	rawBase    string
	apiBase    string
	httpClient *http.Client
}

func New(dataDir string) *Cache {
	return &Cache{
		dir:        filepath.Join(dataDir, "languages"),
		repo:       defaultRepo,
		branch:     defaultBranch,
		rawBase:    "https://raw.githubusercontent.com",
		apiBase:    "https://api.github.com",
		httpClient: &http.Client{Timeout: fetchTimeout},
	}
}

func (c *Cache) WithHTTPClient(client *http.Client) *Cache {
	if client != nil {
		c.httpClient = client
	}
	return c
}

func (c *Cache) Dir() string { return c.dir }

func DisplayName(id string) string {
	if id == "" {
		return "english (built-in)"
	}
	return strings.ReplaceAll(id, "_", " ")
}

func (c *Cache) languagePath(id string) string {
	return filepath.Join(c.dir, id+".json")
}

func (c *Cache) manifestPath() string {
	return filepath.Join(c.dir, manifestFile)
}

func (c *Cache) Cached(id string) bool {
	if id == "" {
		return true
	}
	_, err := os.Stat(c.languagePath(id))
	return err == nil
}

func (c *Cache) Words(id string) ([]string, error) {
	if id == "" {
		return nil, fmt.Errorf("built-in language has no cache entry")
	}
	if err := c.Ensure(id); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(c.languagePath(id))
	if err != nil {
		return nil, err
	}
	return parseWords(data)
}

func (c *Cache) Ensure(id string) error {
	if id == "" || c.Cached(id) {
		return nil
	}
	return c.download(id)
}

func (c *Cache) List() ([]string, error) {
	entries, err := c.ListAvailable()
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.ID)
	}
	return ids, err
}

func (c *Cache) ListAvailable() ([]Entry, error) {
	cached, err := c.loadManifest()
	if err == nil && len(cached) > 0 && !c.manifestStale() {
		return cached, nil
	}

	fresh, fetchErr := c.fetchManifest()
	if fetchErr != nil {
		if len(cached) > 0 {
			return cached, nil
		}
		return nil, fetchErr
	}

	if err := c.saveManifest(fresh); err != nil {
		return fresh, err
	}
	return fresh, nil
}

func (c *Cache) PlanDownloadAll() (DownloadPlan, error) {
	entries, err := c.fetchManifest()
	if err != nil {
		return DownloadPlan{}, err
	}
	if err := c.saveManifest(entries); err != nil {
		return DownloadPlan{}, err
	}

	plan := DownloadPlan{TotalRemote: len(entries)}
	for _, entry := range entries {
		if c.Cached(entry.ID) {
			plan.AlreadyCached++
			continue
		}
		plan.Entries = append(plan.Entries, entry)
		plan.TotalBytes += entry.Size
	}
	return plan, nil
}

func (c *Cache) DownloadAll(plan DownloadPlan, workers int, progress func(done, total int, id string, err error)) []error {
	total := len(plan.Entries)
	if total == 0 {
		return nil
	}
	if workers <= 0 {
		workers = defaultDownloadWorkers
	}
	if workers > total {
		workers = total
	}

	jobs := make(chan Entry, total)
	for _, entry := range plan.Entries {
		jobs <- entry
	}
	close(jobs)

	var (
		mu       sync.Mutex
		failures []error
		done     atomic.Int32
		wg       sync.WaitGroup
	)
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for entry := range jobs {
				err := c.Ensure(entry.ID)
				if progress != nil {
					progress(int(done.Add(1)), total, entry.ID, err)
				}
				if err == nil {
					continue
				}
				mu.Lock()
				failures = append(failures, fmt.Errorf("%s: %w", entry.ID, err))
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	return failures
}

func FormatSize(bytes int64) string {
	switch {
	case bytes >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1<<30))
	case bytes >= 1<<20:
		return fmt.Sprintf("%.0f MB", float64(bytes)/(1<<20))
	case bytes >= 1<<10:
		return fmt.Sprintf("%.0f KB", float64(bytes)/(1<<10))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func (c *Cache) manifestStale() bool {
	info, err := os.Stat(c.manifestPath())
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > manifestMaxAge
}

func (c *Cache) loadManifest() ([]Entry, error) {
	data, err := os.ReadFile(c.manifestPath())
	if err != nil {
		return nil, err
	}

	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (c *Cache) saveManifest(entries []Entry) error {
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		return err
	}

	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return os.WriteFile(c.manifestPath(), data, 0o600)
}

func (c *Cache) fetchManifest() ([]Entry, error) {
	url := fmt.Sprintf("%s/repos/%s/contents/%s?ref=%s&per_page=1000", c.apiBase, c.repo, languagesPath, c.branch)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ttype")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list languages failed: %s: %s", resp.Status, snippet(resp.Body))
	}

	var items []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var entries []Entry
	for _, item := range items {
		if !strings.HasSuffix(item.Name, ".json") || item.Name == manifestFile {
			continue
		}
		id := strings.TrimSuffix(item.Name, ".json")
		entries = append(entries, Entry{ID: id, Name: DisplayName(id), Size: item.Size})
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("list languages returned no entries")
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries, nil
}

func (c *Cache) download(id string) error {
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s/%s/%s/%s.json", c.rawBase, c.repo, c.branch, languagesPath, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ttype")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %q failed: %s: %s", id, resp.Status, snippet(resp.Body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if _, err := parseWords(data); err != nil {
		return fmt.Errorf("download %q: %w", id, err)
	}

	tmp := c.languagePath(id) + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, c.languagePath(id))
}

func snippet(r io.Reader) string {
	body, _ := io.ReadAll(io.LimitReader(r, 512))
	return strings.TrimSpace(string(body))
}

type languageFile struct {
	Words []string `json:"words"`
}

func parseWords(data []byte) ([]string, error) {
	var lf languageFile
	if err := json.Unmarshal(data, &lf); err != nil {
		return nil, err
	}
	if len(lf.Words) == 0 {
		return nil, fmt.Errorf("empty word list")
	}
	return lf.Words, nil
}
