package virtual_security

import (
	"fmt"
)

func NewVirtualSecurity() VirtualSecurity {
	return &virtualSecurity{
		clock:        newClock(),
		priceService: newPriceService(newClock(), getPriceStore(newClock())),
		stockService: newStockService(newUUIDGenerator(), getStockOrderStore(), getStockPositionStore()),
	}
}

type VirtualSecurity interface {
	RegisterPrice(symbolPrice RegisterPriceRequest) error      // 銘柄価格の登録
	StockOrder(order *StockOrderRequest) (*OrderResult, error) // 現物注文
	CancelStockOrder(cancelOrder *CancelOrderRequest) error    // 注文の取り消し
	StockOrders() ([]*StockOrder, error)                       // 現物注文一覧
	StockPositions() ([]*StockPosition, error)                 // 現物ポジション一覧
}

type virtualSecurity struct {
	clock        iClock
	priceService iPriceService
	stockService iStockService
}

// RegisterPrice - 価格の登録
func (s *virtualSecurity) RegisterPrice(symbolPrice RegisterPriceRequest) error {
	if err := s.priceService.validation(symbolPrice); err != nil {
		return err
	}

	// 内部用価格情報に変換
	price, err := s.priceService.toSymbolPrice(symbolPrice)
	if err != nil {
		return err
	}

	// 保存
	if err := s.priceService.set(price); err != nil {
		return err
	}

	// 約定確認
	now := s.clock.now()
	session := s.clock.getStockSession(now) // priceのsessionと一致しないことがあるため現在時刻で取得する
	for _, o := range s.stockService.getStockOrders() {
		res := o.confirmContract(price, now, session)
		if res.isContracted {
			switch o.Side {
			case SideBuy:
				_ = s.stockService.entry(o, res)
			case SideSell:
				_ = s.stockService.exit(o, res)
			}
		}
	}

	return nil
}

// StockOrder - 現物注文
func (s *virtualSecurity) StockOrder(order *StockOrderRequest) (*OrderResult, error) {
	if order == nil {
		return nil, NilArgumentError
	}

	now := s.clock.now()

	// 注文番号発行
	o := &stockOrder{
		Code:               s.stockService.newOrderCode(),
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
	price, err := s.priceService.getBySymbolCode(order.SymbolCode)
	if err == nil {
		// 価格があれば約定確認
		// sessionを特定する
		session := s.clock.getStockSession(now)
		res := o.confirmContract(price, now, session)

		// 約定していたら約定状態に更新し、ポジションも更新する
		if res.isContracted {
			var err error
			switch order.Side {
			case SideBuy:
				err = s.stockService.entry(o, res)
			case SideSell:
				err = s.stockService.exit(o, res)
			}
			if err != nil {
				return nil, err
			}
		}
	} else if err == NoDataError {
		if err := s.stockService.addStockOrder(o); err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}

	// 注文番号を返す
	return &OrderResult{OrderCode: o.Code}, nil
}

// CancelStockOrder - 現物注文の取消
func (s *virtualSecurity) CancelStockOrder(cancelOrder *CancelOrderRequest) error {
	if cancelOrder == nil {
		return fmt.Errorf("cancelOrder is nil, %w", NilArgumentError)
	}

	order, err := s.stockService.getStockOrderByCode(cancelOrder.OrderCode)
	if err != nil {
		return fmt.Errorf("not found stock order(code: %s), %w", cancelOrder.OrderCode, err)
	}

	if !order.OrderStatus.IsCancelable() {
		return UncancellableOrderError
	}

	order.cancel(s.clock.now())
	return nil
}

// StockOrders - 現物注文一覧
func (s *virtualSecurity) StockOrders() ([]*StockOrder, error) {
	now := s.clock.now()
	orders := s.stockService.getStockOrders()

	res := make([]*StockOrder, len(orders))
	i := 0
	for _, o := range orders {
		if o.isDied(now) {
			s.stockService.removeStockOrderByCode(o.Code)
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
func (s *virtualSecurity) StockPositions() ([]*StockPosition, error) {
	positions := s.stockService.getStockPositions()

	res := make([]*StockPosition, len(positions))
	i := 0
	for _, p := range positions {
		if p.isDied() {
			s.stockService.removeStockPositionByCode(p.Code)
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
