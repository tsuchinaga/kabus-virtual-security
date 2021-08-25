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
		{name: "引数がnilならnilを返す",
			arg1: nil,
			arg2: time.Date(2021, 8, 17, 8, 0, 0, 0, time.Local),
			want: nil},
		{name: "有効期限がなければ当日が有効期限になる",
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
		{name: "有効期限があれば指定された有効期限を設定する",
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
		{name: "validationComponentがerrorを返せばerrorを返す",
			isValidMarginOrder1: NilArgumentError,
			want:                NilArgumentError},
		{name: "validationComponentがnilを返せばnilを返す",
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
	}{
		{name: "注文がnilならエラー",
			service:               &marginService{},
			arg1:                  nil,
			arg2:                  &symbolPrice{},
			want:                  NilArgumentError,
			wantPositionStoreSave: []*marginPosition{}},
		{name: "価格がnilならエラー",
			service:               &marginService{},
			arg1:                  &marginOrder{},
			arg2:                  nil,
			want:                  NilArgumentError,
			wantPositionStoreSave: []*marginPosition{}},
		{name: "約定チェックで約定しなければ約定確認後の注文を保存する",
			service: &marginService{
				stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: false}}},
			arg1:                  &marginOrder{Code: "mor-01"},
			arg2:                  &symbolPrice{},
			want:                  nil,
			wantPositionStoreSave: []*marginPosition{}},
		{name: "約定チェックで約定すれば約定確認後の注文と新しいポジションを保存する",
			service: &marginService{
				uuidGenerator:          &testUUIDGenerator{generator1: []string{"01", "02", "03", "04", "05"}},
				stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}},
			arg1:                  &marginOrder{SymbolCode: "1234", Side: SideBuy, Code: "mor-01", OrderQuantity: 100},
			arg2:                  &symbolPrice{},
			want:                  nil,
			wantPositionStoreSave: []*marginPosition{{Code: "mpo-02", OrderCode: "mor-01", SymbolCode: "1234", Side: SideBuy, ContractedQuantity: 100, OwnedQuantity: 100, HoldQuantity: 0, Price: 1000, ContractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}},
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
				!reflect.DeepEqual(test.wantPositionStoreSave, marginPositionStore.saveHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(),
					test.want, test.wantPositionStoreSave, got, marginPositionStore.saveHistory)
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
	}{
		{name: "orderがnilならエラー",
			service:    &marginService{},
			orderStore: &testMarginOrderStore{saveHistory: []*marginOrder{}},
			arg1:       nil,
			arg2:       &symbolPrice{},
			arg3:       time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:       NilArgumentError},
		{name: "symbolPriceがnilならエラー",
			service:    &marginService{},
			orderStore: &testMarginOrderStore{saveHistory: []*marginOrder{}},
			arg1:       &marginOrder{},
			arg2:       nil,
			arg3:       time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:       NilArgumentError},
		{name: "約定しなければそのままstoreに保存して終了",
			service:    &marginService{stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: false}}},
			orderStore: &testMarginOrderStore{saveHistory: []*marginOrder{}},
			arg1:       &marginOrder{},
			arg2:       &symbolPrice{},
			arg3:       time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:       nil},
		{name: "指定したポジションがなければエラー",
			service:       &marginService{stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{getByCode1: nil, getByCode2: NoDataError},
			arg1:          &marginOrder{ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			arg2:          &symbolPrice{},
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          NoDataError},
		{name: "指定したポジションがexitできない状態ならエラー",
			service:       &marginService{stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}}},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{getByCode1: &marginPosition{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 50}, getByCode2: nil},
			arg1:          &marginOrder{ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			arg2:          &symbolPrice{},
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          NotEnoughHoldQuantityError},
		{name: "ポジションをexitし、約定状態を保存する",
			service: &marginService{
				stockContractComponent: &testStockContractComponent{confirmMarginOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local)}},
				uuidGenerator:          &testUUIDGenerator{generator1: []string{"01", "02", "03"}}},
			orderStore:    &testMarginOrderStore{saveHistory: []*marginOrder{}},
			positionStore: &testMarginPositionStore{getByCode1: &marginPosition{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 100}, getByCode2: nil},
			arg1:          &marginOrder{Code: "mor-01", OrderQuantity: 100, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			arg2:          &symbolPrice{},
			arg3:          time.Date(2021, 8, 20, 14, 0, 0, 0, time.Local),
			want:          nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.service.marginOrderStore = test.orderStore
			test.service.marginPositionStore = test.positionStore
			got := test.service.exit(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
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
		{name: "storeの結果が空なら空",
			marginService: &marginService{marginOrderStore: &testMarginOrderStore{getAll1: []*marginOrder{}}},
			want:          []*marginOrder{}},
		{name: "storeの結果をそのまま返す",
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
		{name: "storeがエラーを返したらエラーを返す",
			marginService: &marginService{marginOrderStore: &testMarginOrderStore{getByCode1: nil, getByCode2: NoDataError}},
			arg:           "sor-1",
			want1:         nil,
			want2:         NoDataError},
		{name: "storeがorderを返したらorderを返す",
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
		{name: "引数が有効な注文ならstoreに渡す", store: &testMarginOrderStore{saveHistory: []*marginOrder{}}, arg: &marginOrder{Code: "sor-1"}, wantSaveHistory: []*marginOrder{{Code: "sor-1"}}},
		{name: "引数がnilでもstoreに渡す", store: &testMarginOrderStore{saveHistory: []*marginOrder{}}, arg: nil, wantSaveHistory: []*marginOrder{nil}},
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
		{name: "引数をstoreのremoveに渡す",
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
		{name: "storeが空配列を返したら彼配列",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getAll1: []*marginPosition{}}},
			want:    []*marginPosition{}},
		{name: "storeが複数要素を返したラそのまま返す",
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
		{name: "引数をstoreのremoveに渡す",
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
		name    string
		service *marginService
		arg1    *marginOrder
		want    error
	}{
		{name: "問題なければエラーなし",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil}},
			arg1:    &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			want:    nil},
		{name: "引数がnilならエラー",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil}},
			arg1:    nil,
			want:    NilArgumentError},
		{name: "注文がExit注文でないならエラー",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil}},
			arg1:    &marginOrder{TradeType: TradeTypeEntry, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			want:    InvalidTradeTypeError},
		{name: "Exitポジション一覧がnilならエラー",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil}},
			arg1:    &marginOrder{TradeType: TradeTypeExit, ExitPositionList: nil},
			want:    InvalidExitPositionError},
		{name: "Exitポジション一覧が空配列ならエラー",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 0}, getByCode2: nil}},
			arg1:    &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{}},
			want:    InvalidExitPositionError},
		{name: "ExitポジションがStoreから取れなければエラー",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getByCode1: nil, getByCode2: NoDataError}},
			arg1:    &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			want:    NoDataError},
		{name: "Storeから取れたポジションをhold出来なければエラー",
			service: &marginService{marginPositionStore: &testMarginPositionStore{getByCode1: &marginPosition{OwnedQuantity: 100, HoldQuantity: 50}, getByCode2: nil}},
			arg1:    &marginOrder{TradeType: TradeTypeExit, ExitPositionList: []ExitPosition{{PositionCode: "mpo-01", Quantity: 100}}},
			want:    NotEnoughOwnedQuantityError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.service.holdExitOrderPositions(test.arg1)
			if !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
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
		{name: "注文がnilならエラー", arg2: &symbolPrice{}, want: NilArgumentError},
		{name: "価格がnilならエラー", arg1: &marginOrder{}, want: NilArgumentError},
		{name: "取引区分がentryでもexitでもなければエラー", arg1: &marginOrder{TradeType: TradeTypeUnspecified}, arg2: &symbolPrice{}, want: InvalidTradeTypeError},
		{name: "取引区分がentryならentryが叩かれる",
			confirmMarginOrderContract: &confirmContractResult{isContracted: false},
			arg1:                       &marginOrder{TradeType: TradeTypeEntry},
			arg2:                       &symbolPrice{},
			want:                       nil},
		{name: "取引区分がexitならexitが叩かれる",
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
