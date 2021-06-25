package virtual_security

import "sync"

type StockService interface {
	NewOrderCode() string
	Entry(order *stockOrder, contractResult *confirmContractResult) error
	Exit(order *stockOrder, contractResult *confirmContractResult) error
	GetStockOrders() []*stockOrder
	GetStockOrderByCode(orderCode string) (*stockOrder, error)
	AddStockOrder(order *stockOrder) error
	RemoveStockOrderByCode(orderCode string)
	GetStockPositions() []*stockPosition
	RemoveStockPositionByCode(positionCode string)
}

type stockService struct {
	uuidGenerator      UUIDGenerator
	stockOrderStore    StockOrderStore
	stockPositionStore StockPositionStore
}

func (s *stockService) NewOrderCode() string {
	return "sor-" + s.uuidGenerator.Generate()
}

func (s *stockService) Entry(order *stockOrder, contractResult *confirmContractResult) error {
	contractCode := "con-" + s.uuidGenerator.Generate()
	positionCode := "spo-" + s.uuidGenerator.Generate()
	order.contract(&Contract{
		ContractCode: contractCode,
		OrderCode:    order.Code,
		PositionCode: positionCode,
		Price:        contractResult.price,
		Quantity:     order.OrderQuantity,
		ContractedAt: contractResult.contractedAt,
	})

	s.stockPositionStore.Add(&stockPosition{
		Code:               positionCode,
		OrderCode:          order.Code,
		SymbolCode:         order.SymbolCode,
		Side:               order.Side,
		ContractedQuantity: order.ContractedQuantity,
		OwnedQuantity:      order.ContractedQuantity,
		ContractedAt:       contractResult.contractedAt,
		mtx:                sync.Mutex{},
	})

	// storeに保存
	return s.stockOrderStore.Add(order)
}

func (s *stockService) Exit(order *stockOrder, contractResult *confirmContractResult) error {
	closePosition, err := s.stockPositionStore.GetByCode(order.ClosePositionCode)
	if err != nil {
		return err
	}

	// まずポジションを注文数拘束し、そのあとポジションを返済する
	if err := closePosition.hold(order.OrderQuantity); err != nil {
		return err
	}
	_ = closePosition.exit(order.OrderQuantity) // 直前でholdしていて確実にexitできるためerrは捨てられる

	// 注文に約定情報を追加
	contractCode := "con-" + s.uuidGenerator.Generate()
	order.contract(&Contract{
		ContractCode: contractCode,
		OrderCode:    order.Code,
		PositionCode: closePosition.Code,
		Price:        contractResult.price,
		Quantity:     order.OrderQuantity,
		ContractedAt: contractResult.contractedAt,
	})

	// storeに保存
	return s.stockOrderStore.Add(order)
}

func (s *stockService) GetStockOrders() []*stockOrder {
	return s.stockOrderStore.GetAll()
}

func (s *stockService) GetStockOrderByCode(orderCode string) (*stockOrder, error) {
	return s.stockOrderStore.GetByCode(orderCode)
}

func (s *stockService) AddStockOrder(order *stockOrder) error {
	return s.stockOrderStore.Add(order)
}

func (s *stockService) RemoveStockOrderByCode(orderCode string) {
	s.stockOrderStore.RemoveByCode(orderCode)
}

func (s *stockService) GetStockPositions() []*stockPosition {
	return s.stockPositionStore.GetAll()
}

func (s *stockService) RemoveStockPositionByCode(positionCode string) {
	s.stockPositionStore.RemoveByCode(positionCode)
}
