package virtual_security

import (
	"sort"
	"sync"
)

var (
	stockPositionStoreSingleton      StockPositionStore
	stockPositionStoreSingletonMutex sync.Mutex
)

func GetStockPositionStore() StockPositionStore {
	stockPositionStoreSingletonMutex.Lock()
	defer stockPositionStoreSingletonMutex.Unlock()

	if stockPositionStoreSingleton == nil {
		stockPositionStoreSingleton = &stockPositionStore{
			store: map[string]*StockPosition{},
		}
	}
	return stockPositionStoreSingleton
}

// StockPositionStore - 現物株式ポジションストアのインターフェース
type StockPositionStore interface {
	GetAll() []*StockPosition
	GetByCode(code string) (*StockPosition, error)
	Add(stockPosition *StockPosition)
	RemoveByCode(code string)
}

// stockPositionStore - 現物株式ポジションのストア
type stockPositionStore struct {
	store map[string]*StockPosition
	mtx   sync.Mutex
}

// GetAll - ストアのすべてのポジションをコード順に並べて返す
func (s *stockPositionStore) GetAll() []*StockPosition {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	positions := make([]*StockPosition, len(s.store))
	var i int
	for _, position := range s.store {
		positions[i] = position
		i++
	}
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].Code < positions[j].Code
	})
	return positions
}

// GetByCode - コードを指定してデータを取得する
func (s *stockPositionStore) GetByCode(code string) (*StockPosition, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if position, ok := s.store[code]; ok {
		return position, nil
	} else {
		return nil, NoDataError
	}
}

// Add - ポジションをストアに追加する
func (s *stockPositionStore) Add(stockPosition *StockPosition) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[stockPosition.Code] = stockPosition
}

// RemoveByCode - コードを指定して削除する
func (s *stockPositionStore) RemoveByCode(code string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.store, code)
}
