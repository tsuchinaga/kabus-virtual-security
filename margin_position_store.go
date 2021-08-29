package virtual_security

import (
	"sort"
	"sync"
)

var (
	marginPositionStoreSingleton      iMarginPositionStore
	marginPositionStoreSingletonMutex sync.Mutex
)

func getMarginPositionStore() iMarginPositionStore {
	marginPositionStoreSingletonMutex.Lock()
	defer marginPositionStoreSingletonMutex.Unlock()

	if marginPositionStoreSingleton == nil {
		marginPositionStoreSingleton = &marginPositionStore{
			store: map[string]*marginPosition{},
		}
	}
	return marginPositionStoreSingleton
}

// iMarginPositionStore - 信用株式ポジションストアのインターフェース
type iMarginPositionStore interface {
	getAll() []*marginPosition
	getByCode(code string) (*marginPosition, error)
	getBySymbolCode(symbolCode string) ([]*marginPosition, error)
	save(marginPosition *marginPosition)
	removeByCode(code string)
}

// marginPositionStore - 信用株式ポジションのストア
type marginPositionStore struct {
	store map[string]*marginPosition
	mtx   sync.Mutex
}

// getAll - ストアのすべてのポジションをコード順に並べて返す
func (s *marginPositionStore) getAll() []*marginPosition {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	positions := make([]*marginPosition, len(s.store))
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
func (s *marginPositionStore) getByCode(code string) (*marginPosition, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if position, ok := s.store[code]; ok {
		return position, nil
	} else {
		return nil, NoDataError
	}
}

// getBySymbolCode - 銘柄コードを指定してデータを取得する
func (s *marginPositionStore) getBySymbolCode(symbolCode string) ([]*marginPosition, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	positions := make([]*marginPosition, 0)
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
func (s *marginPositionStore) save(marginPosition *marginPosition) {
	if marginPosition == nil {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[marginPosition.Code] = marginPosition
}

// removeByCode - コードを指定して削除する
func (s *marginPositionStore) removeByCode(code string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.store, code)
}
