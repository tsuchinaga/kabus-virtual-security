package virtual_security

import (
	"errors"
	"log"
	"reflect"
	"testing"
	"time"
)

type testStockService struct {
	iStockService
	newOrderCode1                    []string
	newOrderCodeCount                int
	confirmContract1                 error
	confirmContractCount             int
	getStockOrders1                  []*stockOrder
	getStockOrderByCode1             *stockOrder
	getStockOrderByCode2             error
	getStockOrderByCodeHistory       []string
	removeStockOrderByCodeHistory    []string
	getStockPositions1               []*stockPosition
	removeStockPositionByCodeHistory []string
	addStockOrder1                   error
	addStockOrderHistory             []*stockOrder
	toStockOrder1                    *stockOrder
	holdSellOrderPositions1          error
	validation1                      error
	cancelAndRelease1                error
}

func (t *testStockService) saveStockOrder(order *stockOrder) {
	t.addStockOrderHistory = append(t.addStockOrderHistory, order)
}

func (t *testStockService) newOrderCode() string {
	defer func() { t.newOrderCodeCount++ }()
	return t.newOrderCode1[t.newOrderCodeCount%len(t.newOrderCode1)]
}

func (t *testStockService) confirmContract(*stockOrder, *symbolPrice, time.Time) error {
	t.confirmContractCount++
	return t.confirmContract1
}

func (t *testStockService) getStockOrders() []*stockOrder {
	return t.getStockOrders1
}

func (t *testStockService) getStockOrderByCode(orderCode string) (*stockOrder, error) {
	t.getStockOrderByCodeHistory = append(t.getStockOrderByCodeHistory, orderCode)
	return t.getStockOrderByCode1, t.getStockOrderByCode2
}

func (t *testStockService) removeStockOrderByCode(orderCode string) {
	t.removeStockOrderByCodeHistory = append(t.removeStockOrderByCodeHistory, orderCode)
}

func (t *testStockService) getStockPositions() []*stockPosition {
	return t.getStockPositions1
}

func (t *testStockService) removeStockPositionByCode(positionCode string) {
	t.removeStockPositionByCodeHistory = append(t.removeStockPositionByCodeHistory, positionCode)
}

func (t *testStockService) toStockOrder(*StockOrderRequest, time.Time) *stockOrder {
	return t.toStockOrder1
}

func (t *testStockService) holdSellOrderPositions(*stockOrder) error {
	return t.holdSellOrderPositions1
}

func (t *testStockService) validation(*stockOrder, time.Time) error {
	return t.validation1
}

func (t *testStockService) cancelAndRelease(*stockOrder, time.Time) error {
	return t.cancelAndRelease1
}

func Test_stockService_entry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                  string
		stockService          *stockService
		arg1                  *stockOrder
		arg2                  *symbolPrice
		arg3                  time.Time
		want                  error
		wantArg1              *stockOrder
		wantPositionStoreSave []*stockPosition
	}{
		{name: "引数1がnilならエラー",
			stockService:          &stockService{},
			arg1:                  nil,
			arg2:                  &symbolPrice{},
			want:                  NilArgumentError,
			wantArg1:              nil,
			wantPositionStoreSave: nil},
		{name: "引数2がnilならエラー",
			stockService:          &stockService{},
			arg1:                  &stockOrder{},
			arg2:                  nil,
			want:                  NilArgumentError,
			wantArg1:              &stockOrder{},
			wantPositionStoreSave: nil},
		{name: "約定チェックで約定していなかったら何もしない",
			stockService: &stockService{
				stockContractComponent: &testStockContractComponent{confirmStockOrderContract1: &confirmContractResult{isContracted: false}}},
			arg1:                  &stockOrder{},
			arg2:                  &symbolPrice{},
			want:                  nil,
			wantArg1:              &stockOrder{},
			wantPositionStoreSave: nil},
		{name: "それぞれコードを生成し、注文、ポジションをstoreに保存する",
			stockService: &stockService{
				stockContractComponent: &testStockContractComponent{confirmStockOrderContract1: &confirmContractResult{
					isContracted: true,
					price:        1000,
					contractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local)}},
				uuidGenerator: &testUUIDGenerator{generator1: []string{"uuid-1", "uuid-2", "uuid-3"}}},
			arg1: &stockOrder{
				Code:               "sor-1",
				SymbolCode:         "1234",
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				OrderQuantity:      100,
				OrderedAt:          time.Date(2021, 6, 21, 10, 0, 0, 0, time.Local),
				ConfirmingCount:    1,
			},
			arg2: &symbolPrice{},
			want: nil,
			wantArg1: &stockOrder{
				Code:               "sor-1",
				OrderStatus:        OrderStatusDone,
				SymbolCode:         "1234",
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				OrderQuantity:      100,
				ContractedQuantity: 100,
				OrderedAt:          time.Date(2021, 6, 21, 10, 0, 0, 0, time.Local),
				Contracts:          []*Contract{{ContractCode: "sco-uuid-1", OrderCode: "sor-1", PositionCode: "spo-uuid-2", Price: 1000, Quantity: 100, ContractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local)}},
				ConfirmingCount:    1,
			},
			wantPositionStoreSave: []*stockPosition{
				{
					Code:               "spo-uuid-2",
					OrderCode:          "sor-1",
					SymbolCode:         "1234",
					Side:               SideBuy,
					ContractedQuantity: 100,
					OwnedQuantity:      100,
					HoldQuantity:       0,
					Price:              1000,
					ContractedAt:       time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
				},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			stockOrderStore := &testStockOrderStore{}
			stockPositionStore := &testStockPositionStore{}
			test.stockService.stockOrderStore = stockOrderStore
			test.stockService.stockPositionStore = stockPositionStore

			got := test.stockService.entry(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantArg1, test.arg1) ||
				!reflect.DeepEqual(test.wantPositionStoreSave, stockPositionStore.saveHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantArg1, test.wantPositionStoreSave,
					got, test.arg1, stockPositionStore.saveHistory)
			}
		})
	}
}

func Test_stockService_exit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		stockService       *stockService
		stockPositionStore *testStockPositionStore
		arg1               *stockOrder
		arg2               *symbolPrice
		arg3               time.Time
		want               error
		wantArg1           *stockOrder
		wantPosition       *stockPosition
	}{
		{name: "引数1がnilならエラー",
			stockService:       &stockService{},
			stockPositionStore: &testStockPositionStore{},
			arg1:               nil,
			arg2:               &symbolPrice{},
			want:               NilArgumentError,
			wantArg1:           nil,
			wantPosition:       nil},
		{name: "引数2がnilならエラー",
			stockService:       &stockService{},
			stockPositionStore: &testStockPositionStore{},
			arg1:               &stockOrder{},
			arg2:               nil,
			want:               NilArgumentError,
			wantArg1:           &stockOrder{},
			wantPosition:       nil},
		{name: "約定チェックで約定していなかったら何もしない",
			stockService:       &stockService{stockContractComponent: &testStockContractComponent{confirmStockOrderContract1: &confirmContractResult{isContracted: false}}},
			stockPositionStore: &testStockPositionStore{},
			arg1:               &stockOrder{},
			arg2:               &symbolPrice{},
			want:               nil,
			wantArg1:           &stockOrder{},
			wantPosition:       nil},
		{name: "注文がholdしていたpositionが見つからなければそのポジションの処理をスキップする",
			stockService:       &stockService{stockContractComponent: &testStockContractComponent{confirmStockOrderContract1: &confirmContractResult{isContracted: true}}},
			stockPositionStore: &testStockPositionStore{getByCode1: nil, getByCode2: NoDataError},
			arg1:               &stockOrder{Code: "sor-1", OrderQuantity: 400, HoldPositions: []*HoldPosition{{PositionCode: "spo-0", HoldQuantity: 400}}},
			arg2:               &symbolPrice{},
			want:               nil,
			wantArg1:           &stockOrder{Code: "sor-1", OrderQuantity: 400, HoldPositions: []*HoldPosition{{PositionCode: "spo-0", HoldQuantity: 400}}},
			wantPosition:       nil},
		{name: "注文がholdしていたpositionをexitする",
			stockService: &stockService{
				uuidGenerator:          &testUUIDGenerator{generator1: []string{"uuid-1", "uuid-2", "uuid-3", "uuid-4", "uuid-5"}},
				stockContractComponent: &testStockContractComponent{confirmStockOrderContract1: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local)}}},
			stockPositionStore: &testStockPositionStore{getByCode1: &stockPosition{Code: "spo-0", OwnedQuantity: 1000, HoldQuantity: 1000}},
			arg1:               &stockOrder{Code: "sor-1", OrderQuantity: 400, HoldPositions: []*HoldPosition{{PositionCode: "spo-0", HoldQuantity: 400}}},
			arg2:               &symbolPrice{},
			want:               nil,
			wantArg1: &stockOrder{
				Code:               "sor-1",
				OrderStatus:        OrderStatusDone,
				OrderQuantity:      400,
				ContractedQuantity: 400,
				Contracts: []*Contract{{
					ContractCode: "sco-uuid-1",
					OrderCode:    "sor-1",
					PositionCode: "spo-0",
					Price:        1000,
					Quantity:     400,
					ContractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
				}},
				HoldPositions: []*HoldPosition{{PositionCode: "spo-0", HoldQuantity: 400, ExitQuantity: 400}}},
			wantPosition: &stockPosition{Code: "spo-0", OwnedQuantity: 600, HoldQuantity: 600}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.stockService.stockPositionStore = test.stockPositionStore
			got := test.stockService.exit(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantArg1, test.arg1) ||
				!reflect.DeepEqual(test.wantPosition, test.stockPositionStore.getByCode1) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantArg1, test.wantPosition,
					got, test.arg1, test.stockPositionStore.getByCode1)
			}
		})
	}
}

func Test_stockService_NewOrderCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		uuidGenerator iUUIDGenerator
		want          string
	}{
		{name: "uuidの結果に接頭辞を付けて返す", uuidGenerator: &testUUIDGenerator{generator1: []string{"uuid-1", "uuid-2", "uuid-3"}}, want: "sor-uuid-1"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{uuidGenerator: test.uuidGenerator}
			got := service.newOrderCode()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockService_GetStockOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		stockService iStockService
		want         []*stockOrder
	}{
		{name: "storeの結果が空なら空",
			stockService: &stockService{stockOrderStore: &testStockOrderStore{getAll1: []*stockOrder{}}},
			want:         []*stockOrder{}},
		{name: "storeの結果をそのまま返す",
			stockService: &stockService{stockOrderStore: &testStockOrderStore{getAll1: []*stockOrder{{Code: "sor-1"}, {Code: "sor-2"}, {Code: "sor-3"}}}},
			want:         []*stockOrder{{Code: "sor-1"}, {Code: "sor-2"}, {Code: "sor-3"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockService.getStockOrders()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockService_GetStockOrderByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		stockService iStockService
		arg          string
		want1        *stockOrder
		want2        error
	}{
		{name: "storeがエラーを返したらエラーを返す",
			stockService: &stockService{stockOrderStore: &testStockOrderStore{getByCode1: nil, getByCode2: NoDataError}},
			arg:          "sor-1",
			want1:        nil,
			want2:        NoDataError},
		{name: "storeがorderを返したらorderを返す",
			stockService: &stockService{stockOrderStore: &testStockOrderStore{getByCode1: &stockOrder{Code: "sor-1"}, getByCode2: nil}},
			arg:          "sor-1",
			want1:        &stockOrder{Code: "sor-1"},
			want2:        nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.stockService.getStockOrderByCode(test.arg)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_stockService_RemoveStockOrderByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                string
		arg                 string
		removeByCodeHistory []string
	}{
		{name: "引数をstoreのremoveに渡す",
			arg:                 "sor-1",
			removeByCodeHistory: []string{"sor-1"}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			store := &testStockOrderStore{}
			service := &stockService{stockOrderStore: store}
			service.removeStockOrderByCode(test.arg)
			if !reflect.DeepEqual(test.removeByCodeHistory, store.removeByCodeHistory) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.removeByCodeHistory, store.removeByCodeHistory)
			}
		})
	}
}

func Test_stockService_GetStockPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		service iStockService
		want    []*stockPosition
	}{
		{name: "storeが空配列を返したら彼配列",
			service: &stockService{stockPositionStore: &testStockPositionStore{getAll1: []*stockPosition{}}},
			want:    []*stockPosition{}},
		{name: "storeが複数要素を返したラそのまま返す",
			service: &stockService{stockPositionStore: &testStockPositionStore{getAll1: []*stockPosition{{Code: "spo-1"}, {Code: "spo-2"}, {Code: "spo-3"}}}},
			want:    []*stockPosition{{Code: "spo-1"}, {Code: "spo-2"}, {Code: "spo-3"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.service.getStockPositions()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockService_RemoveStockPositionByCode(t *testing.T) {
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
			store := &testStockPositionStore{}
			service := &stockService{stockPositionStore: store}
			service.removeStockPositionByCode(test.arg)
			log.Printf("%+v\n", store)
			if !reflect.DeepEqual(test.removeByCodeHistory, store.removeByCodeHistory) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.removeByCodeHistory, store.removeByCodeHistory)
			}
		})
	}
}

func Test_stockService_saveStockOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		store           *testStockOrderStore
		arg             *stockOrder
		wantSaveHistory []*stockOrder
	}{
		{name: "引数が有効な注文ならstoreに渡す", store: &testStockOrderStore{saveHistory: []*stockOrder{}}, arg: &stockOrder{Code: "sor-1"}, wantSaveHistory: []*stockOrder{{Code: "sor-1"}}},
		{name: "引数がnilでもstoreに渡す", store: &testStockOrderStore{saveHistory: []*stockOrder{}}, arg: nil, wantSaveHistory: []*stockOrder{nil}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{stockOrderStore: test.store}
			service.saveStockOrder(test.arg)
			if !reflect.DeepEqual(test.wantSaveHistory, test.store.saveHistory) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantSaveHistory, test.store.saveHistory)
			}
		})
	}
}

func Test_newStockService(t *testing.T) {
	t.Parallel()
	uuid := &testUUIDGenerator{}
	stockOrderStore := &testStockOrderStore{}
	stockPositionStore := &testStockPositionStore{}
	stockContractComponent := &testStockContractComponent{}
	validatorComponent := &testValidatorComponent{}
	want := &stockService{uuidGenerator: uuid, stockOrderStore: stockOrderStore, stockPositionStore: stockPositionStore, stockContractComponent: stockContractComponent, validatorComponent: validatorComponent}
	got := newStockService(uuid, stockOrderStore, stockPositionStore, validatorComponent, stockContractComponent)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_stockService_newContractCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		uuidGenerator iUUIDGenerator
		want          string
	}{
		{name: "uuidの結果に接頭辞を付けて返す", uuidGenerator: &testUUIDGenerator{generator1: []string{"uuid-1", "uuid-2", "uuid-3"}}, want: "sco-uuid-1"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{uuidGenerator: test.uuidGenerator}
			got := service.newContractCode()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockService_newPositionCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		uuidGenerator iUUIDGenerator
		want          string
	}{
		{name: "uuidの結果に接頭辞を付けて返す", uuidGenerator: &testUUIDGenerator{generator1: []string{"uuid-1", "uuid-2", "uuid-3"}}, want: "spo-uuid-1"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{uuidGenerator: test.uuidGenerator}
			got := service.newPositionCode()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockService_toStockOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		service *stockService
		arg1    *StockOrderRequest
		arg2    time.Time
		want    *stockOrder
	}{
		{name: "nilを与えたらnilが返される", service: &stockService{}, arg1: nil, arg2: time.Time{}, want: nil},
		{name: "有効期限がゼロ値なら当日の年月日が入る",
			service: &stockService{uuidGenerator: &testUUIDGenerator{generator1: []string{"1", "2", "3"}}},
			arg1: &StockOrderRequest{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           1000,
				LimitPrice:         0,
				ExpiredAt:          time.Time{},
				StopCondition:      nil,
			},
			arg2: time.Date(2021, 7, 20, 10, 0, 0, 0, time.Local),
			want: &stockOrder{
				Code:               "sor-1",
				OrderStatus:        OrderStatusInOrder,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      1000,
				ContractedQuantity: 0,
				CanceledQuantity:   0,
				LimitPrice:         0,
				ExpiredAt:          time.Date(2021, 7, 20, 0, 0, 0, 0, time.Local),
				StopCondition:      nil,
				OrderedAt:          time.Date(2021, 7, 20, 10, 0, 0, 0, time.Local),
				CanceledAt:         time.Time{},
				Contracts:          []*Contract{},
				ConfirmingCount:    0,
				Message:            "",
			}},
		{name: "有効期限がゼロ値でないなら、指定した有効期限の年月日が入る",
			service: &stockService{uuidGenerator: &testUUIDGenerator{generator1: []string{"1", "2", "3"}}},
			arg1: &StockOrderRequest{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           1000,
				LimitPrice:         0,
				ExpiredAt:          time.Date(2021, 7, 22, 10, 0, 0, 0, time.Local),
				StopCondition:      nil,
			},
			arg2: time.Date(2021, 7, 20, 10, 0, 0, 0, time.Local),
			want: &stockOrder{
				Code:               "sor-1",
				OrderStatus:        OrderStatusInOrder,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      1000,
				ContractedQuantity: 0,
				CanceledQuantity:   0,
				LimitPrice:         0,
				ExpiredAt:          time.Date(2021, 7, 22, 0, 0, 0, 0, time.Local),
				StopCondition:      nil,
				OrderedAt:          time.Date(2021, 7, 20, 10, 0, 0, 0, time.Local),
				CanceledAt:         time.Time{},
				Contracts:          []*Contract{},
				ConfirmingCount:    0,
				Message:            "",
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.service.toStockOrder(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockService_holdSellOrderPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		getAll        []*stockPosition
		arg           *stockOrder
		want          error
		wantPositions []*stockPosition
		wantArg       *stockOrder
	}{
		{name: "引数がnilならエラー",
			getAll:        []*stockPosition{},
			arg:           nil,
			want:          NilArgumentError,
			wantPositions: []*stockPosition{},
			wantArg:       nil},
		{name: "position全てで数量が足りなければエラー",
			getAll:        []*stockPosition{{OwnedQuantity: 50}, {OwnedQuantity: 30}, {OwnedQuantity: 10}},
			arg:           &stockOrder{OrderQuantity: 100},
			want:          NotEnoughOwnedQuantityError,
			wantPositions: []*stockPosition{{OwnedQuantity: 50}, {OwnedQuantity: 30}, {OwnedQuantity: 10}},
			wantArg:       &stockOrder{OrderQuantity: 100}},
		{name: "数量が足りればholdする",
			getAll:        []*stockPosition{{Code: "spo-uuid-01", OwnedQuantity: 50}, {Code: "spo-uuid-02", OwnedQuantity: 30}, {Code: "spo-uuid-03", OwnedQuantity: 30}},
			arg:           &stockOrder{OrderQuantity: 100},
			want:          nil,
			wantPositions: []*stockPosition{{Code: "spo-uuid-01", OwnedQuantity: 50, HoldQuantity: 50}, {Code: "spo-uuid-02", OwnedQuantity: 30, HoldQuantity: 30}, {Code: "spo-uuid-03", OwnedQuantity: 30, HoldQuantity: 20}},
			wantArg:       &stockOrder{OrderQuantity: 100, HoldPositions: []*HoldPosition{{PositionCode: "spo-uuid-01", HoldQuantity: 50}, {PositionCode: "spo-uuid-02", HoldQuantity: 30}, {PositionCode: "spo-uuid-03", HoldQuantity: 20}}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{stockPositionStore: &testStockPositionStore{getAll1: test.getAll}}
			got := service.holdSellOrderPositions(test.arg)
			if !errors.Is(got, test.want) || !reflect.DeepEqual(test.getAll, test.wantPositions) || !reflect.DeepEqual(test.wantArg, test.arg) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(), test.want, test.wantPositions, test.wantArg, got, test.getAll, test.arg)
			}
		})
	}
}

func Test_stockService_validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		getAll            []*stockPosition
		isValidStockOrder error
		want              error
	}{
		{name: "errorを返されたらerrorを返す",
			getAll:            []*stockPosition{},
			isValidStockOrder: NilArgumentError,
			want:              NilArgumentError},
		{name: "nilを返されたらnilを返す",
			getAll:            []*stockPosition{},
			isValidStockOrder: nil,
			want:              nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{stockPositionStore: &testStockPositionStore{getAll1: test.getAll}, validatorComponent: &testValidatorComponent{isValidStockOrder1: test.isValidStockOrder}}
			got := service.validation(nil, time.Time{})
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockService_confirmContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                      string
		confirmStockOrderContract *confirmContractResult
		arg1                      *stockOrder
		arg2                      *symbolPrice
		arg3                      time.Time
		want                      error
	}{
		{name: "注文がnilならエラー", arg1: nil, arg2: &symbolPrice{}, want: NilArgumentError},
		{name: "価格がnilならエラー", arg1: &stockOrder{}, arg2: nil, want: NilArgumentError},
		{name: "売買方向が買いでも売りでもなければエラー", arg1: &stockOrder{Side: SideUnspecified}, arg2: &symbolPrice{}, want: InvalidSideError},
		{name: "売買方向が買いならentry",
			confirmStockOrderContract: &confirmContractResult{isContracted: false},
			arg1:                      &stockOrder{Side: SideBuy},
			arg2:                      &symbolPrice{},
			want:                      nil},
		{name: "売買方向が売りならexit",
			confirmStockOrderContract: &confirmContractResult{isContracted: false},
			arg1:                      &stockOrder{Side: SideSell},
			arg2:                      &symbolPrice{},
			want:                      nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{stockContractComponent: &testStockContractComponent{confirmStockOrderContract1: test.confirmStockOrderContract}}
			got := service.confirmContract(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockService_cancelAndRelease(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		stockPositionStore *testStockPositionStore
		arg1               *stockOrder
		arg2               time.Time
		want               error
		wantArg1           *stockOrder
		wantPosition       *stockPosition
	}{
		{name: "引数がnilならエラー",
			stockPositionStore: &testStockPositionStore{},
			arg1:               nil,
			arg2:               time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:               NilArgumentError,
			wantArg1:           nil,
			wantPosition:       nil},
		{name: "キャンセル可能な状態じゃなければエラー",
			stockPositionStore: &testStockPositionStore{},
			arg1:               &stockOrder{OrderStatus: OrderStatusCanceled},
			arg2:               time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:               UncancellableOrderError,
			wantArg1:           &stockOrder{OrderStatus: OrderStatusCanceled},
			wantPosition:       nil},
		{name: "注文がExitでなければ注文の状態だけ更新して終了",
			stockPositionStore: &testStockPositionStore{},
			arg1:               &stockOrder{Side: SideBuy, OrderStatus: OrderStatusInOrder},
			arg2:               time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:               nil,
			wantArg1:           &stockOrder{Side: SideBuy, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local)},
			wantPosition:       nil},
		{name: "注文がExitでも拘束中のポジションがなければ何もしない",
			stockPositionStore: &testStockPositionStore{},
			arg1:               &stockOrder{Side: SideSell, OrderStatus: OrderStatusInOrder, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 100}}},
			arg2:               time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:               nil,
			wantArg1:           &stockOrder{Side: SideSell, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local), HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 100}}},
			wantPosition:       nil},
		{name: "注文がExitで拘束中のポジションがあれば解放し、ポジションの拘束数を解放する",
			stockPositionStore: &testStockPositionStore{getByCode1: &stockPosition{OwnedQuantity: 70, HoldQuantity: 70}},
			arg1:               &stockOrder{Side: SideSell, OrderStatus: OrderStatusInOrder, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 30}}},
			arg2:               time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:               nil,
			wantArg1:           &stockOrder{Side: SideSell, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local), HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 30, ExitQuantity: 30}}},
			wantPosition:       &stockPosition{OwnedQuantity: 70, HoldQuantity: 0}},
		{name: "注文がExitで拘束中のポジションがなければエラー",
			stockPositionStore: &testStockPositionStore{getByCode2: NoDataError},
			arg1:               &stockOrder{Side: SideSell, OrderStatus: OrderStatusInOrder, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 30}}},
			arg2:               time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:               NoDataError,
			wantArg1:           &stockOrder{Side: SideSell, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local), HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 30}}},
			wantPosition:       nil},
		{name: "注文がExitで拘束中のポジションがあっても、releaseに失敗したらエラーを返す",
			stockPositionStore: &testStockPositionStore{getByCode1: &stockPosition{OwnedQuantity: 70, HoldQuantity: 0}},
			arg1:               &stockOrder{Side: SideSell, OrderStatus: OrderStatusInOrder, HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 100, ExitQuantity: 30}}},
			arg2:               time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local),
			want:               NotEnoughHoldQuantityError,
			wantArg1:           &stockOrder{Side: SideSell, OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 8, 31, 14, 0, 0, 0, time.Local), HoldPositions: []*HoldPosition{{PositionCode: "mpo-uuid-01", HoldQuantity: 30, ExitQuantity: 30}}},
			wantPosition:       &stockPosition{OwnedQuantity: 70, HoldQuantity: 0}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{stockPositionStore: test.stockPositionStore}
			got := service.cancelAndRelease(test.arg1, test.arg2)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantArg1, test.arg1) ||
				!reflect.DeepEqual(test.wantPosition, test.stockPositionStore.getByCode1) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantArg1, test.wantPosition,
					got, test.arg1, test.stockPositionStore.getByCode1)
			}
		})
	}
}
