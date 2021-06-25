package virtual_security

import (
	"reflect"
	"testing"
	"time"
)

func Test_stockOrder_isContractableTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg        Session
		want       bool
	}{
		{name: "場が前場でザラバで約定する注文であればtrue",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMO},
			arg:        SessionMorning,
			want:       true},
		{name: "場が前場で前場の寄りで約定する注文であればtrue",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOMO},
			arg:        SessionMorning,
			want:       true},
		{name: "場が前場で前場の引けで約定する注文であればtrue",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOMC},
			arg:        SessionMorning,
			want:       true},
		{name: "場が前場で後場の寄りで約定する注文であればfalse",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOAO},
			arg:        SessionMorning,
			want:       false},
		{name: "場が前場で前場の引けで約定する注文であればfalse",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOAC},
			arg:        SessionMorning,
			want:       false},
		{name: "場が後場でザラバで約定する注文であればtrue",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMO},
			arg:        SessionAfternoon,
			want:       true},
		{name: "場が後場で前場の寄りで約定する注文であればfalse",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOMO},
			arg:        SessionAfternoon,
			want:       false},
		{name: "場が後場で前場の引けで約定する注文であればtrue",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOMC},
			arg:        SessionAfternoon,
			want:       false},
		{name: "場が後場で後場の寄りで約定する注文であればfalse",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOAO},
			arg:        SessionAfternoon,
			want:       true},
		{name: "場が後場で前場の引けで約定する注文であればfalse",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOAC},
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

func Test_stockOrder_confirmContractItayoseMO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg1       *symbolPrice
		arg2       time.Time
		want       *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			stockOrder: &stockOrder{Side: SideBuy},
			arg1:       nil,
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、買い注文なら、売り気配値で約定する",
			stockOrder: &stockOrder{Side: SideBuy},
			arg1:       &symbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文でも、売り気配値がなければ約定しない",
			stockOrder: &stockOrder{Side: SideBuy},
			arg1:       &symbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文なら、買い気配値で約定する",
			stockOrder: &stockOrder{Side: SideSell},
			arg1:       &symbolPrice{Ask: 900},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        900,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文でも、買い気配値がなければ約定しない",
			stockOrder: &stockOrder{Side: SideSell},
			arg1:       &symbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があっても、現値時刻が5s以内でなければ約定しない",
			stockOrder: &stockOrder{Side: SideSell},
			arg1:       &symbolPrice{Price: 1100, PriceTime: time.Date(2021, 5, 12, 10, 59, 55, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があって、現値時刻が5s以内なら、現値で約定する",
			stockOrder: &stockOrder{Side: SideSell},
			arg1:       &symbolPrice{Price: 1100, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
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
			got := test.stockOrder.confirmContractItayoseMO(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockOrder_confirmContractRegularMO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg1       *symbolPrice
		arg2       time.Time
		want       *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			stockOrder: &stockOrder{Side: SideBuy},
			arg1:       nil,
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "買い注文なら、売り気配値で約定する",
			stockOrder: &stockOrder{Side: SideBuy},
			arg1:       &symbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "買い注文でも、売り気配値がなければ約定しない",
			stockOrder: &stockOrder{Side: SideBuy},
			arg1:       &symbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "売り注文なら、買い気配値で約定する",
			stockOrder: &stockOrder{Side: SideSell},
			arg1:       &symbolPrice{Ask: 900},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        900,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "売り注文でも、買い気配値がなければ約定しない",
			stockOrder: &stockOrder{Side: SideSell},
			arg1:       &symbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
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

func Test_stockOrder_confirmContractItayoseLO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg1       *symbolPrice
		arg2       time.Time
		want       *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1001},
			arg1:       nil,
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値より高いなら、売り気配値で約定する",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1001},
			arg1:       &symbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値と同じなら、売り気配値で約定する",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1000},
			arg1:       &symbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値より安いなら、約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       &symbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、買い注文で、売り気配値がなければ、約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       &symbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値より高いなら、約定しない",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 1001},
			arg1:       &symbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値と同じなら、買い気配値で約定する",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 1000},
			arg1:       &symbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値より安いなら、買い気配値で約定する",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       &symbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文で、買い気配値がなければ、約定しない",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       &symbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s前なら、約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1000},
			arg1:       &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 55, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値より高いなら、現値で約定する",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1001},
			arg1:       &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値と同じなら、現値で約定する",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1000},
			arg1:       &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値より安いなら、約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値より高いなら、約定しない",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 1001},
			arg1:       &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値と同じなら、現値で約定する",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 1000},
			arg1:       &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値より安いなら、現値で約定する",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
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
			got := test.stockOrder.confirmContractItayoseLO(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockOrder_confirmContractRegularLO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg1       *symbolPrice
		arg2       time.Time
		want       *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1001},
			arg1:       nil,
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値より高いなら、指値で約定する",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1001},
			arg1:       &symbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値と同じなら、約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 1000},
			arg1:       &symbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値より安いなら、約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       &symbolPrice{Bid: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "買い注文で、売り気配値がなければ、約定しない",
			stockOrder: &stockOrder{Side: SideBuy, LimitPrice: 999},
			arg1:       &symbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値より高いなら、約定しない",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 1001},
			arg1:       &symbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値と同じなら、約定しない",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 1000},
			arg1:       &symbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値より安いなら、指値で約定する",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       &symbolPrice{Ask: 1000},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "売り注文で、買い気配値がなければ、約定しない",
			stockOrder: &stockOrder{Side: SideSell, LimitPrice: 999},
			arg1:       &symbolPrice{},
			arg2:       time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want:       &confirmContractResult{isContracted: false}},
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

func Test_stockOrder_confirmContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg1       *symbolPrice
		arg2       time.Time
		arg3       Session
		want       *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234"},
			arg1:       nil,
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "銘柄が一致していなければfalse",
			stockOrder: &stockOrder{SymbolCode: "1234"},
			arg1:       &symbolPrice{SymbolCode: "0000"},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionUnspecified,
			want:       &confirmContractResult{isContracted: false}},
		{name: "注文が約定できない状態ならfalse",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusDone},
			arg1:       &symbolPrice{SymbolCode: "1234"},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionUnspecified,
			want:       &confirmContractResult{isContracted: false}},
		{name: "約定できない時間ならfalse",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMO},
			arg1:       &symbolPrice{SymbolCode: "1234"},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionUnspecified,
			want:       &confirmContractResult{isContracted: false}},
		{name: "成行が寄り価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "成行が引け価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "成行がザラバで約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "成行がタイミング不明なら約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄成前場が寄り価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄成前場が2回目以降の確認では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMO, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄成前場が後場では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMO, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄成後場が寄り価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAO},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄成後場が2回目以降の確認では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAO, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄成後場が前場では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAO, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引成前場が引け価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMC},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local)}},
		{name: "引成前場が2回目以降の確認では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMC, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引成前場が後場では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOMC, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引成後場が引け価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAC},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "引成後場が2回目以降の確認では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAC, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引成後場が前場では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionMOAC, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false},
		},
		{name: "IOC成行が寄り価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionIOCMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC成行が引け価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, ExecutionCondition: StockExecutionConditionIOCMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "IOC成行がザラバで約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC成行がタイミング不明なら約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCMO},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "IOC成行が1度でも約定確認されていたら約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCMO, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "指値が寄り価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLO, LimitPrice: 1000},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "指値が引け価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLO, LimitPrice: 1000},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "指値がザラバで約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLO, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "指値がタイミング不明なら約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLO},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄指前場が寄り価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMO, LimitPrice: 1000},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄指前場が2回目以降の確認では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMO, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄指前場が後場では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMO, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄指後場が寄り価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAO, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄指値後場が2回目以降の確認では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMO, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "寄指後場が前場では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAO, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引指前場が引け価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMC, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local)}},
		{name: "引指前場が2回目以降の確認では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMC, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引指前場が後場では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOMC, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引指後場が引け価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAC, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "引指後場が2回目以降の確認では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAC, LimitPrice: 1000, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "引指後場が前場では約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionLOAC, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "不成前場は前場の寄りではオークションの指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			}},
		{name: "不成前場は前場のザラバでは指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			}},
		{name: "不成前場は前場の引けではオークションの成行で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1200, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        1200,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成前場は後場の寄りではオークションの指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成前場は後場のザラバでは指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成前場は後場の引けではオークションの指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariM, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 990, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			}},
		{name: "不成後場は前場の寄りではオークションの指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			}},
		{name: "不成後場は前場のザラバでは指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			}},
		{name: "不成後場は前場の引けではオークションの指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 990, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionMorning,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成後場は後場の寄りではオークションの指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成後場は後場のザラバでは指値で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        990,
				contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			}},
		{name: "不成後場は後場の引けではオークションの成行で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionFunariA, LimitPrice: 1000, ConfirmingCount: 0},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1200, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want: &confirmContractResult{
				isContracted: true,
				price:        1200,
				contractedAt: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			}},
		{name: "逆指値注文が待機中なら約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusWait, Side: SideBuy, ExecutionCondition: StockExecutionConditionStop, LimitPrice: 1000, ConfirmingCount: 0, StopCondition: &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO}},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1200, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			arg3:       SessionAfternoon,
			want:       &confirmContractResult{isContracted: false}},
		{name: "逆指値注文が注文中で、逆指値条件が成行なら成行と同じ処理が行なわれる",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionStop, StopCondition: &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO}},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "逆指値注文が注文中で、逆指値条件が指値なら指値と同じ処理が行なわれる",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionStop, ConfirmingCount: 1, StopCondition: &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionLO, LimitPriceAfterHit: 1000.0}},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC指値が寄り価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCLO, LimitPrice: 1000},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC指値が引け価格で約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCLO, LimitPrice: 1000},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg2:       time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "IOC指値がザラバで約定する",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCLO, LimitPrice: 1000},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 990, kind: PriceKindRegular},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: true, price: 990, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC指値がタイミング不明なら約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCLO},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
		{name: "IOC指値が1度でも約定確認されていたら約定しない",
			stockOrder: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionIOCLO, ConfirmingCount: 1},
			arg1:       &symbolPrice{SymbolCode: "1234", Bid: 1000, kind: PriceKindUnspecified},
			arg2:       time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			arg3:       SessionMorning,
			want:       &confirmContractResult{isContracted: false}},
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

func Test_stockOrder_contract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		stockOrder             *stockOrder
		arg                    *Contract
		wantContractedQuantity float64
		wantStatus             OrderStatus
	}{
		{name: "引数がnilなら何もしない",
			stockOrder:             &stockOrder{OrderQuantity: 3, ContractedQuantity: 0, OrderStatus: OrderStatusUnspecified},
			arg:                    nil,
			wantContractedQuantity: 0,
			wantStatus:             OrderStatusUnspecified},
		{name: "約定後、約定数量が0なら注文中",
			stockOrder:             &stockOrder{OrderQuantity: 3, ContractedQuantity: 0, OrderStatus: OrderStatusUnspecified},
			arg:                    &Contract{Quantity: 0},
			wantContractedQuantity: 0,
			wantStatus:             OrderStatusInOrder},
		{name: "約定後、約定数量が注文数量未満なら部分約定",
			stockOrder:             &stockOrder{OrderQuantity: 3, ContractedQuantity: 0, OrderStatus: OrderStatusUnspecified},
			arg:                    &Contract{Quantity: 1},
			wantContractedQuantity: 1,
			wantStatus:             OrderStatusPart},
		{name: "約定後、約定数量が注文数量以上なら全約定",
			stockOrder:             &stockOrder{OrderQuantity: 3, ContractedQuantity: 1, OrderStatus: OrderStatusUnspecified},
			arg:                    &Contract{Quantity: 2},
			wantContractedQuantity: 3,
			wantStatus:             OrderStatusDone},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.stockOrder.contract(test.arg)
			got1 := test.stockOrder.ContractedQuantity
			got2 := test.stockOrder.OrderStatus
			if !reflect.DeepEqual(test.wantContractedQuantity, got1) || !reflect.DeepEqual(test.wantStatus, got2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.wantContractedQuantity, test.wantStatus, got1, got2)
			}
		})
	}
}

func Test_stockOrder_cancel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		stockOrder      *stockOrder
		arg             time.Time
		wantOrderStatus OrderStatus
		wantCanceledAt  time.Time
	}{
		{name: "ステータスがnewなら取消状態に更新",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusNew},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled,
			wantCanceledAt:  time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local)},
		{name: "ステータスがin_orderなら取消状態に更新",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusInOrder},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled,
			wantCanceledAt:  time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local)},
		{name: "ステータスがpartなら取消状態に更新",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusPart},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled,
			wantCanceledAt:  time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local)},
		{name: "ステータスがdoneなら取消状態に更新できない",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusDone},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusDone,
			wantCanceledAt:  time.Time{}},
		{name: "ステータスがcanceledなら取消状態に更新できない",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 5, 18, 10, 0, 0, 0, time.Local)},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled,
			wantCanceledAt:  time.Date(2021, 5, 18, 10, 0, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.stockOrder.cancel(test.arg)
			got1 := test.stockOrder.OrderStatus
			got2 := test.stockOrder.CanceledAt
			if !reflect.DeepEqual(test.wantOrderStatus, got1) || !reflect.DeepEqual(test.wantCanceledAt, got2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.wantOrderStatus, test.wantCanceledAt, got1, got2)
			}
		})
	}
}

func Test_stockOrder_activate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg1       *symbolPrice
		arg2       time.Time
		wantStatus OrderStatus
	}{
		{name: "条件を満たせば注文中になる",
			stockOrder: &stockOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition: &StockStopCondition{
					StopPrice:                  100.0,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				}},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local)},
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusInOrder},
		{name: "現在値の時間が5s前以前なら有効にならない",
			stockOrder: &stockOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition: &StockStopCondition{
					StopPrice:                  100.0,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				}},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 30, 20, 31, 54, 0, time.Local)},
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusWait},
		{name: "価格情報がなければ何もしない",
			stockOrder: &stockOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition: &StockStopCondition{
					StopPrice:                  100.0,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				}},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 0},
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusWait},
		{name: "逆指値条件が設定されていなければ何もしない",
			stockOrder: &stockOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local)},
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusWait},
		{name: "逆指値注文でなければ何もしない",
			stockOrder: &stockOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionMO,
				StopCondition: &StockStopCondition{
					StopPrice:                  100.0,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				}},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local)},
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusWait},
		{name: "注文の状態が待機でなければ何もしない",
			stockOrder: &stockOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusPart,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition: &StockStopCondition{
					StopPrice:                  100.0,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				}},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local)},
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusPart},
		{name: "銘柄が違えば何もしない",
			stockOrder: &stockOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition: &StockStopCondition{
					StopPrice:                  100.0,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				}},
			arg1:       &symbolPrice{SymbolCode: "0000", Price: 1000, PriceTime: time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local)},
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusWait},
		{name: "引数がnilなら何もしない",
			stockOrder: &stockOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition: &StockStopCondition{
					StopPrice:                  100.0,
					ComparisonOperator:         ComparisonOperatorLE,
					ExecutionConditionAfterHit: StockExecutionConditionMO,
				}},
			arg1:       nil,
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusWait},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.stockOrder.activate(test.arg1, test.arg2)
			got1 := test.stockOrder.OrderStatus
			if !reflect.DeepEqual(test.wantStatus, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantStatus, got1)
			}
		})
	}
}

func Test_stockOrder_executionCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		want       StockExecutionCondition
	}{
		{name: "逆指値で待機中でなく逆指値条件があれば、逆指値発動後の条件が返される",
			stockOrder: &stockOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO}},
			want: StockExecutionConditionMO},
		{name: "逆指値で待機中でなくても、逆指値条件がなければそのまま返す",
			stockOrder: &stockOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      nil},
			want: StockExecutionConditionStop},
		{name: "逆指値でも待機中なら、そのまま返す",
			stockOrder: &stockOrder{
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO}},
			want: StockExecutionConditionStop},
		{name: "逆指値注文でなければそのまま返す",
			stockOrder: &stockOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionLO,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO}},
			want: StockExecutionConditionLO},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.executionCondition()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockOrder_limitPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		want       float64
	}{
		{name: "逆指値で待機中でなく逆指値条件があれば、逆指値発動後の指値価格が返される",
			stockOrder: &stockOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionStop,
				LimitPrice:         1000,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO, LimitPriceAfterHit: 1500}},
			want: 1500},
		{name: "逆指値で待機中でなくても、逆指値条件がなければそのまま返す",
			stockOrder: &stockOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      nil},
			want: 0},
		{name: "逆指値でも待機中なら、そのまま返す",
			stockOrder: &stockOrder{
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO, LimitPriceAfterHit: 1500}},
			want: 0},
		{name: "逆指値注文でなければそのまま返す",
			stockOrder: &stockOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionLO,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO, LimitPriceAfterHit: 1500}},
			want: 0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.limitPrice()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockOrder_expired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		stockOrder      *stockOrder
		arg             time.Time
		wantOrderStatus OrderStatus
	}{
		{name: "有効期限がゼロ値なら何もしない",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusInOrder, ExpiredAt: time.Time{}},
			arg:             time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local),
			wantOrderStatus: OrderStatusInOrder},
		{name: "有効期限が現在時刻よりも過去なら取消済みにする",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusInOrder, ExpiredAt: time.Date(2021, 6, 7, 13, 0, 0, 0, time.Local)},
			arg:             time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled},
		{name: "有効期限が現在時刻と一致しているなら状態を変えない",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusInOrder, ExpiredAt: time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local)},
			arg:             time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local),
			wantOrderStatus: OrderStatusInOrder},
		{name: "有効期限が現在時刻よりも未来なら状態を変えない",
			stockOrder:      &stockOrder{OrderStatus: OrderStatusInOrder, ExpiredAt: time.Date(2021, 6, 7, 15, 0, 0, 0, time.Local)},
			arg:             time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local),
			wantOrderStatus: OrderStatusInOrder},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.stockOrder.expired(test.arg)
			if !reflect.DeepEqual(test.wantOrderStatus, test.stockOrder.OrderStatus) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantOrderStatus, test.stockOrder.OrderStatus)
			}
		})
	}
}

func Test_stockOrder_isDied(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg        time.Time
		want       bool
	}{
		{name: "未終了の注文なら生きている",
			stockOrder: &stockOrder{OrderStatus: OrderStatusInOrder},
			arg:        time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:       false},
		{name: "取消済み注文で、取消から1日以内なら生きている",
			stockOrder: &stockOrder{OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 6, 14, 11, 0, 0, 0, time.Local)},
			arg:        time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:       false},
		{name: "取消済み注文で、取消から1日丁度なら生きている",
			stockOrder: &stockOrder{OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local)},
			arg:        time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:       false},
		{name: "取消済み注文で、取消から1日以上経っていたら死んでいる",
			stockOrder: &stockOrder{OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local)},
			arg:        time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:       true},
		{name: "約定済み注文で、最後の約定から1日以内なら生きている",
			stockOrder: &stockOrder{OrderStatus: OrderStatusDone, Contracts: []*Contract{
				{ContractedAt: time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local)},
				{ContractedAt: time.Date(2021, 6, 14, 11, 0, 0, 0, time.Local)}}},
			arg:  time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want: false},
		{name: "約定済み注文で、最後の約定から1日丁度なら生きている",
			stockOrder: &stockOrder{OrderStatus: OrderStatusDone, Contracts: []*Contract{
				{ContractedAt: time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local)},
				{ContractedAt: time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local)}}},
			arg:  time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want: false},
		{name: "約定済み注文で、最後の約定から1日以上経っていたら死んでいる",
			stockOrder: &stockOrder{OrderStatus: OrderStatusDone, Contracts: []*Contract{
				{ContractedAt: time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local)},
				{ContractedAt: time.Date(2021, 6, 14, 9, 30, 0, 0, time.Local)}}},
			arg:  time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want: true},
		{name: "終了した注文で、取消も約定も情報が無かったら死んだものとする",
			stockOrder: &stockOrder{OrderStatus: OrderStatusDone, Contracts: []*Contract{}},
			arg:        time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:       true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockOrder.isDied(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockOrder_isValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		order *stockOrder
		arg   time.Time
		want  error
	}{
		{name: "方向が不明ならエラー",
			order: &stockOrder{
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
			},
			want: InvalidSideError},
		{name: "執行条件が不明ならエラー",
			order: &stockOrder{
				Side:          SideBuy,
				SymbolCode:    "1234",
				OrderQuantity: 100,
			},
			want: InvalidExecutionConditionError},
		{name: "銘柄がゼロ値ならエラー",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				OrderQuantity:      100,
			},
			want: InvalidSymbolCodeError},
		{name: "数量がゼロ値ならエラー",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
			},
			want: InvalidQuantityError},
		{name: "指値を指定して指値価格がゼロ値ならエラー",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionLO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
			},
			want: InvalidLimitPriceError},
		{name: "指値を指定して指値価格があればエラーなし",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionLO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				LimitPrice:         1000,
			},
			want: nil},
		{name: "有効期限が過去ならエラー",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 6, 18, 15, 0, 0, 0, time.Local),
			},
			arg:  time.Date(2021, 6, 18, 16, 0, 0, 0, time.Local),
			want: InvalidExpiredError},
		{name: "逆指値で逆指値条件が設定されていなければエラー",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
			},
			want: InvalidStopConditionError},
		{name: "逆指値でトリガー価格が設定されていなければエラー",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				StopCondition:      &StockStopCondition{},
			},
			want: InvalidStopConditionError},
		{name: "逆指値でトリガー後の執行条件が逆指値ならエラー",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				StopCondition:      &StockStopCondition{StopPrice: 2000, ExecutionConditionAfterHit: StockExecutionConditionStop},
			},
			want: InvalidStopConditionError},
		{name: "逆指値でトリガー後の執行条件が指値で指値価格が指定されていなければエラー",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				StopCondition:      &StockStopCondition{StopPrice: 2000, ExecutionConditionAfterHit: StockExecutionConditionLO},
			},
			want: InvalidStopConditionError},
		{name: "逆指値でトリガー後の執行条件が指値で指値価格が指定されていればエラーなし",
			order: &stockOrder{
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				StopCondition:      &StockStopCondition{StopPrice: 2000, ExecutionConditionAfterHit: StockExecutionConditionLO, LimitPriceAfterHit: 2100},
			},
			want: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.order.isValid(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
