package tui

import (
	"strings"
	"testing"

	"github.com/alirezaudev/ttype/internal/domain"
)

func benchModel(b *testing.B) TestModel {
	b.Helper()

	target := strings.TrimSpace(strings.Repeat("the quick brown fox jumps over the lazy dog ", 40))
	cfg := domain.TestConfig{Kind: domain.TestKindTimed, Duration: domain.Duration60}
	m := newBenchTestModel(b, target, cfg)
	for _, r := range target[:400] {
		m.session.InputRune(r)
	}
	return m
}

func BenchmarkRenderWords(b *testing.B) {
	m := benchModel(b)
	b.ReportAllocs()
	b.ResetTimer()

	total := 0
	for i := 0; i < b.N; i++ {
		total += len(m.renderWords())
	}
	b.ReportMetric(float64(total)/float64(b.N), "bytes/frame")
}
