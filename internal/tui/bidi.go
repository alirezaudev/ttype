package tui

import (
	"os"
	"unicode"
)

// Most terminals draw characters in the order they arrive, so Persian, Arabic
// and Hebrew come out backwards unless the text is sent in display order.
// Terminals built on VTE, Konsole and mlterm reorder it themselves, and doing
// it twice would reverse it again.
var reorderRTL = shouldReorderRTL(os.Getenv)

func shouldReorderRTL(getenv func(string) string) bool {
	// TTYPE_BIDI says who reorders: ttype, or the terminal.
	switch getenv("TTYPE_BIDI") {
	case "ttype":
		return true
	case "terminal":
		return false
	}
	for _, name := range []string{"VTE_VERSION", "KONSOLE_VERSION", "MLTERM"} {
		if getenv(name) != "" {
			return false
		}
	}
	return true
}

func isRTLRune(r rune) bool {
	switch {
	case r >= 0x0590 && r <= 0x08FF: // Hebrew, Arabic, Syriac, Thaana, NKo
		return true
	case r >= 0xFB1D && r <= 0xFDFF: // Hebrew and Arabic presentation forms
		return true
	case r >= 0xFE70 && r <= 0xFEFF:
		return true
	}
	return false
}

// rightToLeft reports whether the text reads right to left, going by its
// first letter.
func rightToLeft(text []rune) bool {
	for _, r := range text {
		if isRTLRune(r) {
			return true
		}
		if unicode.IsLetter(r) {
			return false
		}
	}
	return false
}

// visualOrder lists the indices of a right-to-left line from left to right:
// everything reversed, except runs of digits and Latin letters, which still
// read left to right.
func visualOrder(text []rune, start, end int) []int {
	order := make([]int, 0, end-start)
	for i := end - 1; i >= start; i-- {
		order = append(order, i)
	}
	for i := 0; i < len(order); {
		if !leftToRightRune(text[order[i]]) {
			i++
			continue
		}
		j := i
		for j < len(order) && leftToRightRune(text[order[j]]) {
			j++
		}
		for a, b := i, j-1; a < b; a, b = a+1, b-1 {
			order[a], order[b] = order[b], order[a]
		}
		i = j
	}
	return order
}

func leftToRightRune(r rune) bool {
	return unicode.IsDigit(r) || (unicode.IsLetter(r) && !isRTLRune(r))
}

// Brackets face the other way in right-to-left text.
func mirrorRune(r rune) rune {
	switch r {
	case '(':
		return ')'
	case ')':
		return '('
	case '[':
		return ']'
	case ']':
		return '['
	case '{':
		return '}'
	case '}':
		return '{'
	case '<':
		return '>'
	case '>':
		return '<'
	case '«':
		return '»'
	case '»':
		return '«'
	}
	return r
}
