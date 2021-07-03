package virtual_security

import (
	"fmt"
)

type Security interface {
	RegisterPrice(symbolPrice RegisterPriceRequest) error      // 銘柄価格の登録
	StockOrder(order *StockOrderRequest) (*OrderResult, error) // 現物注文
	CancelStockOrder(cancelOrder *CancelOrderRequest) error    // 注文の取り消し
	StockOrders() ([]*StockOrder, error)                       // 現物注文一覧
	StockPositions() ([]*StockPosition, error)                 // 現物ポジション一覧
}

type security struct {
	clock        Clock
	priceStore   PriceStore
	priceService PriceService
	stockService StockService
}

// RegisterPrice - 価格の登録
func (s *security) RegisterPrice(symbolPrice RegisterPriceRequest) error {
	if err := s.priceService.validation(symbolPrice); err != nil {
		return err
	}

	// 内部用価格情報に変換
	price, err := s.priceService.toSymbolPrice(symbolPrice)
	if err != nil {
		return err
	}

	// 保存
	if err := s.priceStore.Set(price); err != nil {
		return err
	}

	// 約定確認
	now := s.clock.Now()
	session := s.clock.GetStockSession(now) // priceのsessionと一致しないことがあるため現在時刻で取得する
	for _, o := range s.stockService.GetStockOrders() {
		res := o.confirmContract(price, now, session)
		if res.isContracted {
			switch o.Side {
			case SideBuy:
				_ = s.stockService.Entry(o, res)
			case SideSell:
				_ = s.stockService.Exit(o, res)
			}
		}
	}

	return nil
}

// StockOrder - 現物注文
func (s *security) StockOrder(order *StockOrderRequest) (*OrderResult, error) {
	if order == nil {
		return nil, NilArgumentError
	}

	now := s.clock.Now()

	// 注文番号発行
	o := &stockOrder{
		Code:               s.stockService.NewOrderCode(),
		OrderStatus:        OrderStatusInOrder,
		Side:               order.Side,
		ExecutionCondition: order.ExecutionCondition,
		SymbolCode:         order.SymbolCode,
		OrderQuantity:      order.Quantity,
		LimitPrice:         order.LimitPrice,
		ExpiredAt:          order.ExpiredAt,
		StopCondition:      order.StopCondition,
		OrderedAt:          now,
		Contracts:          []*Contract{},
	}

	// validation
	if err := o.isValid(now); err != nil {
		return nil, err
	}

	// 該当銘柄の価格取得
	price, err := s.priceStore.GetBySymbolCode(order.SymbolCode)
	if err == nil {
		// 価格があれば約定確認
		// sessionを特定する
		session := s.clock.GetStockSession(now)
		res := o.confirmContract(price, now, session)

		// 約定していたら約定状態に更新し、ポジションも更新する
		if res.isContracted {
			var err error
			switch order.Side {
			case SideBuy:
				err = s.stockService.Entry(o, res)
			case SideSell:
				err = s.stockService.Exit(o, res)
			}
			if err != nil {
				return nil, err
			}
		}
	} else if err == NoDataError {
		if err := s.stockService.AddStockOrder(o); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 注文番号を返す
	return &OrderResult{OrderCode: o.Code}, nil
}

// CancelStockOrder - 現物注文の取消
func (s *security) CancelStockOrder(cancelOrder *CancelOrderRequest) error {
	if cancelOrder == nil {
		return fmt.Errorf("cancelOrder is nil, %w", NilArgumentError)
	}

	order, err := s.stockService.GetStockOrderByCode(cancelOrder.OrderCode)
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
	orders := s.stockService.GetStockOrders()

	res := make([]*StockOrder, len(orders))
	i := 0
	for _, o := range orders {
		if o.isDied(now) {
			s.stockService.RemoveStockOrderByCode(o.Code)
			res = res[:len(res)-1]
			continue
		}
		res[i] = &StockOrder{
			Code:               o.Code,
			OrderStatus:        o.OrderStatus,
			Side:               o.Side,
			ExecutionCondition: o.ExecutionCondition,
			SymbolCode:         o.SymbolCode,
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
			ClosePositionCode:  o.ClosePositionCode,
		}
		i++
	}
	return res, nil
}

// StockPositions - 現物ポジション一覧
func (s *security) StockPositions() ([]*StockPosition, error) {
	positions := s.stockService.GetStockPositions()

	res := make([]*StockPosition, len(positions))
	i := 0
	for _, p := range positions {
		if p.isDied() {
			s.stockService.RemoveStockPositionByCode(p.Code)
			res = res[:len(res)-1]
			continue
		}
		res[i] = &StockPosition{
			Code:               p.Code,
			OrderCode:          p.OrderCode,
			SymbolCode:         p.SymbolCode,
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
