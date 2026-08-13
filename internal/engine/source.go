package engine

import "github.com/alirezaudev/ttype/internal/domain"

// StaticText feeds a replay the exact text its run was typed on.
type StaticText string

func (s StaticText) Generate(domain.GenerateOptions) (string, error) {
	return string(s), nil
}
