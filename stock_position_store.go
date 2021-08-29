package virtual_security

import (
	"sort"
	"sync"
)

var (
	stockPositionStoreSingleton      iStockPositionStore
	stockPositionStoreSingletonMutex sync.Mutex
)

func getStockPositionStore() iStockPositionStore {
	stockPositionStoreSingletonMutex.Lock()
	defer stockPositionStoreSingletonMutex.Unlock()

	if stockPositionStoreSingleton == nil {
		stockPositionStoreSingleton = &stockPositionStore{
			store: map[string]*stockPosition{},
		}
	}
	return stockPositionStoreSingleton
}

// iStockPositionStore - 現物株式ポジションストアのインターフェース
type iStockPositionStore interface {
	getAll() []*stockPosition
	getByCode(code string) (*stockPosition, error)
	getBySymbolCode(symbolCode string) ([]*stockPosition, error)
	save(stockPosition *stockPosition)
	removeByCode(code string)
}

// stockPositionStore - 現物株式ポジションのストア
type stockPositionStore struct {
	store map[string]*stockPosition
	mtx   sync.Mutex
}

// getAll - ストアのすべてのポジションをコード順に並べて返す
func (s *stockPositionStore) getAll() []*stockPosition {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	positions := make([]*stockPosition, len(s.store))
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

// getByCode - コードを指定してデータを取得する
func (s *stockPositionStore) getByCode(code string) (*stockPosition, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if position, ok := s.store[code]; ok {
		return position, nil
	} else {
		return nil, NoDataError
	}
}

// getBySymbolCode - 銘柄コードを指定してデータを取得する
func (s *stockPositionStore) getBySymbolCode(symbolCode string) ([]*stockPosition, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	positions := make([]*stockPosition, 0)
	for _, position := range s.store {
		if symbolCode == position.SymbolCode {
			positions = append(positions, position)
		}
	}
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].Code < positions[j].Code
	})
	return positions, nil
}

// save - ポジションをストアに追加する
func (s *stockPositionStore) save(stockPosition *stockPosition) {
	if stockPosition == nil {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[stockPosition.Code] = stockPosition
}

// removeByCode - コードを指定して削除する
func (s *stockPositionStore) removeByCode(code string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.store, code)
}
