package virtual_security

import (
	"reflect"
	"testing"
	"time"
)

func Test_StockOrder_isContractableTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *StockOrder
		arg        Session
		want       bool
	}{
		{name: "場が前場でザラバで約定する注文であればtrue",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMO},
			arg:        SessionMorning,
			want:       true},
		{name: "場が前場で前場の寄りで約定する注文であればtrue",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMOMO},
			arg:        SessionMorning,
			want:       true},
		{name: "場が前場で前場の引けで約定する注文であればtrue",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMOMC},
			arg:        SessionMorning,
			want:       true},
		{name: "場が前場で後場の寄りで約定する注文であればfalse",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMOAO},
			arg:        SessionMorning,
			want:       false},
		{name: "場が前場で前場の引けで約定する注文であればfalse",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMOAC},
			arg:        SessionMorning,
			want:       false},
		{name: "場が後場でザラバで約定する注文であればtrue",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMO},
			arg:        SessionAfternoon,
			want:       true},
		{name: "場が後場で前場の寄りで約定する注文であればfalse",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMOMO},
			arg:        SessionAfternoon,
			want:       false},
		{name: "場が後場で前場の引けで約定する注文であればtrue",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMOMC},
			arg:        SessionAfternoon,
			want:       false},
		{name: "場が後場で後場の寄りで約定する注文であればfalse",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMOAO},
			arg:        SessionAfternoon,
			want:       true},
		{name: "場が後場で前場の引けで約定する注文であればfalse",
			stockOrder: &StockOrder{ExecutionCondition: StockExecutionConditionMOAC},
			arg:        SessionAfternoon,
			want:       true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.isContractableTime(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockOrder_confirmContractAuctionMO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *StockOrder
		arg1       SymbolPrice
		arg2       time.Time
		want       *confirmContractResult
	}{
		{name: "現値がなく、買い注文なら、売り気配値で約定する",
			stockOrder: &StockOrder{Side: SideBuy},
			arg1:       SymbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文でも、売り気配値がなければ約定しない",
			stockOrder: &StockOrder{Side: SideBuy},
			arg1:       SymbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文なら、買い気配値で約定する",
			stockOrder: &StockOrder{Side: SideSell},
			arg1:       SymbolPrice{Ask: 900},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        900,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文でも、買い気配値がなければ約定しない",
			stockOrder: &StockOrder{Side: SideSell},
			arg1:       SymbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があっても、現値時刻が5s以内でなければ約定しない",
			stockOrder: &StockOrder{Side: SideSell},
			arg1:       SymbolPrice{Price: 1100, PriceTime: time.Date(2021, 5, 12, 10, 59, 55, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があって、現値時刻が5s以内なら、現値で約定する",
			stockOrder: &StockOrder{Side: SideSell},
			arg1:       SymbolPrice{Price: 1100, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1100,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.confirmContractAuctionMO(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockOrder_confirmContractRegularMO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *StockOrder
		arg1       SymbolPrice
		arg2       time.Time
		want       *confirmContractResult
	}{
		{name: "買い注文なら、売り気配値で約定する",
			stockOrder: &StockOrder{Side: SideBuy},
			arg1:       SymbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "買い注文でも、売り気配値がなければ約定しない",
			stockOrder: &StockOrder{Side: SideBuy},
			arg1:       SymbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "売り注文なら、買い気配値で約定する",
			stockOrder: &StockOrder{Side: SideSell},
			arg1:       SymbolPrice{Ask: 900},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        900,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "売り注文でも、買い気配値がなければ約定しない",
			stockOrder: &StockOrder{Side: SideSell},
			arg1:       SymbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.confirmContractRegularMO(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockOrder_confirmContractAuctionLO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *StockOrder
		arg1       SymbolPrice
		arg2       time.Time
		want       *confirmContractResult
	}{
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値より高いなら、売り気配値で約定する",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 1001},
			arg1:       SymbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値と同じなら、売り気配値で約定する",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 1000},
			arg1:       SymbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値より安いなら、約定しない",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       SymbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、買い注文で、売り気配値がなければ、約定しない",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       SymbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値より高いなら、約定しない",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 1001},
			arg1:       SymbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値と同じなら、買い気配値で約定する",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 1000},
			arg1:       SymbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値より安いなら、買い気配値で約定する",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       SymbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文で、買い気配値がなければ、約定しない",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       SymbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s前なら、約定しない",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 1000},
			arg1:       SymbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 55, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値より高いなら、現値で約定する",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 1001},
			arg1:       SymbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値と同じなら、現値で約定する",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 1000},
			arg1:       SymbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値より安いなら、約定しない",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       SymbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値より高いなら、約定しない",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 1001},
			arg1:       SymbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値と同じなら、現値で約定する",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 1000},
			arg1:       SymbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値より安いなら、現値で約定する",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       SymbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.confirmContractAuctionLO(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockOrder_confirmContractRegularLO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *StockOrder
		arg1       SymbolPrice
		arg2       time.Time
		want       *confirmContractResult
	}{
		{name: "買い注文で、売り気配値があり、指値が売り気配値より高いなら、指値で約定する",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 1001},
			arg1:       SymbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値と同じなら、約定しない",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 1000},
			arg1:       SymbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値より安いなら、約定しない",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       SymbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "買い注文で、売り気配値がなければ、約定しない",
			stockOrder: &StockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       SymbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値より高いなら、約定しない",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 1001},
			arg1:       SymbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値と同じなら、約定しない",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 1000},
			arg1:       SymbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値より安いなら、指値で約定する",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       SymbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "売り注文で、買い気配値がなければ、約定しない",
			stockOrder: &StockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       SymbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.confirmContractRegularLO(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockOrder_confirmContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *StockOrder
		arg1       SymbolPrice
		arg2       time.Time
		arg3       Session
		want       *confirmContractResult
	}{
		{name: "銘柄が一致していなければfalse",
			stockOrder: &StockOrder{SymbolCode: "1234"},
			arg1:       SymbolPrice{SymbolCode: "0000"},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionUnspecified,
			want:       &confirmContractResult{isContracted: false}},
		{name: "銘柄が一致していても市場が一致していなければfalse",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionUnspecified,
			want:       &confirmContractResult{isContracted: false}},
		{name: "注文が約定できない状態ならfalse",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusDone},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionUnspecified,
			want:       &confirmContractResult{isContracted: false}},
		{name: "約定できない時間ならfalse",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionUnspecified,
			want:       &confirmContractResult{isContracted: false}},
		{name: "成行が寄り価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "成行が引け価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "成行がザラバで約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 1000, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "成行がタイミング不明なら約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 1000, kind: PriceKindUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄成前場が寄り価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄成前場が2回目以降の確認では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMO, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄成前場が後場では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMO, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄成後場が寄り価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄成後場が2回目以降の確認では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAO, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄成後場が前場では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAO, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引成前場が引け価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMC},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local)}},
		{name: "引成前場が2回目以降の確認では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMC, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引成前場が後場では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMC, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引成後場が引け価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAC},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "引成後場が2回目以降の確認では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAC, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引成後場が前場では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAC, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false},
		},
		{name: "IOC成行が寄り価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionIOCMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC成行が引け価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionIOCMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "IOC成行がザラバで約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 1000, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC成行がタイミング不明なら約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCMO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 1000, kind: PriceKindUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "IOC成行が1度でも約定確認されていたら約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCMO, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 1000, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "指値が寄り価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLO, LimitPrice: 1000},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "指値が引け価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLO, LimitPrice: 1000},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "指値がザラバで約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLO, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "指値がタイミング不明なら約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLO},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 1000, kind: PriceKindUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄指前場が寄り価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMO, LimitPrice: 1000},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄指前場が2回目以降の確認では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMO, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄指前場が後場では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMO, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄指後場が寄り価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAO, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄指値後場が2回目以降の確認では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMO, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄指後場が前場では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAO, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引指前場が引け価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMC, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local)}},
		{name: "引指前場が2回目以降の確認では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMC, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引指前場が後場では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMC, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引指後場が引け価格で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAC, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "引指後場が2回目以降の確認では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAC, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引指後場が前場では約定しない",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAC, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "不成前場は前場の寄りではオークションの指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			}},
		{name: "不成前場は前場のザラバでは指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			}},
		{name: "不成前場は前場の引けではオークションの成行で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1200, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        1200,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成前場は後場の寄りではオークションの指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成前場は後場のザラバでは指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成前場は後場の引けではオークションの指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 990, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			}},
		{name: "不成後場は前場の寄りではオークションの指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			}},
		{name: "不成後場は前場のザラバでは指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			}},
		{name: "不成後場は前場の引けではオークションの指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 990, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成後場は後場の寄りではオークションの指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成後場は後場のザラバでは指値で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成後場は後場の引けではオークションの成行で約定する",
			stockOrder: &StockOrder{SymbolCode: "1234", Exchange: ExchangeToushou, OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFUNARIA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       SymbolPrice{SymbolCode: "1234", Exchange: ExchangeToushou, Price: 1200, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        1200,
				contractedAt: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.confirmContract(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
