package domain

type TestKind string

const (
	TestKindTimed TestKind = "timed"
	TestKindWords TestKind = "words"
)

type TestConfig struct {
	Kind        TestKind `json:"test_kind"`
	Duration    Duration `json:"duration"`
	WordCount   int      `json:"word_count,omitempty"`
	TextMode    TextMode `json:"text_mode"`
	Language    string   `json:"language,omitempty"`
	Theme       string   `json:"theme"`
	Width       int      `json:"width,omitempty"`
	Punctuation bool     `json:"punctuation,omitempty"`
	Numbers     bool     `json:"numbers,omitempty"`
	Blind       bool     `json:"blind,omitempty"`
	Zen         bool     `json:"zen,omitempty"`
	MinWPM      int      `json:"min_wpm,omitempty"`
	Seed        int64    `json:"seed,omitempty"`
}

type GenerateOptions struct {
	Mode        TextMode
	WordLimit   int
	Language    string
	Punctuation bool
	Numbers     bool
	Blind       bool
	Zen         bool
	MinWPM      int
	Seed        int64
}

func (c TestConfig) IsWordsMode() bool {
	return c.Kind == TestKindWords && c.WordCount > 0
}
