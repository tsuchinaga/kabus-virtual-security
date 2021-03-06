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
	confirmContract(order *marginOrder, price *symbolPrice, now time.Time) error
	holdExitOrderPositions(order *marginOrder) error
	getMarginOrders() []*marginOrder
	getMarginOrderByCode(orderCode string) (*marginOrder, error)
	saveMarginOrder(order *marginOrder)
	removeMarginOrderByCode(orderCode string)
	getMarginPositions() []*marginPosition
	removeMarginPositionByCode(positionCode string)
	cancelAndRelease(order *marginOrder, now time.Time) error
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

func (s *marginService) confirmContract(order *marginOrder, price *symbolPrice, now time.Time) error {
	// ????????????validation
	if order == nil || price == nil {
		return NilArgumentError
	}

	switch order.TradeType {
	case TradeTypeEntry:
		return s.entry(order, price, now)
	case TradeTypeExit:
		return s.exit(order, price, now)
	default:
		return InvalidTradeTypeError
	}
}

func (s *marginService) entry(order *marginOrder, price *symbolPrice, now time.Time) error {
	// ????????????validation
	if order == nil || price == nil {
		return NilArgumentError
	}

	// ??????????????????????????????
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
	// ????????????validation
	if order == nil || price == nil {
		return NilArgumentError
	}

	// ??????????????????????????????
	contractResult := s.stockContractComponent.confirmMarginOrderContract(order, price, now)
	if !contractResult.isContracted {
		return nil
	}

	// ??????????????????????????????????????????????????????exit????????????????????????
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

		// ?????????????????????????????????????????????
		// TODO ??????exit?????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????????
		_ = p.exit(ep.Quantity)
		order.addExitPosition(p.Code, ep.Quantity) // ?????????????????????????????????????????????

		// ??????????????????????????????
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
	// ????????????validation
	if order == nil {
		return NilArgumentError
	}
	if order.TradeType != TradeTypeExit {
		return InvalidTradeTypeError
	}
	if order.ExitPositionList == nil || len(order.ExitPositionList) == 0 {
		return InvalidExitPositionError
	}

	// position???hold????????????????????????
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

	// ???????????????????????????????????????hold??????
	for i, exit := range order.ExitPositionList {
		pos := posList[i]
		_ = pos.hold(exit.Quantity)
		order.addHoldPosition(pos.Code, exit.Quantity)
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

func (s *marginService) cancelAndRelease(order *marginOrder, now time.Time) error {
	if order == nil {
		return NilArgumentError
	}
	if !order.OrderStatus.IsCancelable() {
		return UncancellableOrderError
	}
	order.cancel(now)

	// ????????????????????????????????????????????????????????????
	var res error
	if order.TradeType == TradeTypeExit {
		for _, hp := range order.HoldPositions {
			// Exit????????????????????????
			if hp.HoldQuantity-hp.ExitQuantity == 0 {
				continue
			}
			pos, err := s.marginPositionStore.getByCode(hp.PositionCode)
			if err != nil {
				res = fmt.Errorf("?????????????????????????????????????????????????????????????????????: %w", err)
				continue
			}
			if err := pos.release(hp.HoldQuantity - hp.ExitQuantity); err != nil {
				res = fmt.Errorf("?????????????????????????????????????????????????????????????????????: %w", err)
			}
			hp.HoldQuantity = hp.ExitQuantity
		}
	}

	return res
}
