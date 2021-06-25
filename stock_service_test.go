package virtual_security

import (
	"errors"
	"log"
	"reflect"
	"testing"
	"time"
)

type testStockService struct {
	newOrderCode      []string
	newOrderCodeCount int
	entry             error
	entryCount        int
	entryHistory      []struct {
		order          *stockOrder
		contractResult *confirmContractResult
	}
	exit        error
	exitCount   int
	exitHistory []struct {
		order          *stockOrder
		contractResult *confirmContractResult
	}
	getStockOrders                   []*stockOrder
	getStockOrderByCode1             *stockOrder
	getStockOrderByCode2             error
	getStockOrderByCodeHistory       []string
	removeStockOrderByCodeHistory    []string
	getStockPositions                []*stockPosition
	removeStockPositionByCodeHistory []string
	addStockOrder                    error
	addStockOrderHistory             []*stockOrder
}

func (t *testStockService) AddStockOrder(order *stockOrder) error {
	t.addStockOrderHistory = append(t.addStockOrderHistory, order)
	return t.addStockOrder
}

func (t *testStockService) NewOrderCode() string {
	defer func() { t.newOrderCodeCount++ }()
	return t.newOrderCode[t.newOrderCodeCount%len(t.newOrderCode)]
}

func (t *testStockService) Entry(order *stockOrder, contractResult *confirmContractResult) error {
	t.entryHistory = append(t.entryHistory, struct {
		order          *stockOrder
		contractResult *confirmContractResult
	}{order: order, contractResult: contractResult})
	t.entryCount++
	return t.entry
}

func (t *testStockService) Exit(order *stockOrder, contractResult *confirmContractResult) error {
	t.exitHistory = append(t.exitHistory, struct {
		order          *stockOrder
		contractResult *confirmContractResult
	}{order: order, contractResult: contractResult})
	t.exitCount++
	return t.exit
}

func (t *testStockService) GetStockOrders() []*stockOrder {
	return t.getStockOrders
}

func (t *testStockService) GetStockOrderByCode(orderCode string) (*stockOrder, error) {
	t.getStockOrderByCodeHistory = append(t.getStockOrderByCodeHistory, orderCode)
	return t.getStockOrderByCode1, t.getStockOrderByCode2
}

func (t *testStockService) RemoveStockOrderByCode(orderCode string) {
	t.removeStockOrderByCodeHistory = append(t.removeStockOrderByCodeHistory, orderCode)
}

func (t *testStockService) GetStockPositions() []*stockPosition {
	return t.getStockPositions
}

func (t *testStockService) RemoveStockPositionByCode(positionCode string) {
	t.removeStockPositionByCodeHistory = append(t.removeStockPositionByCodeHistory, positionCode)
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
			stockService: &stockService{uuidGenerator: &testUUIDGenerator{generator: []string{"uuid-1", "uuid-2", "uuid-3"}}},
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
							ContractCode: "con-uuid-1",
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

			got := test.stockService.Entry(test.arg1, test.arg2)
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
		{name: "指定したポジションがなければエラー",
			stockService:       &stockService{},
			stockOrderStore:    &testStockOrderStore{addHistory: []*stockOrder{}},
			stockPositionStore: &testStockPositionStore{getByCode1: nil, getByCode2: NoDataError, addHistory: []*stockPosition{}},
			arg1:               &stockOrder{},
			arg2:               &confirmContractResult{},
			want:               NoDataError,
			wantOrderStoreAdd:  []*stockOrder{},
		},
		{name: "指定したポジションをholdできなければエラー",
			stockService:       &stockService{},
			stockOrderStore:    &testStockOrderStore{addHistory: []*stockOrder{}},
			stockPositionStore: &testStockPositionStore{getByCode1: &stockPosition{Code: "spo-1", OwnedQuantity: 300, HoldQuantity: 300}, addHistory: []*stockPosition{}},
			arg1:               &stockOrder{ClosePositionCode: "spo-1", OrderQuantity: 100},
			arg2:               &confirmContractResult{},
			want:               NotEnoughOwnedQuantityError,
			wantOrderStoreAdd:  []*stockOrder{},
		},
		{name: "注文に約定情報を追加し、ストアに保存する",
			stockService:       &stockService{uuidGenerator: &testUUIDGenerator{generator: []string{"uuid-1", "uuid-2", "uuid-3"}}},
			stockOrderStore:    &testStockOrderStore{addHistory: []*stockOrder{}},
			stockPositionStore: &testStockPositionStore{getByCode1: &stockPosition{Code: "spo-1", OwnedQuantity: 300, HoldQuantity: 200}, addHistory: []*stockPosition{}},
			arg1:               &stockOrder{Code: "sor-1", ClosePositionCode: "spo-1", OrderQuantity: 100},
			arg2: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 6, 21, 10, 1, 0, 0, time.Local),
			},
			want: nil,
			wantOrderStoreAdd: []*stockOrder{{
				Code:               "sor-1",
				OrderStatus:        OrderStatusDone,
				ClosePositionCode:  "spo-1",
				OrderQuantity:      100,
				ContractedQuantity: 100,
				Contracts: []*Contract{
					{
						ContractCode: "con-uuid-1",
						OrderCode:    "sor-1",
						PositionCode: "spo-1",
						Price:        1000,
						Quantity:     100,
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
			got := test.stockService.Exit(test.arg1, test.arg2)
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
		uuidGenerator UUIDGenerator
		want          string
	}{
		{name: "uuidの結果に接頭辞を付けて返す", uuidGenerator: &testUUIDGenerator{generator: []string{"uuid-1", "uuid-2", "uuid-3"}}, want: "sor-uuid-1"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{uuidGenerator: test.uuidGenerator}
			got := service.NewOrderCode()
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
		stockService StockService
		want         []*stockOrder
	}{
		{name: "storeの結果が空なら空",
			stockService: &stockService{stockOrderStore: &testStockOrderStore{getAll: []*stockOrder{}}},
			want:         []*stockOrder{}},
		{name: "storeの結果をそのまま返す",
			stockService: &stockService{stockOrderStore: &testStockOrderStore{getAll: []*stockOrder{{Code: "sor-1"}, {Code: "sor-2"}, {Code: "sor-3"}}}},
			want:         []*stockOrder{{Code: "sor-1"}, {Code: "sor-2"}, {Code: "sor-3"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockService.GetStockOrders()
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
		stockService StockService
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
			got1, got2 := test.stockService.GetStockOrderByCode(test.arg)
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
			service.RemoveStockOrderByCode(test.arg)
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
		service StockService
		want    []*stockPosition
	}{
		{name: "storeが空配列を返したら彼配列",
			service: &stockService{stockPositionStore: &testStockPositionStore{getAll: []*stockPosition{}}},
			want:    []*stockPosition{}},
		{name: "storeが複数要素を返したラそのまま返す",
			service: &stockService{stockPositionStore: &testStockPositionStore{getAll: []*stockPosition{{Code: "spo-1"}, {Code: "spo-2"}, {Code: "spo-3"}}}},
			want:    []*stockPosition{{Code: "spo-1"}, {Code: "spo-2"}, {Code: "spo-3"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.service.GetStockPositions()
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
			service.RemoveStockPositionByCode(test.arg)
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
		{name: "errorが返されたらそのerrorを返す", store: &testStockOrderStore{addHistory: []*stockOrder{}, add: NilArgumentError}, arg: nil, want: NilArgumentError, wantAddHistory: []*stockOrder{nil}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &stockService{stockOrderStore: test.store}
			got := service.AddStockOrder(test.arg)
			if !errors.Is(got, test.want) || !reflect.DeepEqual(test.wantAddHistory, test.store.addHistory) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantAddHistory, got, test.store.addHistory)
			}
		})
	}
}
