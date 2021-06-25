package virtual_security

import (
	"time"
)

type Clock interface {
	Now() time.Time
	GetStockSession(now time.Time) Session
}

func newClock() Clock {
	return &clock{}
}

type clock struct{}

func (c *clock) Now() time.Time {
	return time.Now()
}

func (c *clock) GetStockSession(now time.Time) Session {
	switch {
	case stockMorningSessionTime.between(now):
		return SessionMorning
	case stockAfternoonSessionTime.between(now):
		return SessionAfternoon
	}
	return SessionUnspecified
}
