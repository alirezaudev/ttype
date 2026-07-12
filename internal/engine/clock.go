package engine

import "time"

type Clock interface {
	Now() time.Time
}

type RealClock struct{}

func (RealClock) Now() time.Time {
	return time.Now()
}

type FakeClock struct {
	current time.Time
}

func NewFakeClock(start time.Time) *FakeClock {
	return &FakeClock{current: start}
}

func (c *FakeClock) Now() time.Time {
	return c.current
}

func (c *FakeClock) Advance(d time.Duration) {
	c.current = c.current.Add(d)
}
