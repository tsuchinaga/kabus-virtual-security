package virtual_security

import (
	"fmt"
)

func NewVirtualSecurity() VirtualSecurity {
	return &virtualSecurity{
		clock:         newClock(),
		priceService:  newPriceService(newClock(), getPriceStore(newClock())),
		stockService:  newStockService(newUUIDGenerator(), getStockOrderStore(), getStockPositionStore(), newValidatorComponent(), newStockContractComponent()),
		marginService: newMarginService(newUUIDGenerator(), getMarginOrderStore(), getMarginPositionStore(), newValidatorComponent(), newStockContractComponent()),
	}
}

type VirtualSecurity interface {
	RegisterPrice(symbolPrice RegisterPriceRequest) error // 銘柄価格の登録

	StockOrder(order *StockOrderRequest) (*OrderResult, error) // 現物注文
	CancelStockOrder(cancelOrder *CancelOrderRequest) error    // 現物注文の取り消し
	StockOrders() ([]*StockOrder, error)                       // 現物注文一覧
	StockPositions() ([]*StockPosition, error)                 // 現物ポジション一覧

	MarginOrder(order *MarginOrderRequest) (*OrderResult, error) // 信用注文
	CancelMarginOrder(cancelOrder *CancelOrderRequest) error     // 信用注文の取り消し
	MarginOrders() ([]*MarginOrder, error)                       // 信用注文一覧
	MarginPositions() ([]*MarginPosition, error)                 // 信用ポジション一覧
}

type virtualSecurity struct {
	clock         iClock
	priceService  iPriceService
	stockService  iStockService
	marginService iMarginService
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

	// 現物約定確認
	now := s.clock.now()
	for _, o := range s.stockService.getStockOrders() {
		switch o.Side {
		case SideBuy:
			_ = s.stockService.entry(o, price, now)
		case SideSell:
			_ = s.stockService.exit(o, price, now)
		}
	}

	// 信用約定確認
	for _, o := range s.marginService.getMarginOrders() {
		switch o.TradeType {
		case TradeTypeEntry:
			_ = s.marginService.entry(o, price, now)
		case TradeTypeExit:
			_ = s.marginService.exit(o, price, now)
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
	o := s.stockService.toStockOrder(order, now)

	// validation
	if err := s.stockService.validation(o, now); err != nil {
		return nil, err
	}

	// sell注文ならsellするポジションをholdする
	if o.Side == SideSell {
		if err := s.stockService.holdSellOrderPositions(o); err != nil {
			return nil, err
		}
	}

	// ここまでこれば有効な注文なので、処理後に保存する
	defer s.stockService.saveStockOrder(o)

	// 該当銘柄の価格取得
	price, priceErr := s.priceService.getBySymbolCode(order.SymbolCode)
	if priceErr != nil && priceErr != NoDataError {
		return nil, priceErr
	}

	// 価格情報がNoDataでなければ最初の約定確認処理をする
	// 注文でエラーがでても使い道がないので捨てる
	if priceErr != NoDataError {
		switch order.Side {
		case SideBuy:
			_ = s.stockService.entry(o, price, now)
		case SideSell:
			_ = s.stockService.exit(o, price, now)
		}
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
			Price:              p.Price,
		}
		i++
	}
	return res, nil
}

// MarginOrder - 信用注文
func (s *virtualSecurity) MarginOrder(order *MarginOrderRequest) (*OrderResult, error) {
	if order == nil {
		return nil, NilArgumentError
	}

	now := s.clock.now()

	// 内部用注文に変換
	o := s.marginService.toMarginOrder(order, now)

	// validation
	if err := s.marginService.validation(o, now); err != nil {
		return nil, err
	}

	// exit注文ならexitするポジションをholdする
	if o.TradeType == TradeTypeExit {
		if err := s.marginService.holdExitOrderPositions(o); err != nil {
			return nil, err
		}
	}

	// ここまでこれば有効な注文なので、処理後に保存する
	defer s.marginService.saveMarginOrder(o)

	// 該当銘柄の価格取得
	price, priceErr := s.priceService.getBySymbolCode(order.SymbolCode)
	if priceErr != nil && priceErr != NoDataError {
		return nil, priceErr
	}

	// 価格情報がNoDataでなければ最初の約定確認処理をする
	// 注文でエラーがでても使い道がないので捨てる
	if priceErr != NoDataError {
		switch o.TradeType {
		case TradeTypeEntry:
			_ = s.marginService.entry(o, price, now)
		case TradeTypeExit:
			_ = s.marginService.exit(o, price, now)
		}
	}

	// 注文番号を返す
	return &OrderResult{OrderCode: o.Code}, nil
}

// CancelMarginOrder - 信用注文の取消
func (s *virtualSecurity) CancelMarginOrder(cancelOrder *CancelOrderRequest) error {
	if cancelOrder == nil {
		return fmt.Errorf("cancelOrder is nil, %w", NilArgumentError)
	}

	order, err := s.marginService.getMarginOrderByCode(cancelOrder.OrderCode)
	if err != nil {
		return fmt.Errorf("not found margin order(code: %s), %w", cancelOrder.OrderCode, err)
	}

	if !order.OrderStatus.IsCancelable() {
		return UncancellableOrderError
	}

	order.cancel(s.clock.now())
	return nil
}

// MarginOrders - 信用注文一覧
func (s *virtualSecurity) MarginOrders() ([]*MarginOrder, error) {
	now := s.clock.now()
	orders := s.marginService.getMarginOrders()

	res := make([]*MarginOrder, len(orders))
	i := 0
	for _, o := range orders {
		if o.isDied(now) {
			s.marginService.removeMarginOrderByCode(o.Code)
			res = res[:len(res)-1]
			continue
		}
		res[i] = &MarginOrder{
			Code:               o.Code,
			OrderStatus:        o.OrderStatus,
			TradeType:          o.TradeType,
			Side:               o.Side,
			ExecutionCondition: o.ExecutionCondition,
			SymbolCode:         o.SymbolCode,
			OrderQuantity:      o.OrderQuantity,
			ContractedQuantity: o.ContractedQuantity,
			CanceledQuantity:   o.CanceledQuantity,
			LimitPrice:         o.LimitPrice,
			ExpiredAt:          o.ExpiredAt,
			StopCondition:      o.StopCondition,
			ExitPositionList:   o.ExitPositionList,
			OrderedAt:          o.OrderedAt,
			CanceledAt:         o.CanceledAt,
			Contracts:          o.Contracts,
			Message:            o.Message,
		}
		i++
	}
	return res, nil
}

// MarginPositions - 信用ポジション一覧
func (s *virtualSecurity) MarginPositions() ([]*MarginPosition, error) {
	positions := s.marginService.getMarginPositions()

	res := make([]*MarginPosition, len(positions))
	i := 0
	for _, p := range positions {
		if p.isDied() {
			s.marginService.removeMarginPositionByCode(p.Code)
			res = res[:len(res)-1]
			continue
		}
		res[i] = &MarginPosition{
			Code:               p.Code,
			OrderCode:          p.OrderCode,
			SymbolCode:         p.SymbolCode,
			Side:               p.Side,
			ContractedQuantity: p.ContractedQuantity,
			OwnedQuantity:      p.OwnedQuantity,
			HoldQuantity:       p.HoldQuantity,
			ContractedAt:       p.ContractedAt,
			Price:              p.Price,
		}
		i++
	}
	return res, nil
}
