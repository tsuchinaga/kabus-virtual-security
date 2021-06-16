package virtual_security

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_security_StockOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security security
		want1    []*StockOrder
		want2    error
	}{
		{name: "storeに注文がなければ空配列",
			security: security{
				clock:           &testClock{now: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockOrderStore: &testStockOrderStore{getAll: []*stockOrder{}},
			},
			want1: []*StockOrder{},
			want2: nil},
		{name: "storeにある注文をStockOrderに入れ替えて返す",
			security: security{
				clock: &testClock{now: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockOrderStore: &testStockOrderStore{getAll: []*stockOrder{
					{
						Code:               "sor_1234",
						OrderStatus:        OrderStatusPart,
						Side:               SideBuy,
						ExecutionCondition: StockExecutionConditionMO,
						SymbolCode:         "1234",
						Exchange:           ExchangeToushou,
						OrderQuantity:      300,
						ContractedQuantity: 100,
						CanceledQuantity:   0,
						LimitPrice:         0,
						ExpiredAt:          time.Date(2021, 6, 14, 15, 0, 0, 0, time.Local),
						StopCondition:      nil,
						OrderedAt:          time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local),
						CanceledAt:         time.Time{},
						Contracts:          []*Contract{},
						ConfirmingCount:    20,
						Message:            "",
					},
				}},
			},
			want1: []*StockOrder{
				{
					Code:               "sor_1234",
					OrderStatus:        OrderStatusPart,
					Side:               SideBuy,
					ExecutionCondition: StockExecutionConditionMO,
					SymbolCode:         "1234",
					Exchange:           ExchangeToushou,
					OrderQuantity:      300,
					ContractedQuantity: 100,
					CanceledQuantity:   0,
					LimitPrice:         0,
					ExpiredAt:          time.Date(2021, 6, 14, 15, 0, 0, 0, time.Local),
					StopCondition:      nil,
					OrderedAt:          time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local),
					CanceledAt:         time.Time{},
					Contracts:          []*Contract{},
					Message:            "",
				},
			},
			want2: nil},
		{name: "storeに複数注文があれば全部返す",
			security: security{
				clock: &testClock{now: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockOrderStore: &testStockOrderStore{getAll: []*stockOrder{
					{Code: "sor_1234", OrderStatus: OrderStatusInOrder},
					{Code: "sor_2345", OrderStatus: OrderStatusInOrder},
					{Code: "sor_3456", OrderStatus: OrderStatusInOrder},
				}},
			},
			want1: []*StockOrder{
				{Code: "sor_1234", OrderStatus: OrderStatusInOrder},
				{Code: "sor_2345", OrderStatus: OrderStatusInOrder},
				{Code: "sor_3456", OrderStatus: OrderStatusInOrder},
			},
			want2: nil},
		{name: "storeにある注文が死んだ注文なら返さない",
			security: security{
				clock: &testClock{now: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockOrderStore: &testStockOrderStore{getAll: []*stockOrder{
					{Code: "sor_1234", OrderStatus: OrderStatusDone},
					{Code: "sor_2345", OrderStatus: OrderStatusInOrder},
					{Code: "sor_3456", OrderStatus: OrderStatusCanceled},
				}},
			},
			want1: []*StockOrder{
				{Code: "sor_2345", OrderStatus: OrderStatusInOrder},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.security.StockOrders()
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_security_StockPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security security
		want1    []*StockPosition
		want2    error
	}{
		{name: "storeにデータが無ければ空配列を返す",
			security: security{stockPositionStore: &testStockPositionStore{
				getAll: []*stockPosition{},
			}},
			want1: []*StockPosition{},
			want2: nil},
		{name: "storeにあるデータをStockPositionに詰め替えて返す",
			security: security{stockPositionStore: &testStockPositionStore{
				getAll: []*stockPosition{
					{
						Code:               "spo_1234",
						OrderCode:          "sor_0123",
						SymbolCode:         "1234",
						Exchange:           ExchangeToushou,
						Side:               SideBuy,
						ContractedQuantity: 300,
						OwnedQuantity:      300,
						HoldQuantity:       100,
						ContractedAt:       time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local),
					},
				},
			}},
			want1: []*StockPosition{
				{
					Code:               "spo_1234",
					OrderCode:          "sor_0123",
					SymbolCode:         "1234",
					Exchange:           ExchangeToushou,
					Side:               SideBuy,
					ContractedQuantity: 300,
					OwnedQuantity:      300,
					HoldQuantity:       100,
					ContractedAt:       time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local),
				},
			},
			want2: nil},
		{name: "storeに複数データがあれば全部返す",
			security: security{stockPositionStore: &testStockPositionStore{
				getAll: []*stockPosition{
					{Code: "spo_1234", OwnedQuantity: 100},
					{Code: "spo_2345", OwnedQuantity: 100},
					{Code: "spo_3456", OwnedQuantity: 100},
				},
			}},
			want1: []*StockPosition{
				{Code: "spo_1234", OwnedQuantity: 100},
				{Code: "spo_2345", OwnedQuantity: 100},
				{Code: "spo_3456", OwnedQuantity: 100},
			},
			want2: nil},
		{name: "storeのデータが死んでいたら返さない",
			security: security{stockPositionStore: &testStockPositionStore{
				getAll: []*stockPosition{
					{Code: "spo_1234", OwnedQuantity: 0},
					{Code: "spo_2345", OwnedQuantity: 100},
					{Code: "spo_3456", OwnedQuantity: 0},
				},
			}},
			want1: []*StockPosition{
				{Code: "spo_2345", OwnedQuantity: 100},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.security.StockPositions()
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_security_CancelStockOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security *security
		arg      *CancelOrderRequest
		want     error
	}{
		{name: "注文がなければstoreからエラーが返されるので、そのエラーかラップしたエラーを返す",
			security: &security{
				clock: &testClock{now: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockOrderStore: &testStockOrderStore{
					getByCode1: nil,
					getByCode2: NoDataError,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: NoDataError},
		{name: "引数がnilならエラー",
			security: &security{
				clock:           &testClock{now: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockOrderStore: &testStockOrderStore{}},
			arg:  nil,
			want: NilArgumentError},
		{name: "キャンセル不可な状態の注文ならエラー",
			security: &security{
				clock: &testClock{now: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockOrderStore: &testStockOrderStore{
					getByCode1: &stockOrder{Code: "sor_1234", OrderStatus: OrderStatusCanceled},
					getByCode2: nil,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: UncancellableOrderError},
		{name: "キャンセル可能な注文ならエラーなし",
			security: &security{
				clock: &testClock{now: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockOrderStore: &testStockOrderStore{
					getByCode1: &stockOrder{Code: "sor_1234", OrderStatus: OrderStatusInOrder},
					getByCode2: nil,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.security.CancelStockOrder(test.arg)
			if !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
