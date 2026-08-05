package text

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/alirezaudev/ttype/assets"
)

const (
	languagesRepo   = "alirezaudev/monkeytype"
	languagesBranch = "master"
	languagesDir    = "frontend/static/languages"
)

var client = &http.Client{Timeout: 15 * time.Second}

type Provider struct {
	builtin []string
	words   []string
}

func NewProvider() (*Provider, error) {
	words, err := assets.LoadWords()
	if err != nil {
		return nil, err
	}

	return &Provider{
		builtin: words,
		words:   words,
	}, nil
}

func DisplayName(id string) string {
	if id == "" {
		return "english (built-in)"
	}
	return strings.ReplaceAll(id, "_", " ")
}

func (p *Provider) Languages() ([]string, error) {
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
		if !strings.HasSuffix(item.Name, ".json") {
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

	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s.json", languagesRepo, languagesBranch, languagesDir, id)
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
		return fmt.Errorf("download %q failed: %s", id, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var file struct {
		Words []string `json:"words"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	if len(file.Words) == 0 {
		return fmt.Errorf("language %q has no words", id)
	}

	p.words = file.Words
	return nil
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
