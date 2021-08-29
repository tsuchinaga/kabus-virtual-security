package virtual_security

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_virtualSecurity_StockOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security virtualSecurity
		want1    []*StockOrder
		want2    error
	}{
		{name: "storeに注文がなければ空配列",
			security: virtualSecurity{
				clock:        &testClock{now1: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{getStockOrders1: []*stockOrder{}},
			},
			want1: []*StockOrder{},
			want2: nil},
		{name: "storeにある注文をStockOrderに入れ替えて返す",
			security: virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{getStockOrders1: []*stockOrder{
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
				},
			},
			want2: nil},
		{name: "storeに複数注文があれば全部返す",
			security: virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{getStockOrders1: []*stockOrder{
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
			security: virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{getStockOrders1: []*stockOrder{
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

func Test_virtualSecurity_StockPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security virtualSecurity
		want1    []*StockPosition
		want2    error
	}{
		{name: "storeにデータが無ければ空配列を返す",
			security: virtualSecurity{stockService: &testStockService{
				getStockPositions1: []*stockPosition{},
			}},
			want1: []*StockPosition{},
			want2: nil},
		{name: "storeにあるデータをStockPositionに詰め替えて返す",
			security: virtualSecurity{stockService: &testStockService{
				getStockPositions1: []*stockPosition{
					{
						Code:               "spo_1234",
						OrderCode:          "sor_0123",
						SymbolCode:         "1234",
						Side:               SideBuy,
						ContractedQuantity: 300,
						OwnedQuantity:      300,
						HoldQuantity:       100,
						Price:              1000,
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
					Price:              1000,
					ContractedAt:       time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local),
				},
			},
			want2: nil},
		{name: "storeに複数データがあれば全部返す",
			security: virtualSecurity{stockService: &testStockService{
				getStockPositions1: []*stockPosition{
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
			security: virtualSecurity{stockService: &testStockService{
				getStockPositions1: []*stockPosition{
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

func Test_virtualSecurity_CancelStockOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security *virtualSecurity
		arg      *CancelOrderRequest
		want     error
	}{
		{name: "注文がなければstoreからエラーが返されるので、そのエラーかラップしたエラーを返す",
			security: &virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{
					getStockOrderByCode1: nil,
					getStockOrderByCode2: NoDataError,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: NoDataError},
		{name: "引数がnilならエラー",
			security: &virtualSecurity{
				clock:        &testClock{now1: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{}},
			arg:  nil,
			want: NilArgumentError},
		{name: "キャンセル不可な状態の注文ならエラー",
			security: &virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				stockService: &testStockService{
					getStockOrderByCode1: &stockOrder{Code: "sor_1234", OrderStatus: OrderStatusCanceled},
					getStockOrderByCode2: nil,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: UncancellableOrderError},
		{name: "キャンセル可能な注文ならエラーなし",
			security: &virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
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

func Test_virtualSecurity_StockOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                     string
		clock                    *testClock
		priceService             *testPriceService
		stockService             *testStockService
		arg                      *StockOrderRequest
		want1                    *OrderResult
		want2                    error
		wantConfirmContractCount int
		wantSaveStockOrder       []*stockOrder
	}{
		{name: "引数がnilであればエラーを返す", stockService: &testStockService{}, arg: nil, want1: nil, want2: NilArgumentError},
		{name: "validationでエラーがあればエラーを返す",
			clock:        &testClock{now1: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			stockService: &testStockService{toStockOrder1: &stockOrder{}, validation1: InvalidSideError},
			arg:          &StockOrderRequest{},
			want1:        nil,
			want2:        InvalidSideError},
		{name: "sell注文のholdに失敗したらエラーを返す",
			clock:        &testClock{now1: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			stockService: &testStockService{toStockOrder1: &stockOrder{Side: SideSell}, validation1: nil, holdSellOrderPositions1: NotEnoughOwnedQuantityError},
			arg:          &StockOrderRequest{Side: SideSell},
			want1:        nil,
			want2:        NotEnoughOwnedQuantityError},
		{name: "該当銘柄の価格情報を取得し、価格情報なし以外のエラーが返されたらエラーを返す",
			clock:              &testClock{now1: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			priceService:       &testPriceService{getBySymbolCode1: nil, getBySymbolCode2: InvalidSymbolCodeError},
			stockService:       &testStockService{toStockOrder1: &stockOrder{Code: "sor-01", Side: SideSell}},
			arg:                &StockOrderRequest{Side: SideSell},
			want1:              nil,
			want2:              InvalidSymbolCodeError,
			wantSaveStockOrder: []*stockOrder{{Code: "sor-01", Side: SideSell}}},
		{name: "該当銘柄の価格情報を取得し、価格情報なしならentryもexitもしない",
			clock:              &testClock{now1: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			priceService:       &testPriceService{getBySymbolCode1: nil, getBySymbolCode2: NoDataError},
			stockService:       &testStockService{toStockOrder1: &stockOrder{Code: "sor-01", Side: SideSell}},
			arg:                &StockOrderRequest{Side: SideSell},
			want1:              &OrderResult{OrderCode: "sor-01"},
			want2:              nil,
			wantSaveStockOrder: []*stockOrder{{Code: "sor-01", Side: SideSell}}},
		{name: "該当銘柄の価格情報を取得し、価格情報があれば、買い注文はentryする",
			clock: &testClock{now1: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			priceService: &testPriceService{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Bid:        1000,
				BidTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				Ask:        1000,
				AskTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			stockService:             &testStockService{toStockOrder1: &stockOrder{Code: "sor-01", SymbolCode: "1234", Side: SideBuy}},
			arg:                      &StockOrderRequest{SymbolCode: "1234", Side: SideBuy},
			want1:                    &OrderResult{OrderCode: "sor-01"},
			want2:                    nil,
			wantConfirmContractCount: 1,
			wantSaveStockOrder:       []*stockOrder{{Code: "sor-01", SymbolCode: "1234", Side: SideBuy}}},
		{name: "該当銘柄の価格情報を取得し、価格情報があれば、売り注文はexitする",
			clock: &testClock{now1: time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local)},
			priceService: &testPriceService{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Bid:        1000,
				BidTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				Ask:        1000,
				AskTime:    time.Date(2021, 6, 25, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			stockService:             &testStockService{toStockOrder1: &stockOrder{Code: "sor-01", SymbolCode: "1234", Side: SideSell}},
			arg:                      &StockOrderRequest{SymbolCode: "1234", Side: SideSell},
			want1:                    &OrderResult{OrderCode: "sor-01"},
			want2:                    nil,
			wantConfirmContractCount: 1,
			wantSaveStockOrder:       []*stockOrder{{Code: "sor-01", SymbolCode: "1234", Side: SideSell}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			security := &virtualSecurity{clock: test.clock, stockService: test.stockService, priceService: test.priceService}
			got1, got2 := security.StockOrder(test.arg)
			if !reflect.DeepEqual(test.want1, got1) ||
				!errors.Is(got2, test.want2) ||
				!reflect.DeepEqual(test.wantConfirmContractCount, test.stockService.confirmContractCount) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want1, test.want2, test.wantConfirmContractCount,
					got1, got2, test.stockService.confirmContractCount)
			}
		})
	}
}

func Test_virtualSecurity_RegisterPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                           string
		clock                          *testClock
		priceService                   *testPriceService
		stockService                   *testStockService
		marginService                  *testMarginService
		arg                            RegisterPriceRequest
		want                           error
		wantStockConfirmContractCount  int
		wantMarginConfirmContractCount int
	}{
		{name: "validationでエラーがあればエラーを返す",
			priceService:  &testPriceService{validation1: InvalidTimeError},
			stockService:  &testStockService{},
			marginService: &testMarginService{},
			want:          InvalidTimeError},
		{name: "toSymbolPriceでエラーがあればエラーを返す",
			priceService:  &testPriceService{toSymbolPrice2: NilArgumentError},
			stockService:  &testStockService{},
			marginService: &testMarginService{},
			want:          NilArgumentError},
		{name: "storeへの保存でエラーがあればエラーを返す",
			priceService:  &testPriceService{toSymbolPrice1: &symbolPrice{}, set1: NilArgumentError},
			stockService:  &testStockService{},
			marginService: &testMarginService{},
			want:          NilArgumentError},
		{name: "保存された注文がなければ何もしない",
			clock: &testClock{
				now1:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession1: SessionMorning},
			priceService:  &testPriceService{toSymbolPrice1: &symbolPrice{}},
			stockService:  &testStockService{getStockOrders1: []*stockOrder{}},
			marginService: &testMarginService{getMarginOrders1: []*marginOrder{}},
			want:          nil},
		{name: "保存された現物注文がbuyならentryを叩く",
			clock: &testClock{
				now1:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession1: SessionMorning},
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{
				SymbolCode: "1234",
				Price:      1000,
				PriceTime:  time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Ask:        1010,
				AskTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Bid:        990,
				BidTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular}},
			stockService:                  &testStockService{getStockOrders1: []*stockOrder{{Side: SideBuy}}},
			marginService:                 &testMarginService{getMarginOrders1: []*marginOrder{}},
			wantStockConfirmContractCount: 1},
		{name: "保存された現物注文がsellならexitを叩く",
			clock: &testClock{
				now1:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession1: SessionMorning},
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{
				SymbolCode: "1234",
				Price:      1000,
				PriceTime:  time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Ask:        1010,
				AskTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Bid:        990,
				BidTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular}},
			stockService:                  &testStockService{getStockOrders1: []*stockOrder{{Side: SideSell}}},
			marginService:                 &testMarginService{getMarginOrders1: []*marginOrder{}},
			wantStockConfirmContractCount: 1},
		{name: "保存された信用注文がentryならentryを叩く",
			clock: &testClock{
				now1:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession1: SessionMorning},
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{
				SymbolCode: "1234",
				Price:      1000,
				PriceTime:  time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Ask:        1010,
				AskTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Bid:        990,
				BidTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular}},
			stockService:                   &testStockService{getStockOrders1: []*stockOrder{}},
			marginService:                  &testMarginService{getMarginOrders1: []*marginOrder{{TradeType: TradeTypeEntry}}},
			wantMarginConfirmContractCount: 1},
		{name: "保存された信用注文がexitならexitを叩く",
			clock: &testClock{
				now1:             time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				getStockSession1: SessionMorning},
			priceService: &testPriceService{toSymbolPrice1: &symbolPrice{
				SymbolCode: "1234",
				Price:      1000,
				PriceTime:  time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Ask:        1010,
				AskTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				Bid:        990,
				BidTime:    time.Date(2021, 7, 5, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular}},
			stockService:                   &testStockService{getStockOrders1: []*stockOrder{}},
			marginService:                  &testMarginService{getMarginOrders1: []*marginOrder{{TradeType: TradeTypeExit}}},
			wantMarginConfirmContractCount: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			security := &virtualSecurity{clock: test.clock, priceService: test.priceService, stockService: test.stockService, marginService: test.marginService}
			got := security.RegisterPrice(test.arg)
			if !errors.Is(got, test.want) ||
				!reflect.DeepEqual(test.wantStockConfirmContractCount, test.stockService.confirmContractCount) ||
				!reflect.DeepEqual(test.wantMarginConfirmContractCount, test.marginService.confirmContractCount) {

				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.want, test.wantStockConfirmContractCount, test.wantMarginConfirmContractCount,
					got, test.stockService.confirmContractCount, test.marginService.confirmContractCount)
			}
		})
	}
}

func Test_NewVirtualSecurity(t *testing.T) {
	want := &virtualSecurity{
		clock:         newClock(),
		priceService:  newPriceService(newClock(), getPriceStore(newClock())),
		stockService:  newStockService(newUUIDGenerator(), getStockOrderStore(), getStockPositionStore(), newValidatorComponent(), newStockContractComponent()),
		marginService: newMarginService(newUUIDGenerator(), getMarginOrderStore(), getMarginPositionStore(), newValidatorComponent(), newStockContractComponent()),
	}

	got := NewVirtualSecurity()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_virtualSecurity_MarginOrders(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security virtualSecurity
		want1    []*MarginOrder
		want2    error
	}{
		{name: "storeに注文がなければ空配列",
			security: virtualSecurity{
				clock:         &testClock{now1: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				marginService: &testMarginService{getMarginOrders1: []*marginOrder{}},
			},
			want1: []*MarginOrder{},
			want2: nil},
		{name: "storeにある注文をMarginOrderに入れ替えて返す",
			security: virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				marginService: &testMarginService{getMarginOrders1: []*marginOrder{
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
					},
				}},
			},
			want1: []*MarginOrder{
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
				},
			},
			want2: nil},
		{name: "storeに複数注文があれば全部返す",
			security: virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				marginService: &testMarginService{getMarginOrders1: []*marginOrder{
					{Code: "sor_1234", OrderStatus: OrderStatusInOrder},
					{Code: "sor_2345", OrderStatus: OrderStatusInOrder},
					{Code: "sor_3456", OrderStatus: OrderStatusInOrder},
				}},
			},
			want1: []*MarginOrder{
				{Code: "sor_1234", OrderStatus: OrderStatusInOrder},
				{Code: "sor_2345", OrderStatus: OrderStatusInOrder},
				{Code: "sor_3456", OrderStatus: OrderStatusInOrder},
			},
			want2: nil},
		{name: "storeにある注文が死んだ注文なら返さない",
			security: virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local)},
				marginService: &testMarginService{getMarginOrders1: []*marginOrder{
					{Code: "sor_1234", OrderStatus: OrderStatusDone},
					{Code: "sor_2345", OrderStatus: OrderStatusInOrder},
					{Code: "sor_3456", OrderStatus: OrderStatusCanceled},
				}},
			},
			want1: []*MarginOrder{
				{Code: "sor_2345", OrderStatus: OrderStatusInOrder},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.security.MarginOrders()
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_virtualSecurity_MarginPositions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security virtualSecurity
		want1    []*MarginPosition
		want2    error
	}{
		{name: "storeにデータが無ければ空配列を返す",
			security: virtualSecurity{marginService: &testMarginService{
				getMarginPositions1: []*marginPosition{},
			}},
			want1: []*MarginPosition{},
			want2: nil},
		{name: "storeにあるデータをMarginPositionに詰め替えて返す",
			security: virtualSecurity{marginService: &testMarginService{
				getMarginPositions1: []*marginPosition{
					{
						Code:               "spo_1234",
						OrderCode:          "sor_0123",
						SymbolCode:         "1234",
						Side:               SideBuy,
						ContractedQuantity: 300,
						OwnedQuantity:      300,
						HoldQuantity:       100,
						Price:              1000,
						ContractedAt:       time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local),
					},
				},
			}},
			want1: []*MarginPosition{
				{
					Code:               "spo_1234",
					OrderCode:          "sor_0123",
					SymbolCode:         "1234",
					Side:               SideBuy,
					ContractedQuantity: 300,
					OwnedQuantity:      300,
					HoldQuantity:       100,
					Price:              1000,
					ContractedAt:       time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local),
				},
			},
			want2: nil},
		{name: "storeに複数データがあれば全部返す",
			security: virtualSecurity{marginService: &testMarginService{
				getMarginPositions1: []*marginPosition{
					{Code: "spo_1234", OwnedQuantity: 100},
					{Code: "spo_2345", OwnedQuantity: 100},
					{Code: "spo_3456", OwnedQuantity: 100},
				},
			}},
			want1: []*MarginPosition{
				{Code: "spo_1234", OwnedQuantity: 100},
				{Code: "spo_2345", OwnedQuantity: 100},
				{Code: "spo_3456", OwnedQuantity: 100},
			},
			want2: nil},
		{name: "storeのデータが死んでいたら返さない",
			security: virtualSecurity{marginService: &testMarginService{
				getMarginPositions1: []*marginPosition{
					{Code: "spo_1234", OwnedQuantity: 0},
					{Code: "spo_2345", OwnedQuantity: 100},
					{Code: "spo_3456", OwnedQuantity: 0},
				},
			}},
			want1: []*MarginPosition{
				{Code: "spo_2345", OwnedQuantity: 100},
			},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.security.MarginPositions()
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_virtualSecurity_CancelMarginOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		security *virtualSecurity
		arg      *CancelOrderRequest
		want     error
	}{
		{name: "注文がなければstoreからエラーが返されるので、そのエラーかラップしたエラーを返す",
			security: &virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				marginService: &testMarginService{
					getMarginOrderByCode1: nil,
					getMarginOrderByCode2: NoDataError,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: NoDataError},
		{name: "引数がnilならエラー",
			security: &virtualSecurity{
				clock:         &testClock{now1: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				marginService: &testMarginService{}},
			arg:  nil,
			want: NilArgumentError},
		{name: "キャンセル不可な状態の注文ならエラー",
			security: &virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				marginService: &testMarginService{
					getMarginOrderByCode1: &marginOrder{Code: "sor_1234", OrderStatus: OrderStatusCanceled},
					getMarginOrderByCode2: nil,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: UncancellableOrderError},
		{name: "キャンセル可能な注文ならエラーなし",
			security: &virtualSecurity{
				clock: &testClock{now1: time.Date(2021, 6, 17, 10, 0, 0, 0, time.Local)},
				marginService: &testMarginService{
					getMarginOrderByCode1: &marginOrder{Code: "sor_1234", OrderStatus: OrderStatusInOrder},
					getMarginOrderByCode2: nil,
				}},
			arg:  &CancelOrderRequest{OrderCode: "sor_1234"},
			want: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.security.CancelMarginOrder(test.arg)
			if !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_virtualSecurity_MarginOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                       string
		clock                      *testClock
		priceService               *testPriceService
		marginService              *testMarginService
		arg                        *MarginOrderRequest
		want1                      *OrderResult
		want2                      error
		wantSaveMarginOrderHistory []*marginOrder
		wantConfirmContract        int
		wantHoldExitOrderPositions int
	}{
		{name: "引数がnilであればエラーを返す", marginService: &testMarginService{}, arg: nil, want1: nil, want2: NilArgumentError},
		{name: "validationでエラーがあればエラーを返す",
			clock:         &testClock{now1: time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local)},
			marginService: &testMarginService{toMarginOrder1: &marginOrder{}, validation1: InvalidTradeTypeError},
			arg:           &MarginOrderRequest{},
			want1:         nil,
			want2:         InvalidTradeTypeError},
		{name: "該当銘柄の価格情報を取得し、価格情報がなくエラーが返されたらエラーを返す",
			clock:                      &testClock{now1: time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local)},
			priceService:               &testPriceService{getBySymbolCode1: nil, getBySymbolCode2: InvalidSymbolCodeError},
			marginService:              &testMarginService{toMarginOrder1: &marginOrder{Code: "sor-1", TradeType: TradeTypeEntry}},
			arg:                        &MarginOrderRequest{TradeType: TradeTypeEntry},
			want1:                      nil,
			want2:                      InvalidSymbolCodeError,
			wantSaveMarginOrderHistory: []*marginOrder{{Code: "sor-1", TradeType: TradeTypeEntry}}},
		{name: "該当銘柄の価格情報を取得し、価格情報がなければ、注文の保存を行ない、注文結果を返す",
			clock:                      &testClock{now1: time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local)},
			priceService:               &testPriceService{getBySymbolCode1: nil, getBySymbolCode2: NoDataError},
			marginService:              &testMarginService{toMarginOrder1: &marginOrder{Code: "sor-1", TradeType: TradeTypeEntry}},
			arg:                        &MarginOrderRequest{TradeType: TradeTypeEntry},
			want1:                      &OrderResult{OrderCode: "sor-1"},
			want2:                      nil,
			wantSaveMarginOrderHistory: []*marginOrder{{Code: "sor-1", TradeType: TradeTypeEntry}}},
		{name: "該当銘柄の価格情報を取得できたEntry注文なら、entryの約定確認をする",
			clock: &testClock{now1: time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local), getStockSession1: SessionMorning},
			priceService: &testPriceService{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Bid:        1000,
				BidTime:    time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local),
				Ask:        1000,
				AskTime:    time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			marginService:              &testMarginService{toMarginOrder1: &marginOrder{Code: "sor-1", TradeType: TradeTypeEntry}},
			arg:                        &MarginOrderRequest{TradeType: TradeTypeEntry},
			want1:                      &OrderResult{OrderCode: "sor-1"},
			want2:                      nil,
			wantConfirmContract:        1,
			wantSaveMarginOrderHistory: []*marginOrder{{Code: "sor-1", TradeType: TradeTypeEntry}}},
		{name: "該当銘柄の価格情報を取得できたEntry注文なら、entryの約定確認をし、エラーがあってもエラーは返さない",
			clock: &testClock{now1: time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local), getStockSession1: SessionMorning},
			priceService: &testPriceService{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Bid:        1000,
				BidTime:    time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local),
				Ask:        1000,
				AskTime:    time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			marginService:              &testMarginService{toMarginOrder1: &marginOrder{Code: "sor-1", TradeType: TradeTypeEntry}, confirmContract1: NilArgumentError},
			arg:                        &MarginOrderRequest{TradeType: TradeTypeEntry},
			want1:                      &OrderResult{OrderCode: "sor-1"},
			want2:                      nil,
			wantConfirmContract:        1,
			wantSaveMarginOrderHistory: []*marginOrder{{Code: "sor-1", TradeType: TradeTypeEntry}}},
		{name: "該当銘柄の価格情報を取得できたExit注文なら、exitの約定確認をする",
			clock: &testClock{now1: time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local), getStockSession1: SessionMorning},
			priceService: &testPriceService{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Bid:        1000,
				BidTime:    time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local),
				Ask:        1000,
				AskTime:    time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			marginService:              &testMarginService{toMarginOrder1: &marginOrder{Code: "sor-1", TradeType: TradeTypeExit}},
			arg:                        &MarginOrderRequest{TradeType: TradeTypeExit},
			want1:                      &OrderResult{OrderCode: "sor-1"},
			want2:                      nil,
			wantConfirmContract:        1,
			wantHoldExitOrderPositions: 1,
			wantSaveMarginOrderHistory: []*marginOrder{{Code: "sor-1", TradeType: TradeTypeExit}}},
		{name: "該当銘柄の価格情報を取得できたExit注文なら、exitの約定確認をし、エラーがあってもエラーは返さない",
			clock: &testClock{now1: time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local), getStockSession1: SessionMorning},
			priceService: &testPriceService{getBySymbolCode1: &symbolPrice{
				SymbolCode: "1234",
				Bid:        1000,
				BidTime:    time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local),
				Ask:        1000,
				AskTime:    time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local),
				kind:       PriceKindRegular,
			}, getBySymbolCode2: nil},
			marginService:              &testMarginService{toMarginOrder1: &marginOrder{Code: "sor-1", TradeType: TradeTypeExit}, confirmContract1: NilArgumentError},
			arg:                        &MarginOrderRequest{TradeType: TradeTypeExit},
			want1:                      &OrderResult{OrderCode: "sor-1"},
			want2:                      nil,
			wantConfirmContract:        1,
			wantHoldExitOrderPositions: 1,
			wantSaveMarginOrderHistory: []*marginOrder{{Code: "sor-1", TradeType: TradeTypeExit}}},
		{name: "exit注文でもholdに失敗したらエラー",
			clock:                      &testClock{now1: time.Date(2021, 8, 23, 10, 0, 0, 0, time.Local), getStockSession1: SessionMorning},
			priceService:               &testPriceService{},
			marginService:              &testMarginService{toMarginOrder1: &marginOrder{Code: "sor-1", TradeType: TradeTypeExit}, holdExitOrderPositions1: InvalidExitPositionError},
			arg:                        &MarginOrderRequest{TradeType: TradeTypeExit},
			want1:                      nil,
			want2:                      InvalidExitPositionError,
			wantHoldExitOrderPositions: 1},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			security := &virtualSecurity{clock: test.clock, marginService: test.marginService, priceService: test.priceService}
			got1, got2 := security.MarginOrder(test.arg)
			if !reflect.DeepEqual(test.want1, got1) ||
				!errors.Is(got2, test.want2) ||
				!reflect.DeepEqual(test.wantConfirmContract, test.marginService.confirmContractCount) ||
				!reflect.DeepEqual(test.wantHoldExitOrderPositions, test.marginService.holdExitOrderPositionsCount) ||
				!reflect.DeepEqual(test.wantSaveMarginOrderHistory, test.marginService.saveMarginOrderHistory) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v, %+v, %+v\ngot: %+v, %+v, %+v, %+v, %+v\n", t.Name(),
					test.want1, test.want2, test.wantConfirmContract, test.wantHoldExitOrderPositions, test.wantSaveMarginOrderHistory,
					got1, got2, test.marginService.confirmContractCount, test.marginService.holdExitOrderPositionsCount, test.marginService.saveMarginOrderHistory)
			}
		})
	}
}
