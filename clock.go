package virtual_security

import (
	"time"
)

type Clock interface {
	Now() time.Time
}

func newClock() Clock {
	return &clock{}
}

type clock struct{}

func (c *clock) Now() time.Time {
	return time.Now()
}
