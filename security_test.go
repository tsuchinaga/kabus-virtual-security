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
				clock:        &testClock{now: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{getStockOrders: []*stockOrder{}},
			},
			want1: []*StockOrder{},
			want2: nil},
		{name: "storeにある注文をStockOrderに入れ替えて返す",
			security: security{
				clock: &testClock{now: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{getStockOrders: []*stockOrder{
					{
						Code:               "sor_1234",
						OrderStatus:        OrderStatusPart,
						Side:               SideBuy,
						ExecutionCondition: StockExecutionConditionMO,
						SymbolCode:         "1234",
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
						ClosePositionCode:  "spo-1234",
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
					ClosePositionCode:  "spo-1234",
				},
			},
			want2: nil},
		{name: "storeに複数注文があれば全部返す",
			security: security{
				clock: &testClock{now: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{getStockOrders: []*stockOrder{
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
				stockService: &testStockService{getStockOrders: []*stockOrder{
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
			security: security{stockService: &testStockService{
				getStockPositions: []*stockPosition{},
			}},
			want1: []*StockPosition{},
			want2: nil},
		{name: "storeにあるデータをStockPositionに詰め替えて返す",
			security: security{stockService: &testStockService{
				getStockPositions: []*stockPosition{
					{
						Code:               "spo_1234",
						OrderCode:          "sor_0123",
						SymbolCode:         "1234",
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
					Side:               SideBuy,
					ContractedQuantity: 300,
					OwnedQuantity:      300,
					HoldQuantity:       100,
					ContractedAt:       time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local),
				},
			},
			want2: nil},
		{name: "storeに複数データがあれば全部返す",
			security: security{stockService: &testStockService{
				getStockPositions: []*stockPosition{
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
			security: security{stockService: &testStockService{
				getStockPositions: []*stockPosition{
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
				stockService: &testStockService{
					getStockOrderByCode1: nil,
					getStockOrderByCode2: NoDataError,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: NoDataError},
		{name: "引数がnilならエラー",
			security: &security{
				clock:        &testClock{now: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{}},
			arg:  nil,
			want: NilArgumentError},
		{name: "キャンセル不可な状態の注文ならエラー",
			security: &security{
				clock: &testClock{now: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{
					getStockOrderByCode1: &stockOrder{Code: "sor_1234", OrderStatus: OrderStatusCanceled},
					getStockOrderByCode2: nil,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: UncancellableOrderError},
		{name: "キャンセル可能な注文ならエラーなし",
			security: &security{
				clock: &testClock{now: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{
					getStockOrderByCode1: &stockOrder{Code: "sor_1234", OrderStatus: OrderStatusInOrder},
					getStockOrderByCode2: nil,
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

func Test_Security_StockOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		clock          *testClock
		priceStore     *testPriceStore
		stockService   *testStockService
		arg            *StockOrderRequest
		want1          *OrderResult
		want2          error
		wantEntryCount int
		wantExitCount  int
	}{
		{name: "引数がnilであればエラーを返す", stockService: &testStockService{}, arg: nil, want1: nil, want2: NilArgumentError},
		{name: "validationでエラーがあればエラーを返す",
			clock:        &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: nil, addStockOrderHistory: []*stockOrder{}},
			arg:          &StockOrderRequest{},
			want1:        nil, want2: InvalidSideError},
		{name: "該当銘柄の価格情報を取得し、価格情報がなくエラーが返されたらエラーを返す",
			clock:        &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			priceStore:   &testPriceStore{getBySymbolCode1: nil, getBySymbolCode2: InvalidSymbolCodeError},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: nil, addStockOrderHistory: []*stockOrder{}},
			arg: &StockOrderRequest{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 6, 25, 15, 0, 0, 0, time.Local),
			},
			want1: nil,
			want2: InvalidSymbolCodeError},
		{name: "該当銘柄の価格情報を取得し、価格情報がなければ、注文の保存を行ない、注文結果を返す",
			clock:        &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			priceStore:   &testPriceStore{getBySymbolCode1: nil, getBySymbolCode2: NoDataError},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: nil, addStockOrderHistory: []*stockOrder{}},
			arg: &StockOrderRequest{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 6, 25, 15, 0, 0, 0, time.Local),
			},
			want1: &OrderResult{OrderCode: "sor-1"},
			want2: nil},
		{name: "該当銘柄の価格情報を取得し、価格情報がない場合の注文の保存でエラーがあれば、エラーを返す",
			clock:        &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			priceStore:   &testPriceStore{getBySymbolCode1: nil, getBySymbolCode2: NoDataError},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: NilArgumentError, addStockOrderHistory: []*stockOrder{}},
			arg: &StockOrderRequest{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 6, 25, 15, 0, 0, 0, time.Local),
			},
			want1: nil,
			want2: NilArgumentError},
		{name: "該当銘柄の価格情報を取得し、価格情報で約定しない場合は注文の保存を行い、注文結果を返す",
			clock:        &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local), getStockSession: SessionUnspecified},
			priceStore:   &testPriceStore{getBySymbolCode1: &symbolPrice{Price: 1000}, getBySymbolCode2: nil},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: NilArgumentError, addStockOrderHistory: []*stockOrder{}},
			arg: &StockOrderRequest{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 6, 25, 15, 0, 0, 0, time.Local),
			},
			want1: &OrderResult{OrderCode: "sor-1"},
			want2: nil},
		{name: "該当銘柄の価格情報を取得し、価格情報で買い注文が約定した場合はエントリーし、注文結果を返す",
			clock: &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local), getStockSession: SessionMorning},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Ask:        1000,
				AskTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				Bid:        1000,
				BidTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: NilArgumentError, addStockOrderHistory: []*stockOrder{}},
			arg: &StockOrderRequest{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 6, 25, 15, 0, 0, 0, time.Local),
			},
			want1:          &OrderResult{OrderCode: "sor-1"},
			want2:          nil,
			wantEntryCount: 1},
		{name: "該当銘柄の価格情報を取得し、価格情報で買い注文が約定できずエラーが返されたら、エラーを返す",
			clock: &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local), getStockSession: SessionMorning},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Ask:        1000,
				AskTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				Bid:        1000,
				BidTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: NilArgumentError, addStockOrderHistory: []*stockOrder{}, entry: NoDataError},
			arg: &StockOrderRequest{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 6, 25, 15, 0, 0, 0, time.Local),
			},
			want1:          nil,
			want2:          NoDataError,
			wantEntryCount: 1},
		{name: "該当銘柄の価格情報を取得し、価格情報で売り注文が約定した場合はエグジットし、注文結果を返す",
			clock: &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local), getStockSession: SessionMorning},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Ask:        1000,
				AskTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				Bid:        1000,
				BidTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: NilArgumentError, addStockOrderHistory: []*stockOrder{}},
			arg: &StockOrderRequest{
				Side:               SideSell,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 6, 25, 15, 0, 0, 0, time.Local),
			},
			want1:         &OrderResult{OrderCode: "sor-1"},
			want2:         nil,
			wantExitCount: 1},
		{name: "該当銘柄の価格情報を取得し、価格情報で売り注文が約定できずエラーが返されたら、エラーを返す",
			clock: &testClock{now: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local), getStockSession: SessionMorning},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Ask:        1000,
				AskTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				Bid:        1000,
				BidTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			stockService: &testStockService{newOrderCode: []string{"sor-1"}, addStockOrder: NilArgumentError, addStockOrderHistory: []*stockOrder{}, exit: NoDataError},
			arg: &StockOrderRequest{
				Side:               SideSell,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				Quantity:           100,
				ExpiredAt:          time.Date(2021, 6, 25, 15, 0, 0, 0, time.Local),
			},
			want1:         nil,
			want2:         NoDataError,
			wantExitCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			security := &security{clock: test.clock, priceStore: test.priceStore, stockService: test.stockService}
			got1, got2 := security.StockOrder(test.arg)
			if !reflect.DeepEqual(test.want1, got1) ||
				!errors.Is(got2, test.want2) ||
				!reflect.DeepEqual(test.wantEntryCount, test.stockService.entryCount) ||
				!reflect.DeepEqual(test.wantExitCount, test.stockService.exitCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v\n", t.Name(),
					test.want1, test.want2, test.wantEntryCount, test.wantExitCount,
					got1, got2, test.stockService.entryCount, test.stockService.exitCount)
			}
		})
	}
}

func Test_security_RegisterPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		clock          *testClock
		priceService   *testPriceService
		priceStore     *testPriceStore
		stockService   *testStockService
		arg            RegisterPriceRequest
		want           error
		wantEntryCount int
		wantExitCount  int
	}{
		{name: "validationでエラーがあればエラーを返す",
			priceService: &testPriceService{validation1: InvalidTimeError},
			stockService: &testStockService{},
			want:         InvalidTimeError},
		{name: "toSymbolPriceでエラーがあればエラーを返す",
			priceService: &testPriceService{toSymbolPrice2: NilArgumentError},
			stockService: &testStockService{},
			want:         NilArgumentError},
		{name: "storeへの保存でエラーがあればエラーを返す",
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{}},
			priceStore:   &testPriceStore{set: NilArgumentError},
			stockService: &testStockService{},
			want:         NilArgumentError},
		{name: "保存された注文がなければ何もしない",
			clock: &testClock{
				now:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession: SessionMorning},
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{}},
			priceStore:   &testPriceStore{},
			stockService: &testStockService{getStockOrders: []*stockOrder{}},
			want:         nil},
		{name: "保存された注文があっても約定不可なら何もしない",
			clock: &testClock{
				now:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession: SessionMorning},
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{SymbolCode: "1234"}},
			priceStore:   &testPriceStore{},
			stockService: &testStockService{getStockOrders: []*stockOrder{{SymbolCode: "0000"}}},
			want:         nil},
		{name: "保存された注文が約定可能で買いならEntryを実行する",
			clock: &testClock{
				now:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession: SessionMorning},
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{
				SymbolCode: "1234",
				Price:      1000,
				PriceTime:  time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Bid:        1010,
				BidTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Ask:        990,
				AskTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular}},
			priceStore: &testPriceStore{},
			stockService: &testStockService{getStockOrders: []*stockOrder{{
				SymbolCode:         "1234",
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				OrderStatus:        OrderStatusInOrder}}},
			wantEntryCount: 1},
		{name: "保存された注文が約定可能で売りならExitを実行する",
			clock: &testClock{
				now:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession: SessionMorning},
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{
				SymbolCode: "1234",
				Price:      1000,
				PriceTime:  time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Bid:        1010,
				BidTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Ask:        990,
				AskTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular}},
			priceStore: &testPriceStore{},
			stockService: &testStockService{getStockOrders: []*stockOrder{{
				SymbolCode:         "1234",
				Side:               SideSell,
				ExecutionCondition: StockExecutionConditionMO,
				OrderStatus:        OrderStatusInOrder}}},
			wantExitCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			security := &security{clock: test.clock, priceStore: test.priceStore, priceService: test.priceService, stockService: test.stockService}
			got := security.RegisterPrice(test.arg)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantEntryCount, test.stockService.entryCount) ||
				!reflect.DeepEqual(test.wantExitCount, test.stockService.exitCount) {

				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantEntryCount, test.wantExitCount,
					got, test.stockService.entryCount, test.stockService.exitCount)
			}
		})
	}
}
