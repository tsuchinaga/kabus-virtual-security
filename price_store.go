package virtual_security

import (
	"sync"
	"time"
)

var (
	priceStoreSingleton      PriceStore
	priceStoreSingletonMutex sync.Mutex
)

// getPriceStore - 価格ストアの取得
func getPriceStore(clock Clock) PriceStore {
	priceStoreSingletonMutex.Lock()
	defer priceStoreSingletonMutex.Unlock()

	if priceStoreSingleton == nil {
		store := &priceStore{
			store: map[string]*symbolPrice{},
			clock: clock,
		}
		store.setCalculatedExpireTime(clock.Now())
		priceStoreSingleton = store
	}
	return priceStoreSingleton
}

// PriceStore - 価格ストアのインターフェース
type PriceStore interface {
	GetBySymbolCode(symbolCode string) (*symbolPrice, error)
	Set(price *symbolPrice) error
}

// priceStore - 価格ストア
type priceStore struct {
	store      map[string]*symbolPrice
	clock      Clock
	expireTime time.Time
	mtx        sync.Mutex
}

func (s *priceStore) isExpired(now time.Time) bool {
	return !s.expireTime.After(now)
}

func (s *priceStore) setCalculatedExpireTime(now time.Time) {
	s.expireTime = time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.Local)
	if now.Hour() >= 8 {
		s.expireTime = s.expireTime.AddDate(0, 0, 1)
	}
}

// GetBySymbolCode - ストアから指定した銘柄コードの価格を取り出す
func (s *priceStore) GetBySymbolCode(symbolCode string) (*symbolPrice, error) {
	if price, ok := s.store[symbolCode]; ok {
		if s.isExpired(s.clock.Now()) {
			return nil, ExpiredDataError
		}
		return price, nil
	} else {
		return nil, NoDataError
	}
}

// Set - ストアに銘柄の価格情報を登録する
func (s *priceStore) Set(price *symbolPrice) error {
	if price == nil {
		return NilArgumentError
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	// 現値、売り気配、買い気配のいずれの時間もゼロ値なら無視
	if price.maxTime().IsZero() {
		return nil
	}

	// storeの有効期限が切れていたら初期化
	//   有効期限は次の8時まで
	now := s.clock.Now()
	if s.isExpired(now) {
		s.store = map[string]*symbolPrice{}
		s.setCalculatedExpireTime(now)
	}

	// ストアにセット
	s.store[price.SymbolCode] = price
	return nil
}
