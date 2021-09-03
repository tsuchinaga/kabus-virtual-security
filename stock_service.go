package virtual_security

import (
	"fmt"
	"sync"
	"time"
)

func newStockService(
	uuidGenerator iUUIDGenerator,
	stockOrderStore iStockOrderStore,
	stockPositionStore iStockPositionStore,
	validatorComponent iValidatorComponent,
	stockContractComponent iStockContractComponent,
) iStockService {
	return &stockService{
		uuidGenerator:          uuidGenerator,
		stockOrderStore:        stockOrderStore,
		stockPositionStore:     stockPositionStore,
		validatorComponent:     validatorComponent,
		stockContractComponent: stockContractComponent,
	}
}

type iStockService interface {
	toStockOrder(order *StockOrderRequest, now time.Time) *stockOrder
	confirmContract(order *stockOrder, price *symbolPrice, now time.Time) error
	getStockOrders() []*stockOrder
	getStockOrderByCode(orderCode string) (*stockOrder, error)
	saveStockOrder(order *stockOrder)
	removeStockOrderByCode(orderCode string)
	getStockPositions() []*stockPosition
	removeStockPositionByCode(positionCode string)
	holdSellOrderPositions(order *stockOrder) error
	validation(order *stockOrder, now time.Time) error
	cancelAndRelease(order *stockOrder, now time.Time) error
}

type stockService struct {
	uuidGenerator          iUUIDGenerator
	stockOrderStore        iStockOrderStore
	stockPositionStore     iStockPositionStore
	validatorComponent     iValidatorComponent
	stockContractComponent iStockContractComponent
}

func (s *stockService) newOrderCode() string {
	return "sor-" + s.uuidGenerator.generate()
}

func (s *stockService) newContractCode() string {
	return "sco-" + s.uuidGenerator.generate()
}

func (s *stockService) newPositionCode() string {
	return "spo-" + s.uuidGenerator.generate()
}

func (s *stockService) confirmContract(order *stockOrder, price *symbolPrice, now time.Time) error {
	// 最低限のvalidation
	if order == nil || price == nil {
		return NilArgumentError
	}

	switch order.Side {
	case SideBuy:
		return s.entry(order, price, now)
	case SideSell:
		return s.exit(order, price, now)
	default:
		return InvalidSideError
	}
}

func (s *stockService) entry(order *stockOrder, price *symbolPrice, now time.Time) error {
	// 最低限のvalidation
	if order == nil || price == nil {
		return NilArgumentError
	}

	// 約定可能かのチェック
	contractResult := s.stockContractComponent.confirmStockOrderContract(order, price, now)
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

	s.stockPositionStore.save(&stockPosition{
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

func (s *stockService) exit(order *stockOrder, price *symbolPrice, now time.Time) error {
	// 最低限のvalidation
	if order == nil || price == nil {
		return NilArgumentError
	}

	// 約定可能かのチェックし保存
	contractResult := s.stockContractComponent.confirmStockOrderContract(order, price, now)
	if !contractResult.isContracted {
		return nil
	}

	// 注文が拘束しているポジションをexitしていく
	for _, hp := range order.HoldPositions {
		p, err := s.stockPositionStore.getByCode(hp.PositionCode)
		if err != nil {
			// TODO 注文エラーにする
			continue
		}

		// TODO exit時にエラーが出る可能性があれば、エラーをコントロールできるようにする
		_ = p.exit(hp.HoldQuantity)
		order.addExitPosition(p.Code, hp.HoldQuantity) // 注文による返済数に加算しておく

		// 注文に約定情報を追加
		contractCode := s.newContractCode()
		order.contract(&Contract{
			ContractCode: contractCode,
			OrderCode:    order.Code,
			PositionCode: p.Code,
			Price:        contractResult.price,
			Quantity:     hp.HoldQuantity,
			ContractedAt: contractResult.contractedAt,
		})
	}

	return nil
}

func (s *stockService) getStockOrders() []*stockOrder {
	return s.stockOrderStore.getAll()
}

func (s *stockService) getStockOrderByCode(orderCode string) (*stockOrder, error) {
	return s.stockOrderStore.getByCode(orderCode)
}

func (s *stockService) saveStockOrder(order *stockOrder) {
	s.stockOrderStore.save(order)
}

func (s *stockService) removeStockOrderByCode(orderCode string) {
	s.stockOrderStore.removeByCode(orderCode)
}

func (s *stockService) getStockPositions() []*stockPosition {
	return s.stockPositionStore.getAll()
}

func (s *stockService) removeStockPositionByCode(positionCode string) {
	s.stockPositionStore.removeByCode(positionCode)
}

func (s *stockService) toStockOrder(order *StockOrderRequest, now time.Time) *stockOrder {
	if order == nil {
		return nil
	}

	o := &stockOrder{
		Code:               s.newOrderCode(),
		OrderStatus:        OrderStatusInOrder,
		Side:               order.Side,
		ExecutionCondition: order.ExecutionCondition,
		SymbolCode:         order.SymbolCode,
		OrderQuantity:      order.Quantity,
		LimitPrice:         order.LimitPrice,
		ExpiredAt:          time.Time{},
		StopCondition:      order.StopCondition,
		OrderedAt:          now,
		Contracts:          []*Contract{},
	}

	if order.ExpiredAt.IsZero() {
		o.ExpiredAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	} else {
		o.ExpiredAt = time.Date(order.ExpiredAt.Year(), order.ExpiredAt.Month(), order.ExpiredAt.Day(), 0, 0, 0, 0, time.Local)
	}
	return o
}

func (s *stockService) holdSellOrderPositions(order *stockOrder) error {
	if order == nil {
		return NilArgumentError
	}

	positions := s.stockPositionStore.getAll()

	// 全数が足りるか
	var totalOrderableQuantity float64
	for _, pos := range positions {
		totalOrderableQuantity += pos.orderableQuantity()
	}
	if totalOrderableQuantity < order.OrderQuantity {
		return NotEnoughOwnedQuantityError
	}

	// 足りれば、個別にholdしていく
	required := order.OrderQuantity
	for _, pos := range positions {
		// ポジションの保有数が返したい数より少なければhold可能な数だけholdし、返したい数が少なければ返したい数だけ返す
		orderableQuantity := pos.orderableQuantity()
		if orderableQuantity < required {
			_ = pos.hold(orderableQuantity)
			order.addHoldPosition(pos.Code, orderableQuantity)
			required -= orderableQuantity
		} else {
			_ = pos.hold(required)
			order.addHoldPosition(pos.Code, required)
			break
		}
	}

	return nil
}

func (s *stockService) validation(order *stockOrder, now time.Time) error {
	return s.validatorComponent.isValidStockOrder(order, now, s.stockPositionStore.getAll())
}

func (s *stockService) cancelAndRelease(order *stockOrder, now time.Time) error {
	if order == nil {
		return NilArgumentError
	}
	if !order.OrderStatus.IsCancelable() {
		return UncancellableOrderError
	}
	order.cancel(now)

	// 売り注文なら拘束したポジションを開放する
	var res error
	if order.Side == SideSell {
		for _, hp := range order.HoldPositions {
			// Exit済みならスキップ
			if hp.HoldQuantity-hp.ExitQuantity == 0 {
				continue
			}
			pos, err := s.stockPositionStore.getByCode(hp.PositionCode)
			if err != nil {
				res = fmt.Errorf("拘束中の一部のポジションを解放できませんでした: %w", err)
				continue
			}
			if err := pos.release(hp.HoldQuantity - hp.ExitQuantity); err != nil {
				res = fmt.Errorf("拘束中の一部のポジションを解放できませんでした: %w", err)
			}
			hp.HoldQuantity = hp.ExitQuantity
		}
	}

	return res
}
