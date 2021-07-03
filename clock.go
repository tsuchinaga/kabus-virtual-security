package virtual_security

import (
	"time"
)

type Clock interface {
	Now() time.Time
	GetStockSession(now time.Time) Session
	GetSession(exchangeType ExchangeType, now time.Time) Session
	GetBusinessDay(exchangeType ExchangeType, now time.Time) time.Time
}

func newClock() Clock {
	return &clock{}
}

type clock struct{}

func (c *clock) Now() time.Time {
	return time.Now()
}

// TODO 引数にExchangeTypeをもらって振り分けれるようにする

func (c *clock) GetStockSession(now time.Time) Session {
	switch {
	case contractableMorningSessionTime.between(now):
		return SessionMorning
	case contractableAfternoonSessionTime.between(now):
		return SessionAfternoon
	}
	return SessionUnspecified
}

func (c *clock) GetSession(exchangeType ExchangeType, now time.Time) Session {
	if now.IsZero() {
		return SessionUnspecified
	}

	switch exchangeType {
	case ExchangeTypeStock, ExchangeTypeMargin:
		return c.GetStockSession(now)
	case ExchangeTypeFuture:
		// TODO 先物の場合の振り分け
	}
	return SessionUnspecified
}

func (c *clock) GetBusinessDay(exchangeType ExchangeType, now time.Time) time.Time {
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
