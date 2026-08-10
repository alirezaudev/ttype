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
	items map[domain.TextMode][]string
	cache *langcache.Cache
}

func NewProvider(dataDir string) (*Provider, error) {
	loaders := map[domain.TextMode]func() ([]string, error){
		domain.TextModeWords:     assets.LoadWords,
		domain.TextModeSentences: assets.LoadSentences,
		domain.TextModeSQL:       assets.LoadSQL,
		domain.TextModeGo:        assets.LoadGo,
		domain.TextModeBackend:   assets.LoadBackend,
		domain.TextModePython:    assets.LoadPython,
		domain.TextModeShell:     assets.LoadShell,
		domain.TextModeRegex:     assets.LoadRegex,
	}

	items := make(map[domain.TextMode][]string, len(loaders))
	for mode, load := range loaders {
		loaded, err := load()
		if err != nil {
			return nil, fmt.Errorf("load %s: %w", mode, err)
		}
		items[mode] = loaded
	}

	return &Provider{items: items, cache: langcache.New(dataDir)}, nil
}

func (p *Provider) LanguageCache() *langcache.Cache {
	return p.cache
}

func (p *Provider) Generate(opts domain.GenerateOptions) (string, error) {
	mode := opts.Mode
	if mode == "" {
		mode = domain.TextModeWords
	}
	wordsMode := mode == domain.TextModeWords

	items, err := p.itemsForMode(mode)
	if err != nil {
		return "", err
	}

	if opts.Language != "" && wordsMode {
		remote, err := p.cache.Words(opts.Language)
		if err != nil {
			return "", fmt.Errorf("language %q: %w", opts.Language, err)
		}
		items = remote
	}

	if opts.WordLimit < 0 {
		return "", errors.New("words limit out of range")
	}

	rng := newRNG(opts.Seed)

	sampled := sample(items, opts.WordLimit, rng)

	if opts.Numbers && wordsMode {
		applyNumbers(sampled, rng)
	}
	return joinWords(sampled, opts.Punctuation && wordsMode, rng), nil
}

func newRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewPCG(uint64(seed), uint64(seed>>32)))
}

func (p *Provider) itemsForMode(mode domain.TextMode) ([]string, error) {
	items, ok := p.items[mode]
	if !ok {
		return nil, fmt.Errorf("unsupported mode %q", mode)
	}
	return items, nil
}

// sample draws limit words, or enough words to fill a timed test's buffer.
func sample(words []string, limit int, rng *rand.Rand) []string {
	if limit > 0 {
		out := make([]string, limit)
		for i := range out {
			out[i] = words[rng.IntN(len(words))]
		}
		return out
	}

	var out []string
	for length := 0; length < minTargetRunes; {
		word := words[rng.IntN(len(words))]
		if len(out) > 0 {
			length++
		}
		length += len(word)
		out = append(out, word)
	}
	return out
}

func applyNumbers(words []string, rng *rand.Rand) {
	for i := range words {
		if rng.Float64() < numberChance {
			words[i] = strconv.Itoa(rng.IntN(9999) + 1)
		}
	}
}

func joinWords(words []string, punct bool, rng *rand.Rand) string {
	if len(words) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(words[0])
	for _, word := range words[1:] {
		sep := " "
		if punct && rng.Float64() < punctChance {
			sep = ". "
			if rng.Float64() < commaShare {
				sep = ", "
			}
		}
		b.WriteString(sep)
		b.WriteString(word)
	}
	return b.String()
}
