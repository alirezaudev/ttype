package text

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/alirezaudev/ttype/assets"
	"github.com/alirezaudev/ttype/internal/domain"
	"github.com/alirezaudev/ttype/internal/text/langcache"
)

const minTargetRunes = 4096

const (
	numberChance = 0.1
	punctChance  = 0.14
	commaShare   = 0.55
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

	sampled := sample(words, opts.WordLimit)
	if opts.Numbers {
		applyNumbers(sampled)
	}
	return joinWords(sampled, opts.Punctuation), nil
}

// sample draws limit words, or enough words to fill a timed test's buffer.
func sample(words []string, limit int) []string {
	if limit > 0 {
		out := make([]string, limit)
		for i := range out {
			out[i] = words[rand.IntN(len(words))]
		}
		return out
	}

	var out []string
	for length := 0; length < minTargetRunes; {
		word := words[rand.IntN(len(words))]
		if len(out) > 0 {
			length++
		}
		length += len(word)
		out = append(out, word)
	}
	return out
}

func applyNumbers(words []string) {
	for i := range words {
		if rand.Float64() < numberChance {
			words[i] = strconv.Itoa(rand.IntN(9999) + 1)
		}
	}
}

func joinWords(words []string, punct bool) string {
	if len(words) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(words[0])
	for _, word := range words[1:] {
		sep := " "
		if punct && rand.Float64() < punctChance {
			sep = ". "
			if rand.Float64() < commaShare {
				sep = ", "
			}
		}
		b.WriteString(sep)
		b.WriteString(word)
	}
	return b.String()
}
