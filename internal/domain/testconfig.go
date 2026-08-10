package domain

type TestKind string

const (
	TestKindTimed TestKind = "timed"
	TestKindWords TestKind = "words"
)

type TestConfig struct {
	Kind        TestKind
	Duration    Duration
	WordCount   int
	Width       int
	Theme       string
	Language    string
	TextMode    TextMode
	Punctuation bool
	Numbers     bool
	Seed        int64
}

type GenerateOptions struct {
	Mode        TextMode
	WordLimit   int
	Language    string
	Punctuation bool
	Numbers     bool
	Seed        int64
}

func (c TestConfig) IsWordsMode() bool {
	return c.Kind == TestKindWords && c.WordCount > 0
}
