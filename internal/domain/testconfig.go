package domain

type TestKind string

const (
	TestKindTimed TestKind = "timed"
	TestKindWords TestKind = "words"
)

type TestConfig struct {
	Kind      TestKind
	Duration  Duration
	WordCount int
	Width     int
	Theme     string
	Language  string
}

type GenerateOptions struct {
	WordLimit int
	Language  string
}

func (c TestConfig) GenerateOptions() GenerateOptions {
	return GenerateOptions{
		WordLimit: c.WordCount,
		Language:  c.Language,
	}
}

func (c TestConfig) IsWordsMode() bool {
	return c.Kind == TestKindWords && c.WordCount > 0
}
