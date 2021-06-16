package virtual_security

import "fmt"

type Security interface {
	RegisterPrice(symbolPrice SymbolPrice) (*UpdatedOrders, error) // 銘柄価格の登録
	StockOrder(order *StockOrderRequest) (*OrderResult, error)     // 現物注文
	CancelStockOrder(cancelOrder *CancelOrderRequest) error        // 注文の取り消し
	StockOrders() ([]*StockOrder, error)                           // 現物注文一覧
	StockPositions() ([]*StockPosition, error)                     // 現物ポジション一覧
}

type security struct {
	clock              Clock
	stockOrderStore    StockOrderStore
	stockPositionStore StockPositionStore
}

// CancelStockOrder - 現物注文の取消
func (s *security) CancelStockOrder(cancelOrder *CancelOrderRequest) error {
	if cancelOrder == nil {
		return fmt.Errorf("cancelOrder is nil, %w", NilArgumentError)
	}

	order, err := s.stockOrderStore.GetByCode(cancelOrder.OrderCode)
	if err != nil {
		return fmt.Errorf("not found stock order(code: %s), %w", cancelOrder.OrderCode, err)
	}

	if !order.OrderStatus.IsCancelable() {
		return UncancellableOrderError
	}

	order.cancel(s.clock.Now())
	return nil
}

// StockOrders - 現物注文一覧
func (s *security) StockOrders() ([]*StockOrder, error) {
	now := s.clock.Now()
	orders := s.stockOrderStore.GetAll()

	res := make([]*StockOrder, len(orders))
	i := 0
	for _, o := range orders {
		if o.isDied(now) {
			s.stockOrderStore.RemoveByCode(o.Code)
			res = res[:len(res)-1]
			continue
		}
		res[i] = &StockOrder{
			Code:               o.Code,
			OrderStatus:        o.OrderStatus,
			Side:               o.Side,
			ExecutionCondition: o.ExecutionCondition,
			SymbolCode:         o.SymbolCode,
			Exchange:           o.Exchange,
			OrderQuantity:      o.OrderQuantity,
			ContractedQuantity: o.ContractedQuantity,
			CanceledQuantity:   o.CanceledQuantity,
			LimitPrice:         o.LimitPrice,
			ExpiredAt:          o.ExpiredAt,
			StopCondition:      o.StopCondition,
			OrderedAt:          o.OrderedAt,
			CanceledAt:         o.CanceledAt,
			Contracts:          o.Contracts,
			Message:            o.Message,
		}
		i++
	}
	return res, nil
}

// StockPositions - 現物ポジション一覧
func (s *security) StockPositions() ([]*StockPosition, error) {
	positions := s.stockPositionStore.GetAll()

	res := make([]*StockPosition, len(positions))
	i := 0
	for _, p := range positions {
		if p.isDied() {
			s.stockPositionStore.RemoveByCode(p.Code)
			res = res[:len(res)-1]
			continue
		}
		res[i] = &StockPosition{
			Code:               p.Code,
			OrderCode:          p.OrderCode,
			SymbolCode:         p.SymbolCode,
			Exchange:           p.Exchange,
			Side:               p.Side,
			ContractedQuantity: p.ContractedQuantity,
			OwnedQuantity:      p.OwnedQuantity,
			HoldQuantity:       p.HoldQuantity,
			ContractedAt:       p.ContractedAt,
		}
		i++
	}
	return res, nil
}
