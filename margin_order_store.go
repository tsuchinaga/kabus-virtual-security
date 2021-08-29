package virtual_security

import (
	"sort"
	"sync"
)

var (
	marginOrderStoreSingleton      iMarginOrderStore
	marginOrderStoreSingletonMutex sync.Mutex
)

func getMarginOrderStore() iMarginOrderStore {
	marginOrderStoreSingletonMutex.Lock()
	defer marginOrderStoreSingletonMutex.Unlock()

	if marginOrderStoreSingleton == nil {
		marginOrderStoreSingleton = &marginOrderStore{
			store: map[string]*marginOrder{},
		}
	}
	return marginOrderStoreSingleton
}

// iMarginOrderStore - 信用株式注文ストアのインターフェース
type iMarginOrderStore interface {
	getAll() []*marginOrder
	getByCode(code string) (*marginOrder, error)
	save(marginOrder *marginOrder)
	removeByCode(code string)
}

// marginOrderStore - 信用株式注文のストア
type marginOrderStore struct {
	store map[string]*marginOrder
	mtx   sync.Mutex
}

// getAll - ストアのすべての注文をコード順に並べて返す
func (s *marginOrderStore) getAll() []*marginOrder {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	orders := make([]*marginOrder, len(s.store))
	var i int
	for _, order := range s.store {
		orders[i] = order
		i++
	}
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Code < orders[j].Code
	})
	return orders
}

// getByCode - コードを指定してデータを取得する
func (s *marginOrderStore) getByCode(code string) (*marginOrder, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if order, ok := s.store[code]; ok {
		return order, nil
	} else {
		return nil, NoDataError
	}
}

// save - 注文をストアに追加する
func (s *marginOrderStore) save(marginOrder *marginOrder) {
	if marginOrder == nil {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[marginOrder.Code] = marginOrder
}

// removeByCode - コードを指定して削除する
func (s *marginOrderStore) removeByCode(code string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.store, code)
}
