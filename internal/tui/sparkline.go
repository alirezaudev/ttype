package tui

import (
	"math"
	"strings"
)

const sparklineChars = " .:-=+#"

func renderSparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	maxV := 1.0
	for _, v := range values {
		maxV = math.Max(maxV, v)
	}

	var b strings.Builder
	for _, v := range values {
		idx := int(v / maxV * float64(len(sparklineChars)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparklineChars) {
			idx = len(sparklineChars) - 1
		}
		b.WriteByte(sparklineChars[idx])
	}
	return b.String()
}

func downsampleSeries(vals []float64, maxN int) []float64 {
	if maxN < 1 || len(vals) <= maxN {
		return vals
	}
	if maxN == 1 {
		return vals[:1]
	}

	out := make([]float64, maxN)
	for i := range out {
		t := float64(i) * float64(len(vals)-1) / float64(maxN-1)
		out[i] = vals[int(math.Round(t))]
	}
	return out
}
