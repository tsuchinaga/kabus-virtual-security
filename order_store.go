package virtual_security

import (
	"sort"
	"sync"
)

var (
	stockOrderStoreSingleton      StockOrderStore
	stockOrderStoreSingletonMutex sync.Mutex
)

func GetStockOrderStore() StockOrderStore {
	stockOrderStoreSingletonMutex.Lock()
	defer stockOrderStoreSingletonMutex.Unlock()

	if stockOrderStoreSingleton == nil {
		stockOrderStoreSingleton = &stockOrderStore{
			store: map[string]*StockOrder{},
		}
	}
	return stockOrderStoreSingleton
}

// StockOrderStore - 現物株式注文ストアのインターフェース
type StockOrderStore interface {
	GetAll() []*StockOrder
	GetByCode(code string) (*StockOrder, error)
	Add(stockOrder *StockOrder)
	RemoveByCode(code string)
}

// stockOrderStore - 現物株式注文のストア
type stockOrderStore struct {
	store map[string]*StockOrder
	mtx   sync.Mutex
}

// GetAll - ストアのすべての注文をコード順に並べて返す
func (s *stockOrderStore) GetAll() []*StockOrder {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	orders := make([]*StockOrder, len(s.store))
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

// GetByCode - コードを指定してデータを取得する
func (s *stockOrderStore) GetByCode(code string) (*StockOrder, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if order, ok := s.store[code]; ok {
		return order, nil
	} else {
		return nil, NoDataError
	}
}

// Add - 注文をストアに追加する
func (s *stockOrderStore) Add(stockOrder *StockOrder) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.store[stockOrder.Code] = stockOrder
}

// RemoveByCode - コードを指定して削除する
func (s *stockOrderStore) RemoveByCode(code string) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	delete(s.store, code)
}
