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
	"time"
)

const (
	defaultRepo    = "alirezaudev/monkeytype"
	defaultBranch  = "master"
	languagesPath  = "frontend/static/languages"
	manifestFile   = "_manifest.json"
	manifestMaxAge = 7 * 24 * time.Hour
	fetchTimeout   = 15 * time.Second
)

var client = &http.Client{Timeout: fetchTimeout}

type Cache struct {
	dir string
}

func New(dataDir string) *Cache {
	return &Cache{dir: filepath.Join(dataDir, "languages")}
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

func (c *Cache) manifestStale() bool {
	info, err := os.Stat(c.manifestPath())
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > manifestMaxAge
}

func (c *Cache) loadManifest() ([]string, error) {
	data, err := os.ReadFile(c.manifestPath())
	if err != nil {
		return nil, err
	}

	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

func (c *Cache) saveManifest(ids []string) error {
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		return err
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return err
	}
	return os.WriteFile(c.manifestPath(), data, 0o600)
}

func (c *Cache) fetchManifest() ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s&per_page=1000", defaultRepo, languagesPath, defaultBranch)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "ttype")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list languages failed: %s: %s", resp.Status, snippet(resp.Body))
	}

	var items []struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}

	var ids []string
	for _, item := range items {
		if !strings.HasSuffix(item.Name, ".json") || item.Name == manifestFile {
			continue
		}
		ids = append(ids, strings.TrimSuffix(item.Name, ".json"))
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("list languages returned no entries")
	}

	sort.Strings(ids)
	return ids, nil
}

func (c *Cache) download(id string) error {
	if err := os.MkdirAll(c.dir, 0o700); err != nil {
		return err
	}

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s.json", defaultRepo, defaultBranch, languagesPath, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "ttype")

	resp, err := client.Do(req)
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
