package virtual_security

import (
	"sync"
)

func newStockService(uuidGenerator iUUIDGenerator, stockOrderStore iStockOrderStore, stockPositionStore iStockPositionStore) iStockService {
	return &stockService{
		uuidGenerator:      uuidGenerator,
		stockOrderStore:    stockOrderStore,
		stockPositionStore: stockPositionStore,
	}
}

type iStockService interface {
	newOrderCode() string
	entry(order *stockOrder, contractResult *confirmContractResult) error
	exit(order *stockOrder, contractResult *confirmContractResult) error
	getStockOrders() []*stockOrder
	getStockOrderByCode(orderCode string) (*stockOrder, error)
	addStockOrder(order *stockOrder) error
	removeStockOrderByCode(orderCode string)
	getStockPositions() []*stockPosition
	removeStockPositionByCode(positionCode string)
}

type stockService struct {
	uuidGenerator      iUUIDGenerator
	stockOrderStore    iStockOrderStore
	stockPositionStore iStockPositionStore
}

func (s *stockService) newOrderCode() string {
	return "sor-" + s.uuidGenerator.generate()
}

func (s *stockService) entry(order *stockOrder, contractResult *confirmContractResult) error {
	contractCode := "con-" + s.uuidGenerator.generate()
	positionCode := "spo-" + s.uuidGenerator.generate()
	order.contract(&Contract{
		ContractCode: contractCode,
		OrderCode:    order.Code,
		PositionCode: positionCode,
		Price:        contractResult.price,
		Quantity:     order.OrderQuantity,
		ContractedAt: contractResult.contractedAt,
	})

	s.stockPositionStore.add(&stockPosition{
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

	// storeに保存
	return s.stockOrderStore.add(order)
}

func (s *stockService) exit(order *stockOrder, contractResult *confirmContractResult) error {
	positions, err := s.stockPositionStore.getBySymbolCode(order.SymbolCode)
	if err != nil {
		return err
	}

	// positionの保有数が今回返済したいポジションの数より多いかをチェック
	var totalQuantity float64
	for _, p := range positions {
		totalQuantity += p.orderableQuantity()
	}
	if totalQuantity < order.OrderQuantity {
		return NotEnoughOwnedQuantityError
	}

	// 古いポジションから順に返したい全量まで返せるだけ返していく
	q := order.OrderQuantity
	for _, p := range positions {
		orderQuantity := p.orderableQuantity()
		if q < orderQuantity {
			orderQuantity = q
		}
		if orderQuantity <= 0 {
			continue
		}
		q -= orderQuantity

		// まずポジションを注文数拘束し、そのあとポジションを返済する
		// TODO holdとexit時にエラーが出る可能性が高いのであれば、エラーをコントロールできるようにする
		_ = p.hold(orderQuantity) // エラーが出る可能性はあるが、エラーが出ても途中まで拘束しているポジションを開放できるわけでもないのでエラーは捨ててしまう
		_ = p.exit(orderQuantity) // 直前でholdしていて確実にexitできるためerrは捨てられる

		// 注文に約定情報を追加
		contractCode := "con-" + s.uuidGenerator.generate()
		order.contract(&Contract{
			ContractCode: contractCode,
			OrderCode:    order.Code,
			PositionCode: p.Code,
			Price:        contractResult.price,
			Quantity:     orderQuantity,
			ContractedAt: contractResult.contractedAt,
		})
	}

	// storeに保存
	return s.stockOrderStore.add(order)
}

func (s *stockService) getStockOrders() []*stockOrder {
	return s.stockOrderStore.getAll()
}

func (s *stockService) getStockOrderByCode(orderCode string) (*stockOrder, error) {
	return s.stockOrderStore.getByCode(orderCode)
}

func (s *stockService) addStockOrder(order *stockOrder) error {
	return s.stockOrderStore.add(order)
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
