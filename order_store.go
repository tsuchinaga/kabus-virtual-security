package virtual_security

import (
	"sort"
	"sync"
)

var (
	stockOrderStoreSingleton      StockOrderStore
	stockOrderStoreSingletonMutex sync.Mutex
)

func getStockOrderStore() StockOrderStore {
	stockOrderStoreSingletonMutex.Lock()
	defer stockOrderStoreSingletonMutex.Unlock()

	if stockOrderStoreSingleton == nil {
		stockOrderStoreSingleton = &stockOrderStore{
			store: map[string]*stockOrder{},
		}
	}
	return stockOrderStoreSingleton
}

// StockOrderStore - 現物株式注文ストアのインターフェース
// TODO clean って名前で、一定期間更新されていない終了している注文を消す処理を追加する
type StockOrderStore interface {
	GetAll() []*stockOrder
	GetByCode(code string) (*stockOrder, error)
	Add(stockOrder *stockOrder)
	RemoveByCode(code string)
}

// stockOrderStore - 現物株式注文のストア
type stockOrderStore struct {
	store map[string]*stockOrder
	mtx   sync.Mutex
}

// GetAll - ストアのすべての注文をコード順に並べて返す
func (s *stockOrderStore) GetAll() []*stockOrder {
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

// GetByCode - コードを指定してデータを取得する
func (s *stockOrderStore) GetByCode(code string) (*stockOrder, error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if order, ok := s.store[code]; ok {
		return order, nil
	} else {
		return nil, NoDataError
	}
}

// Add - 注文をストアに追加する
func (s *stockOrderStore) Add(stockOrder *stockOrder) {
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
