package virtual_security

import (
	"errors"
	"log"
	"reflect"
	"testing"
	"time"
)

type testStockService struct {
	newOrderCode1     []string
	newOrderCodeCount int
	entry1            error
	entryCount        int
	entryHistory      []struct {
		order          *stockOrder
		contractResult *confirmContractResult
	}
	exit1       error
	exitCount   int
	exitHistory []struct {
		order          *stockOrder
		contractResult *confirmContractResult
	}
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
}

func (t *testStockService) addStockOrder(order *stockOrder) error {
	t.addStockOrderHistory = append(t.addStockOrderHistory, order)
	return t.addStockOrder1
}

func (t *testStockService) newOrderCode() string {
	defer func() { t.newOrderCodeCount++ }()
	return t.newOrderCode1[t.newOrderCodeCount%len(t.newOrderCode1)]
}

func (t *testStockService) entry(order *stockOrder, contractResult *confirmContractResult) error {
	t.entryHistory = append(t.entryHistory, struct {
		order          *stockOrder
		contractResult *confirmContractResult
	}{order: order, contractResult: contractResult})
	t.entryCount++
	return t.entry1
}

func (t *testStockService) exit(order *stockOrder, contractResult *confirmContractResult) error {
	t.exitHistory = append(t.exitHistory, struct {
		order          *stockOrder
		contractResult *confirmContractResult
	}{order: order, contractResult: contractResult})
	t.exitCount++
	return t.exit1
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

func Test_stockService_Entry(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                 string
		stockService         *stockService
		arg1                 *stockOrder
		arg2                 *confirmContractResult
		want                 error
		wantOrderStoreAdd    []*stockOrder
		wantPositionStoreAdd []*stockPosition
	}{
		{name: "それぞれコードを生成し、注文、ポジションをstoreに保存する",
			stockService: &stockService{uuidGenerator: &testUUIDGenerator{generator1: []string{"uuid-1", "uuid-2", "uuid-3"}}},
			arg1: &stockOrder{
				Code:               "sor-1",
				SymbolCode:         "1234",
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				OrderQuantity:      100,
				OrderedAt:          time.Date(2021, 6, 21, 10, 0, 0, 0, time.Local),
				ConfirmingCount:    1,
			},
			arg2: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
			},
			want: nil,
			wantOrderStoreAdd: []*stockOrder{
				{
					Code:               "sor-1",
					OrderStatus:        OrderStatusDone,
					ExecutionCondition: StockExecutionConditionMO,
					SymbolCode:         "1234",
					Side:               SideBuy,
					OrderQuantity:      100,
					ContractedQuantity: 100,
					OrderedAt:          time.Date(2021, 6, 21, 10, 0, 0, 0, time.Local),
					Contracts: []*Contract{
						{
							ContractCode: "sco-uuid-1",
							OrderCode:    "sor-1",
							PositionCode: "spo-uuid-2",
							Price:        1000,
							Quantity:     100,
							ContractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
						},
					},
					ConfirmingCount: 1,
				},
			},
			wantPositionStoreAdd: []*stockPosition{
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

			got := test.stockService.entry(test.arg1, test.arg2)
			if !errors.Is(got, test.want) || !reflect.DeepEqual(test.wantOrderStoreAdd, stockOrderStore.addHistory) || !reflect.DeepEqual(test.wantPositionStoreAdd, stockPositionStore.addHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantOrderStoreAdd, test.wantPositionStoreAdd,
					got, stockOrderStore.addHistory, stockPositionStore.addHistory)
			}
		})
	}
}

func Test_stockService_Exit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		stockService       *stockService
		stockOrderStore    *testStockOrderStore
		stockPositionStore *testStockPositionStore
		arg1               *stockOrder
		arg2               *confirmContractResult
		want               error
		wantOrderStoreAdd  []*stockOrder
	}{
		{name: "指定した銘柄のポジション取得に失敗したらエラー",
			stockService:       &stockService{},
			stockOrderStore:    &testStockOrderStore{addHistory: []*stockOrder{}},
			stockPositionStore: &testStockPositionStore{getBySymbolCode1: nil, getBySymbolCode2: NilArgumentError},
			arg1:               &stockOrder{},
			arg2:               &confirmContractResult{},
			want:               NilArgumentError,
			wantOrderStoreAdd:  []*stockOrder{}},
		{name: "注文可能なポジションの総数より注文数が多ければエラー",
			stockService:       &stockService{},
			stockOrderStore:    &testStockOrderStore{addHistory: []*stockOrder{}},
			stockPositionStore: &testStockPositionStore{getBySymbolCode1: []*stockPosition{{OwnedQuantity: 100}, {OwnedQuantity: 200}, {OwnedQuantity: 300}}},
			arg1:               &stockOrder{OrderQuantity: 700},
			arg2:               &confirmContractResult{},
			want:               NotEnoughOwnedQuantityError,
			wantOrderStoreAdd:  []*stockOrder{}},
		{name: "注文の数量を丁度満たせるよう古いポジションから順にexitする",
			stockService:       &stockService{uuidGenerator: &testUUIDGenerator{generator1: []string{"uuid-1", "uuid-2", "uuid-3", "uuid-4", "uuid-5"}}},
			stockOrderStore:    &testStockOrderStore{addHistory: []*stockOrder{}},
			stockPositionStore: &testStockPositionStore{getBySymbolCode1: []*stockPosition{{Code: "spo-0", OwnedQuantity: 1000, HoldQuantity: 1000}, {Code: "spo-1", OwnedQuantity: 100}, {Code: "spo-2", OwnedQuantity: 200, HoldQuantity: 100}, {Code: "spo-3", OwnedQuantity: 300}}},
			arg1:               &stockOrder{Code: "sor-1", OrderQuantity: 400},
			arg2: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
			},
			want: nil,
			wantOrderStoreAdd: []*stockOrder{{
				Code:               "sor-1",
				OrderStatus:        OrderStatusDone,
				OrderQuantity:      400,
				ContractedQuantity: 400,
				Contracts: []*Contract{
					{
						ContractCode: "sco-uuid-1",
						OrderCode:    "sor-1",
						PositionCode: "spo-1",
						Price:        1000,
						Quantity:     100,
						ContractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
					},
					{
						ContractCode: "sco-uuid-2",
						OrderCode:    "sor-1",
						PositionCode: "spo-2",
						Price:        1000,
						Quantity:     100,
						ContractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
					},
					{
						ContractCode: "sco-uuid-3",
						OrderCode:    "sor-1",
						PositionCode: "spo-3",
						Price:        1000,
						Quantity:     200,
						ContractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
					},
				}}},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.stockService.stockOrderStore = test.stockOrderStore
			test.stockService.stockPositionStore = test.stockPositionStore
			got := test.stockService.exit(test.arg1, test.arg2)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantOrderStoreAdd, test.stockOrderStore.addHistory) {

				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(),
					test.want, test.wantOrderStoreAdd,
					got, test.stockOrderStore.addHistory)
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

func Test_stockService_AddStockOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		store          *testStockOrderStore
		arg            *stockOrder
		want           error
		wantAddHistory []*stockOrder
	}{
		{name: "引数が有効な注文ならstoreに渡す", store: &testStockOrderStore{addHistory: []*stockOrder{}}, arg: &stockOrder{Code: "sor-1"}, wantAddHistory: []*stockOrder{{Code: "sor-1"}}},
		{name: "引数がnilでもstoreに渡す", store: &testStockOrderStore{addHistory: []*stockOrder{}}, arg: nil, wantAddHistory: []*stockOrder{nil}},
		{name: "errorが返されたらそのerrorを返す", store: &testStockOrderStore{addHistory: []*stockOrder{}, add1: NilArgumentError}, arg: nil, want: NilArgumentError, wantAddHistory: []*stockOrder{nil}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{stockOrderStore: test.store}
			got := service.addStockOrder(test.arg)
			if !errors.Is(got, test.want) || !reflect.DeepEqual(test.wantAddHistory, test.store.addHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantAddHistory, got, test.store.addHistory)
			}
		})
	}
}

func Test_newStockService(t *testing.T) {
	t.Parallel()
	uuid := &testUUIDGenerator{}
	stockOrderStore := &testStockOrderStore{}
	stockPositionStore := &testStockPositionStore{}
	want := &stockService{uuidGenerator: uuid, stockOrderStore: stockOrderStore, stockPositionStore: stockPositionStore}
	got := newStockService(uuid, stockOrderStore, stockPositionStore)
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
