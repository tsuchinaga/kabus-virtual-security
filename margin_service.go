package virtual_security

import (
	"fmt"
	"sync"
	"time"
)

func newMarginService(
	uuidGenerator iUUIDGenerator,
	marginOrderStore iMarginOrderStore,
	marginPositionStore iMarginPositionStore,
	validatorComponent iValidatorComponent,
	stockContractComponent iStockContractComponent,
) iMarginService {
	return &marginService{
		uuidGenerator:          uuidGenerator,
		marginOrderStore:       marginOrderStore,
		marginPositionStore:    marginPositionStore,
		validatorComponent:     validatorComponent,
		stockContractComponent: stockContractComponent,
	}
}

type iMarginService interface {
	toMarginOrder(order *MarginOrderRequest, now time.Time) *marginOrder
	validation(order *marginOrder, now time.Time) error
	entry(order *marginOrder, price *symbolPrice, now time.Time) error
	exit(order *marginOrder, price *symbolPrice, now time.Time) error
	holdExitOrderPositions(order *marginOrder) error
	getMarginOrders() []*marginOrder
	getMarginOrderByCode(orderCode string) (*marginOrder, error)
	saveMarginOrder(order *marginOrder)
	removeMarginOrderByCode(orderCode string)
	getMarginPositions() []*marginPosition
	removeMarginPositionByCode(positionCode string)
	confirmContract(order *marginOrder, price *symbolPrice, now time.Time) *confirmContractResult
}

type marginService struct {
	uuidGenerator          iUUIDGenerator
	marginOrderStore       iMarginOrderStore
	marginPositionStore    iMarginPositionStore
	validatorComponent     iValidatorComponent
	stockContractComponent iStockContractComponent
}

func (s *marginService) newOrderCode() string {
	return "mor-" + s.uuidGenerator.generate()
}

func (s *marginService) newContractCode() string {
	return "mco-" + s.uuidGenerator.generate()
}

func (s *marginService) newPositionCode() string {
	return "mpo-" + s.uuidGenerator.generate()
}

func (s *marginService) toMarginOrder(order *MarginOrderRequest, now time.Time) *marginOrder {
	if order == nil {
		return nil
	}

	o := &marginOrder{
		Code:               s.newOrderCode(),
		OrderStatus:        OrderStatusInOrder,
		TradeType:          order.TradeType,
		Side:               order.Side,
		ExecutionCondition: order.ExecutionCondition,
		SymbolCode:         order.SymbolCode,
		OrderQuantity:      order.Quantity,
		LimitPrice:         order.LimitPrice,
		ExpiredAt:          time.Time{},
		StopCondition:      order.StopCondition,
		OrderedAt:          now,
		ExitPositionList:   order.ExitPositionList,
		Contracts:          []*Contract{},
	}

	if order.ExpiredAt.IsZero() {
		o.ExpiredAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	} else {
		o.ExpiredAt = time.Date(order.ExpiredAt.Year(), order.ExpiredAt.Month(), order.ExpiredAt.Day(), 0, 0, 0, 0, time.Local)
	}
	return o
}

func (s *marginService) validation(order *marginOrder, now time.Time) error {
	return s.validatorComponent.isValidMarginOrder(order, now, s.marginPositionStore.getAll())
}

func (s *marginService) entry(order *marginOrder, price *symbolPrice, now time.Time) error {
	// 最低限のvalidation
	if order == nil || price == nil {
		return NilArgumentError
	}

	// 約定可能かのチェック
	contractResult := s.stockContractComponent.confirmMarginOrderContract(order, price, now)
	if !contractResult.isContracted {
		return nil
	}

	contractCode := s.newContractCode()
	positionCode := s.newPositionCode()
	order.contract(&Contract{
		ContractCode: contractCode,
		OrderCode:    order.Code,
		PositionCode: positionCode,
		Price:        contractResult.price,
		Quantity:     order.OrderQuantity,
		ContractedAt: contractResult.contractedAt,
	})

	s.marginPositionStore.save(&marginPosition{
		Code:               positionCode,
		OrderCode:          order.Code,
		SymbolCode:         order.SymbolCode,
		Side:               order.Side,
		ContractedQuantity: order.ContractedQuantity,
		OwnedQuantity:      order.ContractedQuantity,
		Price:              contractResult.price,
		ContractedAt:       contractResult.contractedAt,
		mtx:                sync.Mutex{},
	})

	return nil
}

func (s *marginService) exit(order *marginOrder, price *symbolPrice, now time.Time) error {
	// 最低限のvalidation
	if order == nil || price == nil {
		return NilArgumentError
	}

	// 約定可能かのチェックし保存
	contractResult := s.stockContractComponent.confirmMarginOrderContract(order, price, now)
	if !contractResult.isContracted {
		return nil
	}

	// 指定されたポジションの一覧を取得し、exit可能かのチェック
	positions := make(map[string]*marginPosition)
	for _, ep := range order.ExitPositionList {
		p, err := s.marginPositionStore.getByCode(ep.PositionCode)
		if err != nil {
			return fmt.Errorf("position code: %s: %w", ep.PositionCode, err)
		}
		if err := p.exitable(ep.Quantity); err != nil {
			return fmt.Errorf("position code: %s: %w", ep.PositionCode, err)
		}
		positions[p.Code] = p
	}

	for _, ep := range order.ExitPositionList {
		p := positions[ep.PositionCode]

		// ポジションの保有数量を返済する
		// TODO 先にexit可能かのチェックをしているから基本的にエラーは無視できるけど、必要ならエラーチェックを追加する
		_ = p.exit(ep.Quantity)

		// 注文に約定情報を追加
		contractCode := s.newContractCode()
		order.contract(&Contract{
			ContractCode: contractCode,
			OrderCode:    order.Code,
			PositionCode: p.Code,
			Price:        contractResult.price,
			Quantity:     ep.Quantity,
			ContractedAt: contractResult.contractedAt,
		})
	}

	return nil
}

func (s *marginService) holdExitOrderPositions(order *marginOrder) error {
	// 最低限のvalidation
	if order == nil {
		return NilArgumentError
	}
	if order.TradeType != TradeTypeExit {
		return InvalidTradeTypeError
	}
	if order.ExitPositionList == nil || len(order.ExitPositionList) == 0 {
		return InvalidExitPositionError
	}

	// positionをhold可能かのチェック
	posList := make([]*marginPosition, len(order.ExitPositionList))
	for i, exit := range order.ExitPositionList {
		pos, err := s.marginPositionStore.getByCode(exit.PositionCode)
		if err != nil {
			return fmt.Errorf("error position_code = %s: %w", exit.PositionCode, err)
		}
		if err := pos.holdable(exit.Quantity); err != nil {
			return fmt.Errorf("error position_code = %s, owned_quantity = %.2f, hold_quantity = %.2f, exit_quantity = %.2f: %w",
				exit.PositionCode, pos.OwnedQuantity, pos.HoldQuantity, exit.Quantity, err)
		}
		posList[i] = pos
	}

	// チェック済みのポジションをholdする
	for i, exit := range order.ExitPositionList {
		pos := posList[i]
		_ = pos.hold(exit.Quantity)
	}

	return nil
}

func (s *marginService) getMarginOrders() []*marginOrder {
	return s.marginOrderStore.getAll()
}

func (s *marginService) getMarginOrderByCode(orderCode string) (*marginOrder, error) {
	return s.marginOrderStore.getByCode(orderCode)
}

func (s *marginService) saveMarginOrder(order *marginOrder) {
	s.marginOrderStore.save(order)
}

func (s *marginService) removeMarginOrderByCode(orderCode string) {
	s.marginOrderStore.removeByCode(orderCode)
}

func (s *marginService) getMarginPositions() []*marginPosition {
	return s.marginPositionStore.getAll()
}

func (s *marginService) removeMarginPositionByCode(positionCode string) {
	s.marginPositionStore.removeByCode(positionCode)
}

func (s *marginService) confirmContract(order *marginOrder, price *symbolPrice, now time.Time) *confirmContractResult {
	if order == nil || price == nil {
		return &confirmContractResult{isContracted: false}
	}

	// 銘柄が同一でなければfalse
	if order.SymbolCode != price.SymbolCode {
		return &confirmContractResult{isContracted: false}
	}

	// 約定確認中は状態を変更されたくないのでロック
	order.lock()
	defer order.unlock()

	// 約定可能な注文状態でなければfalse
	if !order.OrderStatus.IsContractable() {
		return &confirmContractResult{isContracted: false}
	}

	// 約定可能時間でなければ約定しない
	if !s.stockContractComponent.isContractableTime(order.executionCondition(), now) {
		return &confirmContractResult{isContracted: false}
	}

	// 執行条件ごとの約定チェック
	//   ここまできたということは、時間条件などはパスしているということ
	res := s.stockContractComponent.confirmMarginOrderContract(order, price, now)

	order.ConfirmingCount++
	return res
}
