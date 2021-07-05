package virtual_security

import (
	"time"
)

type iClock interface {
	now() time.Time
	getStockSession(now time.Time) Session
	getSession(exchangeType ExchangeType, now time.Time) Session
	getBusinessDay(exchangeType ExchangeType, now time.Time) time.Time
}

func newClock() iClock {
	return &clock{}
}

type clock struct{}

func (c *clock) now() time.Time {
	return time.Now()
}

func (c *clock) getStockSession(now time.Time) Session {
	switch {
	case contractableMorningSessionTime.between(now):
		return SessionMorning
	case contractableAfternoonSessionTime.between(now):
		return SessionAfternoon
	}
	return SessionUnspecified
}

func (c *clock) getSession(exchangeType ExchangeType, now time.Time) Session {
	if now.IsZero() {
		return SessionUnspecified
	}

	switch exchangeType {
	case ExchangeTypeStock, ExchangeTypeMargin:
		return c.getStockSession(now)
	case ExchangeTypeFuture:
		// TODO 先物の場合の振り分け
	}
	return SessionUnspecified
}

func (c *clock) getBusinessDay(exchangeType ExchangeType, now time.Time) time.Time {
	if now.IsZero() {
		return now
	}

	switch exchangeType {
	case ExchangeTypeStock, ExchangeTypeMargin:
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case ExchangeTypeFuture:
		// TODO 先物の場合の振り分け
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}
