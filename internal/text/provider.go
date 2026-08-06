package text

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"

	"github.com/alirezaudev/ttype/assets"
	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/text/langcache"
)

type Provider struct {
	words []string
	cache *langcache.Cache
}

func NewProvider(dataDir string) (*Provider, error) {
	words, err := assets.LoadWords()
	if err != nil {
		return nil, err
	}

	return &Provider{
		words: words,
		cache: langcache.New(dataDir),
	}, nil
}

func (p *Provider) LanguageCache() *langcache.Cache {
	return p.cache
}

func (p *Provider) Generate(opts domain.GenerateOptions) (string, error) {
	words := p.words
	if opts.Language != "" {
		remote, err := p.cache.Words(opts.Language)
		if err != nil {
			return "", fmt.Errorf("language %q: %w", opts.Language, err)
		}
		words = remote
	}

	if opts.WordLimit > len(words) || opts.WordLimit < 0 {
		return "", errors.New("words limit out of range")
	}

	sampled := make([]string, opts.WordLimit)
	for i := range sampled {
		sampled[i] = words[rand.IntN(len(words))]
	}
	return strings.Join(sampled, " "), nil
}
