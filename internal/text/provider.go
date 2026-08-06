package text

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/assets"
)

const (
	languagesRepo   = "alirezaudev/monkeytype"
	languagesBranch = "master"
	languagesDir    = "frontend/static/languages"
	manifestFile    = "_manifest.json"
	manifestMaxAge  = 7 * 24 * time.Hour
)

var client = &http.Client{Timeout: 15 * time.Second}

type Provider struct {
	builtin []string
	words   []string
	dir     string
}

func NewProvider(dataDir string) (*Provider, error) {
	words, err := assets.LoadWords()
	if err != nil {
		return nil, err
	}

	return &Provider{
		builtin: words,
		words:   words,
		dir:     filepath.Join(dataDir, "languages"),
	}, nil
}

func DisplayName(id string) string {
	if id == "" {
		return "english (built-in)"
	}
	return strings.ReplaceAll(id, "_", " ")
}

func (p *Provider) Languages() ([]string, error) {
	cached := p.readManifest()
	if len(cached) > 0 && !p.manifestStale() {
		return cached, nil
	}

	ids, err := p.fetchLanguages()
	if err != nil {
		if len(cached) > 0 {
			return cached, nil
		}
		return nil, err
	}

	p.writeManifest(ids)
	return ids, nil
}

func (p *Provider) manifestPath() string {
	return filepath.Join(p.dir, manifestFile)
}

func (p *Provider) manifestStale() bool {
	info, err := os.Stat(p.manifestPath())
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > manifestMaxAge
}

func (p *Provider) readManifest() []string {
	data, err := os.ReadFile(p.manifestPath())
	if err != nil {
		return nil
	}

	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil
	}
	return ids
}

func (p *Provider) writeManifest(ids []string) {
	if err := os.MkdirAll(p.dir, 0o700); err != nil {
		return
	}

	data, err := json.Marshal(ids)
	if err != nil {
		return
	}
	_ = os.WriteFile(p.manifestPath(), data, 0o600)
}

func (p *Provider) fetchLanguages() ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s?ref=%s&per_page=1000", languagesRepo, languagesDir, languagesBranch)
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
		return nil, fmt.Errorf("list languages failed: %s", resp.Status)
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

func (p *Provider) UseLanguage(id string) error {
	if id == "" {
		p.words = p.builtin
		return nil
	}

	path := filepath.Join(p.dir, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		data, err = p.download(id)
		if err != nil {
			return err
		}
	}

	words, err := parseWords(data)
	if err != nil {
		return fmt.Errorf("language %q: %w", id, err)
	}

	p.words = words
	return nil
}

func (p *Provider) download(id string) ([]byte, error) {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s.json", languagesRepo, languagesBranch, languagesDir, id)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "ttype")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %q failed: %s", id, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if _, err := parseWords(data); err != nil {
		return nil, fmt.Errorf("download %q: %w", id, err)
	}

	if err := os.MkdirAll(p.dir, 0o700); err != nil {
		return nil, err
	}
	tmp := filepath.Join(p.dir, id+".json.tmp")
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return nil, err
	}
	if err := os.Rename(tmp, filepath.Join(p.dir, id+".json")); err != nil {
		return nil, err
	}
	return data, nil
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

func (p *Provider) Generate(wordLimit int) (string, error) {
	if wordLimit > len(p.words) || wordLimit < 0 {
		return "", errors.New("words limit out of range")
	}

	sampled := make([]string, wordLimit)
	for i := range sampled {
		sampled[i] = p.words[rand.IntN(len(p.words))]
	}
	return strings.Join(sampled, " "), nil
}
