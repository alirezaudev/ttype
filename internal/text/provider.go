package text

import (
	"errors"
	"math/rand/v2"
	"strings"

	"github.com/alirezaudev/ttype/assets"
)

type Provider struct {
	words []string
}

func NewProvider() (*Provider, error) {
	words, err := assets.LoadWords()
	if err != nil {
		return nil, err
	}

	return &Provider{
		words: words,
	}, nil
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
