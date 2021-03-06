package virtual_security

import (
	"errors"
	"log"
	"reflect"
	"testing"
	"time"
)

type testMarginService struct {
	iMarginService
	toMarginOrder1              *marginOrder
	validation1                 error
	confirmContract1            error
	confirmContractCount        int
	holdExitOrderPositions1     error
	holdExitOrderPositionsCount int
	getMarginOrders1            []*marginOrder
	getMarginOrderByCode1       *marginOrder
	getMarginOrderByCode2       error
	saveMarginOrderHistory      []*marginOrder
	getMarginPositions1         []*marginPosition
	cancelAndRelease1           error
	cancelAndReleaseCount       int
}

func (t *testMarginService) toMarginOrder(*MarginOrderRequest, time.Time) *marginOrder {
	return t.toMarginOrder1
}
func (t *testMarginService) validation(*marginOrder, time.Time) error { return t.validation1 }
func (t *testMarginService) confirmContract(*marginOrder, *symbolPrice, time.Time) error {
	t.confirmContractCount++
	return t.confirmContract1
}
func (t *testMarginService) holdExitOrderPositions(*marginOrder) error {
	t.holdExitOrderPositionsCount++
	return t.holdExitOrderPositions1
}
func (t *testMarginService) getMarginOrders() []*marginOrder { return t.getMarginOrders1 }
func (t *testMarginService) getMarginOrderByCode(string) (*marginOrder, error) {
	return t.getMarginOrderByCode1, t.getMarginOrderByCode2
}
func (t *testMarginService) saveMarginOrder(order *marginOrder) {
	if t.saveMarginOrderHistory == nil {
		t.saveMarginOrderHistory = make([]*marginOrder, 0)
	}
	t.saveMarginOrderHistory = append(t.saveMarginOrderHistory, order)
}
func (t *testMarginService) removeMarginOrderByCode(string)        {}
func (t *testMarginService) getMarginPositions() []*marginPosition { return t.getMarginPositions1 }
func (t *testMarginService) removeMarginPositionByCode(string)     {}
func (t *testMarginService) cancelAndRelease(*marginOrder, time.Time) error {
	t.cancelAndReleaseCount++
	return t.cancelAndRelease1
}

func Test_marginService_newOrderCode(t *testing.T) {
	t.Parallel()
	want := "mor-1234"
	service := &marginService{uuidGenerator: &testUUIDGenerator{generator1: []string{"1234"}}}
	got := service.newOrderCode()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_marginService_newContractCode(t *testing.T) {
	t.Parallel()
	want := "mco-1234"
	service := &marginService{uuidGenerator: &testUUIDGenerator{generator1: []string{"1234"}}}
	got := service.newContractCode()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}
func Test_marginService_newPositionCode(t *testing.T) {
	t.Parallel()
	want := "mpo-1234"
	service := &marginService{uuidGenerator: &testUUIDGenerator{generator1: []string{"1234"}}}
	got := service.newPositionCode()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_marginService_toMarginOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 *MarginOrderRequest
		arg2 time.Time
		want *marginOrder
	}{
		{name: "?????????nil??????nil?????????",
			arg1: nil,
			arg2: time.Date(2021, 8, 17, 8, 0, 0, 0, time.Local),
			want: nil},
		{name: "?????????????????????????????????????????????????????????",
			arg1: &MarginOrderRequest{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionLO,
				SymbolCode:         "1111",
				Quantity:           100,
				LimitPrice:         1000,
				ExpiredAt:          time.Time{},
			},
			arg2: time.Date(2021, 8, 17, 8, 0, 0, 0, time.Local),
			want: &marginOrder{
				Code:               "mor-1234",
				OrderStatus:        OrderStatusInOrder,
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionLO,
				SymbolCode:         "1111",
				OrderQuantity:      100,
				ContractedQuantity: 0,
				CanceledQuantity:   0,
				LimitPrice:         1000,
				ExpiredAt:          time.Date(2021, 8, 17, 0, 0, 0, 0, time.Local),
				StopCondition:      nil,
				ExitPositionList:   nil,
				OrderedAt:          time.Date(2021, 8, 17, 8, 0, 0, 0, time.Local),
				Contracts:          []*Contract{},
			}},
		{name: "??????????????????????????????????????????????????????????????????",
			arg1: &MarginOrderRequest{
				TradeType:          TradeTypeExit,
				Side:               SideSell,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1111",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 8, 20, 11, 0, 0, 0, time.Local),
				StopCondition: &StockStopCondition{
					StopPrice:                  1000,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				},
				ExitPositionList: []ExitPosition{{PositionCode: "mpo-0000", Quantity: 100}},
			},
			arg2: time.Date(2021, 8, 17, 11, 0, 0, 0, time.Local),
			want: &marginOrder{
				Code:               "mor-1234",
				OrderStatus:        OrderStatusInOrder,
				TradeType:          TradeTypeExit,
				Side:               SideSell,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1111",
				OrderQuantity:      100,
				ContractedQuantity: 0,
				CanceledQuantity:   0,
				LimitPrice:         0,
				ExpiredAt:          time.Date(2021, 8, 20, 0, 0, 0, 0, time.Local),
				StopCondition: &StockStopCondition{
					StopPrice:                  1000,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				},
				ExitPositionList: []ExitPosition{{PositionCode: "mpo-0000", Quantity: 100}},
				OrderedAt:        time.Date(2021, 8, 17, 11, 0, 0, 0, time.Local),
				Contracts:        []*Contract{},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &marginService{uuidGenerator: &testUUIDGenerator{generator1: []string{"1234"}}}
			got := service.toMarginOrder(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginService_validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		isValidMarginOrder1 error
		want                error
	}{
		{name: "validationComponent???error????????????error?????????",
			isValidMarginOrder1: NilArgumentError,
			want:                NilArgumentError},
		{name: "validationComponent???nil????????????nil?????????",
			isValidMarginOrder1: nil,
			want:                nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &marginService{marginPositionStore: &testMarginPositionStore{}, validatorComponent: &testValidatorComponent{isValidMarginOrder1: test.isValidMarginOrder1}}
			got := service.validation(&marginOrder{}, time.Now())
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginService_entry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		service               *marginService
		arg1                  *marginOrder
		arg2                  *symbolPrice
		arg3                  time.Time
		want                  error
		wantPositionStoreSave []*marginPosition
		wantArg1              *marginOrder
	}{
		{name: "?????????nil???????????????",
			service:               &marginService{},
			arg1:                  nil,
			arg2:                  &symbolPrice{},
			want:                  NilArgumentError,
			wantPositionStoreSave: []*marginPosition{},
			wantArg1:              nil},
		{name: "?????????nil???????????????",
			service:               &marginService{},
			arg1:                  &marginOrder{},
			arg2:                  nil,
			want:                  NilArgumentError,
			wantPositionStoreSave: []*marginPosition{},
			wantArg1:              &marginOrder{}},
		{name: "?????????????????????????????????????????????????????????????????????????????????",
			service: &marginService{
				stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: false}}},
			arg1:                  &marginOrder{Code: "mor-01"},
			arg2:                  &symbolPrice{},
			want:                  nil,
			wantPositionStoreSave: []*marginPosition{},
			wantArg1:              &marginOrder{Code: "mor-01"}},
		{name: "??????????????????????????????????????????????????????????????????????????????????????????????????????",
			service: &marginService{
				uuidGenerator:          &testUUIDGenerator{generator1: []string{"01", "02", "03", "04", "05"}},
				stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}},
			arg1:                  &marginOrder{SymbolCode: "1234", Side: SideBuy, Code: "mor-01", OrderQuantity: 100},
			arg2:                  &symbolPrice{},
			want:                  nil,
			wantPositionStoreSave: []*marginPosition{{Code: "mpo-02", OrderCode: "mor-01", SymbolCode: "1234", Side: SideBuy, ContractedQuantity: 100, OwnedQuantity: 100, HoldQuantity: 0, Price: 1000, ContractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}},
			wantArg1:              &marginOrder{SymbolCode: "1234", OrderStatus: OrderStatusDone, Side: SideBuy, Code: "mor-01", OrderQuantity: 100, ContractedQuantity: 100, Contracts: []*Contract{{ContractCode: "mco-01", OrderCode: "mor-01", PositionCode: "mpo-02", Price: 1000, Quantity: 100, ContractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			marginOrderStore := &testMarginOrderStore{saveHistory: []*marginOrder{}}
			test.service.marginOrderStore = marginOrderStore
			marginPositionStore := &testMarginPositionStore{saveHistory: []*marginPosition{}}
			test.service.marginPositionStore = marginPositionStore
			got := test.service.entry(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantPositionStoreSave, marginPositionStore.saveHistory) ||
				!reflect.DeepEqual(test.wantArg1, test.arg1) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantPositionStoreSave, test.wantArg1,
					got, marginPositionStore.saveHistory, test.arg1)
			}
		})
	}
}

func Test_marginService_exit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		service       *marginService
		orderStore    *testMarginOrderStore
		positionStore *testMarginPositionStore
		arg1          *marginOrder
		arg2          *symbolPrice
		arg3          time.Time
		want          error
		wantArg1      *marginOrder
		wantPosition  *marginPosition
	}{
		{name: "order???nil???????????????",
			service:       &marginService{},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{},
			arg1:          nil,
			arg2:          &symbolPrice{},
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          NilArgumentError,
			wantArg1:      nil,
			wantPosition:  nil},
		{name: "symbolPrice???nil???????????????",
			service:       &marginService{},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{},
			arg1:          &marginOrder{},
			arg2:          nil,
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          NilArgumentError,
			wantArg1:      &marginOrder{},
			wantPosition:  nil},
		{name: "???????????????????????????????????????",
			service:       &marginService{stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: false}}},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{},
			arg1:          &marginOrder{},
			arg2:          &symbolPrice{},
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          nil,
			wantArg1:      &marginOrder{},
			wantPosition:  nil},
		{name: "???????????????????????????????????????????????????",
			service:       &marginService{stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{getByCode1: nil, getByCode2: NoDataError},
			arg1:          &marginOrder{ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			arg2:          &symbolPrice{},
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          NoDataError,
			wantArg1:      &marginOrder{ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			wantPosition:  nil},
		{name: "??????????????????????????????exit?????????????????????????????????",
			service:       &marginService{stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{getByCode1: &marginPosition{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 50}, getByCode2: nil},
			arg1:          &marginOrder{ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			arg2:          &symbolPrice{},
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          NotEnoughHoldQuantityError,
			wantArg1:      &marginOrder{ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			wantPosition:  &marginPosition{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 50}},
		{name: "??????????????????exit?????????????????????????????????",
			service: &marginService{
				stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}},
				uuidGenerator:          &testUUIDGenerator{generator1: []string{"01", "02", "03"}}},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{getByCode1: &marginPosition{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 100}, getByCode2: nil},
			arg1:          &marginOrder{Code: "mor-01", OrderQuantity: 100, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-01", HoldQuantity: 100}}},
			arg2:          &symbolPrice{},
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          nil,
			wantArg1:      &marginOrder{Code: "mor-01", OrderStatus: OrderStatusDone, OrderQuantity: 100, ContractedQuantity: 100, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-01", HoldQuantity: 100, ExitQuantity: 100}}, Contracts: []*Contract{{ContractCode: "mco-01", OrderCode: "mor-01", PositionCode: "mpo-01", Price: 1000, Quantity: 100, ContractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}},
			wantPosition:  &marginPosition{Code: "mpo-01", OwnedQuantity: 0, HoldQuantity: 0}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.service.marginOrderStore = test.orderStore
			test.service.marginPositionStore = test.positionStore
			got := test.service.exit(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantArg1, test.arg1) ||
				!reflect.DeepEqual(test.wantPosition, test.positionStore.getByCode1) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantArg1, test.wantPosition,
					got, test.arg1, test.positionStore.getByCode1)
			}
		})
	}
}

func Test_marginService_getMarginOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		marginService iMarginService
		want          []*marginOrder
	}{
		{name: "store????????????????????????",
			marginService: &marginService{marginOrderStore: &testMarginOrderStore{getAll1: []*marginOrder{}}},
			want:          []*marginOrder{}},
		{name: "store??????????????????????????????",
			marginService: &marginService{marginOrderStore: &testMarginOrderStore{getAll1: []*marginOrder{{Code: "mor-1"}, {Code: "mor-2"}, {Code: "mor-3"}}}},
			want:          []*marginOrder{{Code: "mor-1"}, {Code: "mor-2"}, {Code: "mor-3"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.marginService.getMarginOrders()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginService_getMarginOrderByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		marginService iMarginService
		arg           string
		want1         *marginOrder
		want2         error
	}{
		{name: "store?????????????????????????????????????????????",
			marginService: &marginService{marginOrderStore: &testMarginOrderStore{getByCode1: nil, getByCode2: NoDataError}},
			arg:           "sor-1",
			want1:         nil,
			want2:         NoDataError},
		{name: "store???order???????????????order?????????",
			marginService: &marginService{marginOrderStore: &testMarginOrderStore{getByCode1: &marginOrder{Code: "mor-1"}, getByCode2: nil}},
			arg:           "sor-1",
			want1:         &marginOrder{Code: "mor-1"},
			want2:         nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.marginService.getMarginOrderByCode(test.arg)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_marginService_saveMarginOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		store           *testMarginOrderStore
		arg             *marginOrder
		wantSaveHistory []*marginOrder
	}{
		{name: "??????????????????????????????store?????????", store: &testMarginOrderStore{saveHistory: []*marginOrder{}}, arg: &marginOrder{Code: "sor-1"}, wantSaveHistory: []*marginOrder{{Code: "sor-1"}}},
		{name: "?????????nil??????store?????????", store: &testMarginOrderStore{saveHistory: []*marginOrder{}}, arg: nil, wantSaveHistory: []*marginOrder{nil}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &marginService{marginOrderStore: test.store}
			service.saveMarginOrder(test.arg)
			if !reflect.DeepEqual(test.wantSaveHistory, test.store.saveHistory) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantSaveHistory, test.store.saveHistory)
			}
		})
	}
}

func Test_marginService_removeMarginOrderByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		arg                 string
		removeByCodeHistory []string
	}{
		{name: "?????????store???remove?????????",
			arg:                 "mor-1",
			removeByCodeHistory: []string{"mor-1"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &testMarginOrderStore{}
			service := &marginService{marginOrderStore: store}
			service.removeMarginOrderByCode(test.arg)
			if !reflect.DeepEqual(test.removeByCodeHistory, store.removeByCodeHistory) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.removeByCodeHistory, store.removeByCodeHistory)
			}
		})
	}
}

func Test_marginService_getMarginPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		service iMarginService
		want    []*marginPosition
	}{
		{name: "store????????????????????????????????????",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getAll1: []*marginPosition{}}},
			want:    []*marginPosition{}},
		{name: "store????????????????????????????????????????????????",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getAll1: []*marginPosition{{Code: "mpo-1"}, {Code: "mpo-2"}, {Code: "mpo-3"}}}},
			want:    []*marginPosition{{Code: "mpo-1"}, {Code: "mpo-2"}, {Code: "mpo-3"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.service.getMarginPositions()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginService_removeMarginPositionByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		arg                 string
		removeByCodeHistory []string
	}{
		{name: "?????????store???remove?????????",
			arg:                 "spo-1",
			removeByCodeHistory: []string{"spo-1"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &testMarginPositionStore{}
			service := &marginService{marginPositionStore: store}
			service.removeMarginPositionByCode(test.arg)
			log.Printf("%+v\n", store)
			if !reflect.DeepEqual(test.removeByCodeHistory, store.removeByCodeHistory) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.removeByCodeHistory, store.removeByCodeHistory)
			}
		})
	}
}

func Test_marginService_holdExitOrderPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		marginPositionStore *testMarginPositionStore
		arg1                *marginOrder
		want                error
		wantPosition        *marginPosition
		wantArg1            *marginOrder
	}{
		{name: "?????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil},
			arg1:                &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			want:                nil,
			wantPosition:        &marginPosition{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 100},
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-01", HoldQuantity: 100}}}},
		{name: "?????????nil???????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil},
			arg1:                nil,
			want:                NilArgumentError,
			wantPosition:        &marginPosition{OwnedQuantity: 100, HoldQuantity: 0},
			wantArg1:            nil},
		{name: "?????????Exit??????????????????????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil},
			arg1:                &marginOrder{TradeType: TradeTypeEntry, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			want:                InvalidTradeTypeError,
			wantPosition:        &marginPosition{OwnedQuantity: 100, HoldQuantity: 0},
			wantArg1:            &marginOrder{TradeType: TradeTypeEntry, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}}},
		{name: "Exit????????????????????????nil???????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil},
			arg1:                &marginOrder{TradeType: TradeTypeExit, ExitPositionList: nil},
			want:                InvalidExitPositionError,
			wantPosition:        &marginPosition{OwnedQuantity: 100, HoldQuantity: 0},
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, ExitPositionList: nil}},
		{name: "Exit????????????????????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil},
			arg1:                &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{}},
			want:                InvalidExitPositionError,
			wantPosition:        &marginPosition{OwnedQuantity: 100, HoldQuantity: 0},
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{}}},
		{name: "Exit??????????????????Store?????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: nil, getByCode2: NoDataError},
			arg1:                &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			want:                NoDataError,
			wantPosition:        nil,
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}}},
		{name: "Store?????????????????????????????????hold???????????????????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 50}, getByCode2: nil},
			arg1:                &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			want:                NotEnoughOwnedQuantityError,
			wantPosition:        &marginPosition{OwnedQuantity: 100, HoldQuantity: 50},
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &marginService{marginPositionStore: test.marginPositionStore}
			got := service.holdExitOrderPositions(test.arg1)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantPosition, test.marginPositionStore.getByCode1) ||
				!reflect.DeepEqual(test.wantArg1, test.arg1) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(), test.want, test.wantPosition, test.wantArg1, got, test.marginPositionStore.getByCode1, test.arg1)
			}
		})
	}
}

func Test_marginService_confirmContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		confirmMarginOrderContract *confirmContractResult
		arg1                       *marginOrder
		arg2                       *symbolPrice
		arg3                       time.Time
		want                       error
	}{
		{name: "?????????nil???????????????", arg2: &symbolPrice{}, want: NilArgumentError},
		{name: "?????????nil???????????????", arg1: &marginOrder{}, want: NilArgumentError},
		{name: "???????????????entry??????exit???????????????????????????", arg1: &marginOrder{TradeType: TradeTypeUnspecified}, arg2: &symbolPrice{}, want: InvalidTradeTypeError},
		{name: "???????????????entry??????entry???????????????",
			confirmMarginOrderContract: &confirmContractResult{isContracted: false},
			arg1:                       &marginOrder{TradeType: TradeTypeEntry},
			arg2:                       &symbolPrice{},
			want:                       nil},
		{name: "???????????????exit??????exit???????????????",
			confirmMarginOrderContract: &confirmContractResult{isContracted: false},
			arg1:                       &marginOrder{TradeType: TradeTypeExit},
			arg2:                       &symbolPrice{},
			want:                       nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &marginService{stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: test.confirmMarginOrderContract}}
			got := service.confirmContract(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginService_cancelAndRelease(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		marginPositionStore *testMarginPositionStore
		arg1                *marginOrder
		arg2                time.Time
		want                error
		wantArg1            *marginOrder
		wantPosition        *marginPosition
	}{
		{name: "?????????nil???????????????",
			marginPositionStore: &testMarginPositionStore{},
			arg1:                nil,
			arg2:                time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:                NilArgumentError,
			wantArg1:            nil,
			wantPosition:        nil},
		{name: "?????????????????????????????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{},
			arg1:                &marginOrder{OrderStatus: OrderStatusCanceled},
			arg2:                time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:                UncancellableOrderError,
			wantArg1:            &marginOrder{OrderStatus: OrderStatusCanceled},
			wantPosition:        nil},
		{name: "?????????Exit??????????????????????????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{},
			arg1:                &marginOrder{TradeType: TradeTypeEntry, OrderStatus: OrderStatusInOrder},
			arg2:                time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:                nil,
			wantArg1:            &marginOrder{TradeType: TradeTypeEntry, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local)},
			wantPosition:        nil},
		{name: "?????????Exit???????????????????????????????????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{},
			arg1:                &marginOrder{TradeType: TradeTypeExit, OrderStatus: OrderStatusInOrder, ExitPositionList: []ExitPosition{{PositionCode: "mpo-uuid-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 100}}},
			arg2:                time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:                nil,
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local), ExitPositionList: []ExitPosition{{PositionCode: "mpo-uuid-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 100}}},
			wantPosition:        nil},
		{name: "?????????Exit????????????????????????????????????????????????????????????????????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 70, HoldQuantity: 70}},
			arg1:                &marginOrder{TradeType: TradeTypeExit, OrderStatus: OrderStatusInOrder, ExitPositionList: []ExitPosition{{PositionCode: "mpo-uuid-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 30}}},
			arg2:                time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:                nil,
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local), ExitPositionList: []ExitPosition{{PositionCode: "mpo-uuid-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 30, ExitQuantity: 30}}},
			wantPosition:        &marginPosition{OwnedQuantity: 70, HoldQuantity: 0}},
		{name: "?????????Exit??????????????????????????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{getByCode2: NoDataError},
			arg1:                &marginOrder{TradeType: TradeTypeExit, OrderStatus: OrderStatusInOrder, ExitPositionList: []ExitPosition{{PositionCode: "mpo-uuid-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 30}}},
			arg2:                time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:                NoDataError,
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local), ExitPositionList: []ExitPosition{{PositionCode: "mpo-uuid-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 30}}},
			wantPosition:        nil},
		{name: "?????????Exit????????????????????????????????????????????????release????????????????????????????????????",
			marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 70, HoldQuantity: 0}},
			arg1:                &marginOrder{TradeType: TradeTypeExit, OrderStatus: OrderStatusInOrder, ExitPositionList: []ExitPosition{{PositionCode: "mpo-uuid-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 30}}},
			arg2:                time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:                NotEnoughHoldQuantityError,
			wantArg1:            &marginOrder{TradeType: TradeTypeExit, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local), ExitPositionList: []ExitPosition{{PositionCode: "mpo-uuid-01", Quantity: 100}}, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 30, ExitQuantity: 30}}},
			wantPosition:        &marginPosition{OwnedQuantity: 70, HoldQuantity: 0}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &marginService{marginPositionStore: test.marginPositionStore}
			got := service.cancelAndRelease(test.arg1, test.arg2)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantArg1, test.arg1) ||
				!reflect.DeepEqual(test.wantPosition, test.marginPositionStore.getByCode1) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantArg1, test.wantPosition,
					got, test.arg1, test.marginPositionStore.getByCode1)
			}
		})
	}
}
