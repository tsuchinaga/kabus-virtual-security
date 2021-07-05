package virtual_security

import (
	"context"
	"sync"
	"time"

	jbd "gitlab.com/tsuchinaga/jpx-business-day"
)

var (
	businessDayStoreSingleton    iBusinessDayStore
	businessDayStoreSingletonMtx sync.Mutex
)

func getBusinessDayStore(clock iClock) iBusinessDayStore {
	businessDayStoreSingletonMtx.Lock()
	defer businessDayStoreSingletonMtx.Unlock()

	if businessDayStoreSingleton == nil {
		businessDayStoreSingleton = &businessDayStore{
			clock:       clock,
			businessDay: jbd.NewBusinessDay(),
		}
	}

	return businessDayStoreSingleton
}

type iBusinessDayStore interface {
	isBusinessDay(target time.Time) (bool, error)
}

type businessDayStore struct {
	clock       iClock
	businessDay jbd.BusinessDay
	refreshedAt time.Time
}

func (s *businessDayStore) isBusinessDay(target time.Time) (bool, error) {
	now := s.clock.now()
	nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if s.refreshedAt.IsZero() || s.refreshedAt.Before(nowDate) || s.businessDay.LastUpdateDate().IsZero() {
		if err := s.businessDay.Refresh(context.Background()); err != nil {
			return false, err
		}
	}

	return s.businessDay.IsBusinessDay(target), nil
}
