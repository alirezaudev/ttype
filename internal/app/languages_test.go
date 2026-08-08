package app

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfirm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		answer string
		want   bool
	}{
		{answer: "y\n", want: true},
		{answer: "YES\n", want: true},
		{answer: "\n", want: false},
		{answer: "nope\n", want: false},
	}

	for _, test := range tests {
		var out bytes.Buffer

		got, err := confirm(&out, strings.NewReader(test.answer), "go? ")
		if err != nil {
			t.Fatalf("confirm(%q): %v", test.answer, err)
		}
		if got != test.want {
			t.Errorf("confirm(%q) = %v, want %v", test.answer, got, test.want)
		}
		if out.String() != "go? " {
			t.Errorf("prompt = %q", out.String())
		}
	}
}
