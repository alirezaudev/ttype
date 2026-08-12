package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var brailleBits = [4][2]uint8{
	{0x01, 0x08},
	{0x02, 0x10},
	{0x04, 0x20},
	{0x40, 0x80},
}

type chartSeries uint8

const (
	seriesNone chartSeries = iota
	seriesRaw
	seriesNet
	seriesError
)

const (
	minBrailleChartWidth  = 40
	minBrailleChartHeight = 5

	minChartHeight = 2
)

func renderResultChart(history, rawHistory []float64, errorHistory []int, theme Theme, width, height int) string {
	if len(history) < 2 || height < minChartHeight || width < 8 {
		return ""
	}

	peak := 0.0
	for _, v := range history {
		peak = math.Max(peak, v)
	}
	header := theme.HUD.Render("wpm over time") + theme.Help.Render(fmt.Sprintf(" · peak %.0f", peak))

	if width < minBrailleChartWidth || height < minBrailleChartHeight || !unicodeCapable() {
		spark := renderSparkline(downsampleSeries(history, width))
		return header + "\n" + theme.HUDValue.Render(spark)
	}

	return header + "\n" + renderBrailleChart(history, rawHistory, errorHistory, theme, width, height-1)
}

func renderBrailleChart(history, rawHistory []float64, errorHistory []int, theme Theme, width, height int) string {
	maxV := 0.0
	for _, v := range history {
		maxV = math.Max(maxV, v)
	}
	for _, v := range rawHistory {
		maxV = math.Max(maxV, v)
	}
	scale := math.Max(math.Ceil(maxV/10)*10, 10)

	topLabel := fmt.Sprintf("%.0f", scale)
	gutter := len(topLabel)
	plot := newBraillePlot(width-gutter-1, height-1, scale)

	if len(rawHistory) > 1 {
		plot.band(rawHistory, seriesRaw)
	}
	plot.line(history, seriesNet)
	plot.markErrors(history, rawHistory, errorHistory)

	var b strings.Builder
	for r := 0; r < plot.rows; r++ {
		label := ""
		switch r {
		case 0:
			label = topLabel
		case plot.rows - 1:
			label = "0"
		case plot.rows / 2:
			label = fmt.Sprintf("%.0f", scale/2)
		}
		if label == "" {
			b.WriteString(theme.Help.Render(strings.Repeat(" ", gutter) + "│"))
		} else {
			b.WriteString(theme.Help.Render(fmt.Sprintf("%*s┤", gutter, label)))
		}
		b.WriteString(plot.renderRow(r, theme))
		b.WriteByte('\n')
	}
	b.WriteString(theme.Help.Render(chartXAxis(len(history)-1, gutter, plot.cols)))
	return b.String()
}

type braillePlot struct {
	dots   [][]uint8
	series [][]chartSeries
	cols   int
	rows   int
	scale  float64
}

func newBraillePlot(cols, rows int, scale float64) *braillePlot {
	p := &braillePlot{
		dots:   make([][]uint8, rows),
		series: make([][]chartSeries, rows),
		cols:   cols,
		rows:   rows,
		scale:  scale,
	}
	for r := range p.dots {
		p.dots[r] = make([]uint8, cols)
		p.series[r] = make([]chartSeries, cols)
	}
	return p
}

func (p *braillePlot) pixelWidth() int  { return p.cols * 2 }
func (p *braillePlot) pixelHeight() int { return p.rows * 4 }

func (p *braillePlot) y(v float64) int {
	if v < 0 {
		v = 0
	}
	return min(int(math.Round(v/p.scale*float64(p.pixelHeight()-1))), p.pixelHeight()-1)
}

func (p *braillePlot) set(px, py int, s chartSeries) {
	fromTop := p.pixelHeight() - 1 - py
	r, c := fromTop/4, px/2
	p.dots[r][c] |= brailleBits[fromTop%4][px%2]
	if s > p.series[r][c] {
		p.series[r][c] = s
	}
}

func (p *braillePlot) line(vals []float64, s chartSeries) {
	prev := -1
	for px := 0; px < p.pixelWidth(); px++ {
		y := p.y(sampleAt(vals, px, p.pixelWidth()))
		lo, hi := y, y
		if prev >= 0 {
			lo, hi = min(prev, y), max(prev, y)
		}
		for py := lo; py <= hi; py++ {
			p.set(px, py, s)
		}
		prev = y
	}
}

func (p *braillePlot) band(vals []float64, s chartSeries) {
	for px := 0; px < p.pixelWidth(); px++ {
		top := p.y(sampleAt(vals, px, p.pixelWidth()))
		for py := 0; py <= top; py++ {
			p.set(px, py, s)
		}
	}
}

func (p *braillePlot) markErrors(history, rawHistory []float64, errorHistory []int) {
	if len(errorHistory) < 2 {
		return
	}

	for i, count := range errorHistory {
		if count <= 0 {
			continue
		}

		var v float64
		switch {
		case i < len(rawHistory):
			v = rawHistory[i]
		case i < len(history):
			v = history[i]
		default:
			continue
		}

		px := i * (p.pixelWidth() - 1) / (len(errorHistory) - 1)
		row := (p.pixelHeight() - 1 - p.y(v)) / 4
		p.series[row][px/2] = seriesError
	}
}

func (p *braillePlot) renderRow(r int, theme Theme) string {
	styleFor := func(s chartSeries) (lipgloss.Style, bool) {
		switch s {
		case seriesNet:
			return theme.HUDWPM, true
		case seriesRaw:
			return theme.Pending, true
		case seriesError:
			return theme.HUDErr, true
		default:
			return lipgloss.Style{}, false
		}
	}

	var b, run strings.Builder
	current := seriesNone
	flush := func() {
		if run.Len() == 0 {
			return
		}
		if style, ok := styleFor(current); ok {
			b.WriteString(style.Render(run.String()))
		} else {
			b.WriteString(run.String())
		}
		run.Reset()
	}

	for c, dots := range p.dots[r] {
		s := p.series[r][c]
		if s != current {
			flush()
			current = s
		}
		switch {
		case s == seriesError:
			run.WriteRune('×')
		case dots == 0:
			run.WriteByte(' ')
		default:
			run.WriteRune(rune(0x2800 + int(dots)))
		}
	}
	flush()
	return b.String()
}

func sampleAt(vals []float64, px, pw int) float64 {
	t := float64(px) * float64(len(vals)-1) / float64(pw-1)
	i := int(t)
	if i >= len(vals)-1 {
		return vals[len(vals)-1]
	}
	f := t - float64(i)
	return vals[i]*(1-f) + vals[i+1]*f
}

func chartXAxis(seconds, gutter, cols int) string {
	left, right := "0s", fmt.Sprintf("%ds", seconds)
	dashes := max(cols-len(left)-len(right), 0)
	return strings.Repeat(" ", gutter) + "└" + left + strings.Repeat("─", dashes) + right
}
