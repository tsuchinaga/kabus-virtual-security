package virtual_security

import (
	"sort"
	"sync"
)

var (
	stockOrderStoreSingleton      iStockOrderStore
	stockOrderStoreSingletonMutex sync.Mutex
)

func getStockOrderStore() iStockOrderStore {
	stockOrderStoreSingletonMutex.Lock()
	defer stockOrderStoreSingletonMutex.Unlock()

	if stockOrderStoreSingleton == nil {
		stockOrderStoreSingleton = &stockOrderStore{
			store: map[string]*stockOrder{},
		}
	}
	return stockOrderStoreSingleton
}

// iStockOrderStore - 現物株式注文ストアのインターフェース
type iStockOrderStore interface {
	getAll() []*stockOrder
	getByCode(code string) (*stockOrder, error)
	save(stockOrder *stockOrder)
	removeByCode(code string)
}

// stockOrderStore - 現物株式注文のストア
type stockOrderStore struct {
	store map[string]*stockOrder
	mtx   sync.Mutex
}

// getAll - ストアのすべての注文をコード順に並べて返す
func (s *stockOrderStore) getAll() []*stockOrder {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	orders := make([]*stockOrder, len(s.store))
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
func (s *stockOrderStore) getByCode(code string) (*stockOrder, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if order, ok := s.store[code]; ok {
		return order, nil
	} else {
		return nil, NoDataError
	}
}

// save - 注文をストアに追加する
func (s *stockOrderStore) save(stockOrder *stockOrder) {
	if stockOrder == nil {
		return
	}

	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[stockOrder.Code] = stockOrder
}

// removeByCode - コードを指定して削除する
func (s *stockOrderStore) removeByCode(code string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.store, code)
}
