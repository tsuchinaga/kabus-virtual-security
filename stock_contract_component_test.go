package virtual_security

import (
	"reflect"
	"testing"
	"time"
)

type testStockContractComponent struct {
	iStockContractComponent
	isContractableTime1         bool
	confirmStockOrderContract1  *confirmContractResult
	confirmMarginOrderContract1 *confirmContractResult
}

func (t *testStockContractComponent) isContractableTime(StockExecutionCondition, time.Time) bool {
	return t.isContractableTime1
}
func (t *testStockContractComponent) confirmStockOrderContract(*stockOrder, *symbolPrice, time.Time) *confirmContractResult {
	return t.confirmStockOrderContract1
}
func (t *testStockContractComponent) confirmMarginOrderContract(*marginOrder, *symbolPrice, time.Time) *confirmContractResult {
	return t.confirmMarginOrderContract1
}

func Test_newStockContractComponent(t *testing.T) {
	t.Parallel()
	want := &stockContractComponent{}
	got := newStockContractComponent()

	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_stockContractComponent_isContractableTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		stockOrder *stockOrder
		arg1       StockExecutionCondition
		arg2       time.Time
		want       bool
	}{
		{name: "場が前場でザラバで約定する注文であればtrue",
			arg1: StockExecutionConditionMO,
			arg2: time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
			want: true},
		{name: "場が前場で前場の寄りで約定する注文であればtrue",
			stockOrder: &stockOrder{ExecutionCondition: StockExecutionConditionMOMO},
			arg1:       StockExecutionConditionMOMO,
			arg2:       time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
			want:       true},
		{name: "場が前場で前場の引けで約定する注文であればtrue",
			arg1: StockExecutionConditionMOMC,
			arg2: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local),
			want: true},
		{name: "場が前場で後場の寄りで約定する注文であればfalse",
			arg1: StockExecutionConditionMOAO,
			arg2: time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
			want: false},
		{name: "場が前場で前場の引けで約定する注文であればfalse",
			arg1: StockExecutionConditionMOAC,
			arg2: time.Date(0, 1, 1, 10, 0, 0, 0, time.Local),
			want: false},
		{name: "場が後場でザラバで約定する注文であればtrue",
			arg1: StockExecutionConditionMO,
			arg2: time.Date(0, 1, 1, 14, 0, 0, 0, time.Local),
			want: true},
		{name: "場が後場で前場の寄りで約定する注文であればfalse",
			arg1: StockExecutionConditionMOMO,
			arg2: time.Date(0, 1, 1, 14, 0, 0, 0, time.Local),
			want: false},
		{name: "場が後場で前場の引けで約定する注文であればtrue",
			arg1: StockExecutionConditionMOMC,
			arg2: time.Date(0, 1, 1, 14, 0, 0, 0, time.Local),
			want: false},
		{name: "場が後場で後場の寄りで約定する注文であればfalse",
			arg1: StockExecutionConditionMOAO,
			arg2: time.Date(0, 1, 1, 14, 0, 0, 0, time.Local),
			want: true},
		{name: "場が後場で後場の引けで約定する注文であればtrue",
			arg1: StockExecutionConditionMOAC,
			arg2: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
			want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := &stockContractComponent{}
			got := component.isContractableTime(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockContractComponent_confirmContractItayoseMO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 Side
		arg2 *symbolPrice
		arg3 time.Time
		want *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			arg1: SideBuy,
			arg2: nil,
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値がなく、買い注文なら、売り気配値で約定する",
			arg1: SideBuy,
			arg2: &symbolPrice{Ask: 1000},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文でも、売り気配値がなければ約定しない",
			arg1: SideBuy,
			arg2: &symbolPrice{},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文なら、買い気配値で約定する",
			arg1: SideSell,
			arg2: &symbolPrice{Bid: 900},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        900,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文でも、買い気配値がなければ約定しない",
			arg1: SideSell,
			arg2: &symbolPrice{},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値があっても、現値時刻が5s以内でなければ約定しない",
			arg1: SideSell,
			arg2: &symbolPrice{Price: 1100, PriceTime: time.Date(2021, 5, 12, 10, 59, 55, 0, time.Local)},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値があって、現値時刻が5s以内なら、現値で約定する",
			arg1: SideSell,
			arg2: &symbolPrice{Price: 1100, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
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
			component := &stockContractComponent{}
			got := component.confirmContractItayoseMO(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockContractComponent_confirmContractAuctionMO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 Side
		arg2 *symbolPrice
		arg3 time.Time
		want *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			arg1: SideBuy,
			arg2: nil,
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "買い注文なら、売り気配値で約定する",
			arg1: SideBuy,
			arg2: &symbolPrice{Ask: 1000},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "買い注文でも、売り気配値がなければ約定しない",
			arg1: SideBuy,
			arg2: &symbolPrice{},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "売り注文なら、買い気配値で約定する",
			arg1: SideSell,
			arg2: &symbolPrice{Bid: 900},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        900,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "売り注文でも、買い気配値がなければ約定しない",
			arg1: SideSell,
			arg2: &symbolPrice{},
			arg3: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := &stockContractComponent{}
			got := component.confirmContractAuctionMO(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockContractComponent_confirmContractItayoseLO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 Side
		arg2 float64
		arg3 *symbolPrice
		arg4 time.Time
		want *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			arg1: SideBuy,
			arg2: 1001,
			arg3: nil,
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値より高いなら、売り気配値で約定する",
			arg1: SideBuy,
			arg2: 1001,
			arg3: &symbolPrice{Ask: 1000},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値と同じなら、売り気配値で約定する",
			arg1: SideBuy,
			arg2: 1000,
			arg3: &symbolPrice{Ask: 1000},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、買い注文で、売り気配値があり、指値が売り気配値より安いなら、約定しない",
			arg1: SideBuy,
			arg2: 999,
			arg3: &symbolPrice{Ask: 1000},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値がなく、買い注文で、売り気配値がなければ、約定しない",
			arg1: SideBuy,
			arg2: 999,
			arg3: &symbolPrice{},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値より高いなら、約定しない",
			arg1: SideSell,
			arg2: 1001,
			arg3: &symbolPrice{Bid: 1000},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値と同じなら、買い気配値で約定する",
			arg1: SideSell,
			arg2: 1000,
			arg3: &symbolPrice{Bid: 1000},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文で、買い気配値があり、指値が買い気配値より安いなら、買い気配値で約定する",
			arg1: SideSell,
			arg2: 999,
			arg3: &symbolPrice{Bid: 1000},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値がなく、売り注文で、買い気配値がなければ、約定しない",
			arg1: SideSell,
			arg2: 999,
			arg3: &symbolPrice{},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s前なら、約定しない",
			arg1: SideBuy,
			arg2: 1000,
			arg3: &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 55, 0, time.Local)},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値より高いなら、現値で約定する",
			arg1: SideBuy,
			arg2: 1001,
			arg3: &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値と同じなら、現値で約定する",
			arg1: SideBuy,
			arg2: 1000,
			arg3: &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、買い注文で、指値が現値より安いなら、約定しない",
			arg1: SideBuy,
			arg2: 999,
			arg3: &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値より高いなら、約定しない",
			arg1: SideSell,
			arg2: 1001,
			arg3: &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値と同じなら、現値で約定する",
			arg1: SideSell,
			arg2: 1000,
			arg3: &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "現値があり、現値時刻が5s以内で、売り注文で、指値が現値より安いなら、現値で約定する",
			arg1: SideSell,
			arg2: 999,
			arg3: &symbolPrice{Price: 1000, PriceTime: time.Date(2021, 5, 12, 10, 59, 56, 0, time.Local)},
			arg4: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
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
			component := &stockContractComponent{}
			got := component.confirmContractItayoseLO(test.arg1, test.arg2, test.arg3, test.arg4)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockContractComponent_confirmContractAuctionLO(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 Side
		arg2 float64
		arg3 bool
		arg4 *symbolPrice
		arg5 time.Time
		want *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			arg1: SideBuy,
			arg2: 1001,
			arg3: true,
			arg4: nil,
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値より高いなら、指値で約定する",
			arg1: SideBuy,
			arg2: 1001,
			arg3: true,
			arg4: &symbolPrice{Ask: 1000},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1001,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値より高く、初回約定確認なら、気配値で約定する",
			arg1: SideBuy,
			arg2: 1001,
			arg3: false,
			arg4: &symbolPrice{Ask: 1000},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値と同じなら、約定しない",
			arg1: SideBuy,
			arg2: 1000,
			arg3: true,
			arg4: &symbolPrice{Ask: 1000},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "買い注文で、売り気配値があり、指値が売り気配値より安いなら、約定しない",
			arg1: SideBuy,
			arg2: 999,
			arg3: true,
			arg4: &symbolPrice{Ask: 1000},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "買い注文で、売り気配値がなければ、約定しない",
			arg1: SideBuy,
			arg2: 999,
			arg3: true,
			arg4: &symbolPrice{},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値より高いなら、約定しない",
			arg1: SideSell,
			arg2: 1001,
			arg3: true,
			arg4: &symbolPrice{Bid: 1000},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値と同じなら、約定しない",
			arg1: SideSell,
			arg2: 1000,
			arg3: true,
			arg4: &symbolPrice{Bid: 1000},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値より安いなら、指値で約定する",
			arg1: SideSell,
			arg2: 999,
			arg3: true,
			arg4: &symbolPrice{Bid: 1000},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        999,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "売り注文で、買い気配値があり、指値が買い気配値でより安く、初回約定確認なら、板で約定する",
			arg1: SideSell,
			arg2: 999,
			arg3: false,
			arg4: &symbolPrice{Bid: 1000},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{
				isContracted: true,
				price:        1000,
				contractedAt: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			}},
		{name: "売り注文で、買い気配値がなければ、約定しない",
			arg1: SideSell,
			arg2: 999,
			arg3: true,
			arg4: &symbolPrice{},
			arg5: time.Date(2021, 5, 12, 11, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := &stockContractComponent{}
			got := component.confirmContractAuctionLO(test.arg1, test.arg2, test.arg3, test.arg4, test.arg5)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockContractComponent_confirmOrderContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 StockExecutionCondition
		arg2 Side
		arg3 float64
		arg4 bool
		arg5 *symbolPrice
		arg6 time.Time
		want *confirmContractResult
	}{
		{name: "引数がnilなら約定しない",
			arg5: nil,
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "成行が寄り価格で約定する",
			arg1: StockExecutionConditionMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "成行が引け価格で約定する",
			arg1: StockExecutionConditionMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "成行がザラバで約定する",
			arg1: StockExecutionConditionMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 1000, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "成行がタイミング不明なら約定しない",
			arg1: StockExecutionConditionMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 1000, kind: PriceKindUnspecified},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "寄成前場が寄り価格で約定する",
			arg1: StockExecutionConditionMOMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄成前場が2回目以降の確認では約定しない",
			arg1: StockExecutionConditionMOMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "寄成前場が後場では約定しない",
			arg1: StockExecutionConditionMOMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "寄成後場が寄り価格で約定する",
			arg1: StockExecutionConditionMOAO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local)}},
		{name: "寄成後場が2回目以降の確認では約定しない",
			arg1: StockExecutionConditionMOAO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "寄成後場が前場では約定しない",
			arg1: StockExecutionConditionMOAO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "引成前場が引け価格で約定する",
			arg1: StockExecutionConditionMOMC,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 11, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 11, 30, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 11, 30, 3, 0, time.Local)}},
		{name: "引成前場が2回目以降の確認では約定しない",
			arg1: StockExecutionConditionMOMC,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "引成前場が後場では約定しない",
			arg1: StockExecutionConditionMOMC,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "引成後場が引け価格で約定する",
			arg1: StockExecutionConditionMOAC,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "引成後場が2回目以降の確認では約定しない",
			arg1: StockExecutionConditionMOAC,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "引成後場が前場では約定しない",
			arg1: StockExecutionConditionMOAC,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: false},
		},
		{name: "IOC成行が寄り価格で約定する",
			arg1: StockExecutionConditionIOCMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC成行が引け価格で約定する",
			arg1: StockExecutionConditionIOCMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "IOC成行がザラバで約定する",
			arg1: StockExecutionConditionIOCMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 1000, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC成行がタイミング不明なら約定しない",
			arg1: StockExecutionConditionIOCMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 1000, kind: PriceKindUnspecified},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "IOC成行が1度でも約定確認されていたら約定しない",
			arg1: StockExecutionConditionIOCMO,
			arg2: SideBuy,
			arg3: 0.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 1000, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "指値が寄り価格で約定する",
			arg1: StockExecutionConditionLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "指値が引け価格で約定する",
			arg1: StockExecutionConditionLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "指値がザラバで約定する",
			arg1: StockExecutionConditionLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 990, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "指値がタイミング不明なら約定しない",
			arg1: StockExecutionConditionLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 1000, kind: PriceKindUnspecified},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "寄指前場が寄り価格で約定する",
			arg1: StockExecutionConditionLOMO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "寄指前場が2回目以降の確認では約定しない",
			arg1: StockExecutionConditionLOMO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "寄指前場が後場では約定しない",
			arg1: StockExecutionConditionLOMO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 13, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 13, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "寄指後場が寄り価格で約定する",
			arg1: StockExecutionConditionLOAO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local)}},
		{name: "寄指値後場が2回目以降の確認では約定しない",
			arg1: StockExecutionConditionLOMO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "寄指後場が前場では約定しない",
			arg1: StockExecutionConditionLOAO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "引指前場が引け価格で約定する",
			arg1: StockExecutionConditionLOMC,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 11, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 11, 30, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 11, 30, 3, 0, time.Local)}},
		{name: "引指前場が2回目以降の確認では約定しない",
			arg1: StockExecutionConditionLOMC,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "引指前場が後場では約定しない",
			arg1: StockExecutionConditionLOMC,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 12, 30, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "引指後場が引け価格で約定する",
			arg1: StockExecutionConditionLOAC,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "引指後場が2回目以降の確認では約定しない",
			arg1: StockExecutionConditionLOAC,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "引指後場が前場では約定しない",
			arg1: StockExecutionConditionLOAC,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "不成前場は前場の寄りではオークションの指値で約定する",
			arg1: StockExecutionConditionFunariM,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "不成前場は前場のザラバでは指値で約定する",
			arg1: StockExecutionConditionFunariM,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 990, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 990, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "不成前場は前場の引けではオークションの成行で約定する",
			arg1: StockExecutionConditionFunariM,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1200, PriceTime: time.Date(2021, 5, 12, 11, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 11, 30, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1200, contractedAt: time.Date(2021, 5, 12, 11, 30, 0, 0, time.Local)}},
		{name: "不成前場は後場の寄りではオークションの指値で約定する",
			arg1: StockExecutionConditionFunariM,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local)}},
		{name: "不成前場は後場のザラバでは指値で約定する",
			arg1: StockExecutionConditionFunariM,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 990, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 990, contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local)}},
		{name: "不成前場は後場の引けではオークションの指値で約定する",
			arg1: StockExecutionConditionFunariM,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 990, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 990, contractedAt: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local)}},
		{name: "不成後場は前場の寄りではオークションの指値で約定する",
			arg1: StockExecutionConditionFunariA,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "不成後場は前場のザラバでは指値で約定する",
			arg1: StockExecutionConditionFunariA,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 990, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 990, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "不成後場は前場の引けではオークションの指値で約定する",
			arg1: StockExecutionConditionFunariA,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 990, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 990, contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local)}},
		{name: "不成後場は後場の寄りではオークションの指値で約定する",
			arg1: StockExecutionConditionFunariA,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local)}},
		{name: "不成後場は後場のザラバでは指値で約定する",
			arg1: StockExecutionConditionFunariA,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 990, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 990, contractedAt: time.Date(2021, 5, 12, 12, 30, 0, 0, time.Local)}},
		{name: "不成後場は後場の引けではオークションの成行で約定する",
			arg1: StockExecutionConditionFunariA,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1200, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1200, contractedAt: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local)}},
		{name: "執行条件が逆指値の場合は約定しない",
			arg1: StockExecutionConditionStop,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1200, PriceTime: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "IOC指値が寄り価格で約定する",
			arg1: StockExecutionConditionIOCLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local), kind: PriceKindOpening},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC指値が引け価格で約定する",
			arg1: StockExecutionConditionIOCLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 12, 15, 0, 2, 0, time.Local), kind: PriceKindClosing},
			arg6: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 5, 12, 15, 0, 3, 0, time.Local)}},
		{name: "IOC指値がザラバで約定する",
			arg1: StockExecutionConditionIOCLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 990, kind: PriceKindRegular},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 990, contractedAt: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local)}},
		{name: "IOC指値がタイミング不明なら約定しない",
			arg1: StockExecutionConditionIOCLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: false,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 1000, kind: PriceKindUnspecified},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "IOC指値が1度でも約定確認されていたら約定しない",
			arg1: StockExecutionConditionIOCLO,
			arg2: SideBuy,
			arg3: 1000.0,
			arg4: true,
			arg5: &symbolPrice{SymbolCode: "1234", Ask: 1000, kind: PriceKindUnspecified},
			arg6: time.Date(2021, 5, 12, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := &stockContractComponent{}
			got := component.confirmOrderContract(test.arg1, test.arg2, test.arg3, test.arg4, test.arg5, test.arg6)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockContractComponent_confirmStockOrderContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 *stockOrder
		arg2 *symbolPrice
		arg3 time.Time
		want *confirmContractResult
	}{
		{name: "注文がnilなら約定しない",
			arg1: nil,
			want: &confirmContractResult{isContracted: false}},
		{name: "価格がnilなら約定しない",
			arg1: &stockOrder{},
			arg2: nil,
			want: &confirmContractResult{isContracted: false}},
		{name: "注文と価格の銘柄が一致しないなら約定しない",
			arg1: &stockOrder{SymbolCode: "1234"},
			arg2: &symbolPrice{SymbolCode: "0000"},
			want: &confirmContractResult{isContracted: false}},
		{name: "注文が約定可能な状態でないなら約定しない",
			arg1: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusDone},
			arg2: &symbolPrice{SymbolCode: "1234"},
			want: &confirmContractResult{isContracted: false}},
		{name: "confirmOrderContractが呼び出される(未約定)",
			arg1: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder},
			arg2: &symbolPrice{SymbolCode: "1234"},
			arg3: time.Date(2021, 8, 13, 0, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "confirmOrderContractが呼び出される(約定)",
			arg1: &stockOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionMO, OrderQuantity: 1},
			arg2: &symbolPrice{SymbolCode: "1234", Ask: 1000, AskTime: time.Date(2021, 8, 13, 9, 0, 0, 0, time.Local), kind: PriceKindRegular},
			arg3: time.Date(2021, 8, 13, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 13, 9, 0, 0, 0, time.Local)}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := &stockContractComponent{}
			got := component.confirmStockOrderContract(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockContractComponent_confirmMarginOrderContract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 *marginOrder
		arg2 *symbolPrice
		arg3 time.Time
		want *confirmContractResult
	}{
		{name: "注文がnilなら約定しない",
			arg1: nil,
			want: &confirmContractResult{isContracted: false}},
		{name: "価格がnilなら約定しない",
			arg1: &marginOrder{},
			arg2: nil,
			want: &confirmContractResult{isContracted: false}},
		{name: "注文と価格の銘柄が一致しないなら約定しない",
			arg1: &marginOrder{SymbolCode: "1234"},
			arg2: &symbolPrice{SymbolCode: "0000"},
			want: &confirmContractResult{isContracted: false}},
		{name: "注文が約定可能な状態でないなら約定しない",
			arg1: &marginOrder{SymbolCode: "1234", OrderStatus: OrderStatusDone},
			arg2: &symbolPrice{SymbolCode: "1234"},
			want: &confirmContractResult{isContracted: false}},
		{name: "confirmOrderContractが呼び出される(未約定)",
			arg1: &marginOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder},
			arg2: &symbolPrice{SymbolCode: "1234"},
			arg3: time.Date(2021, 8, 13, 0, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: false}},
		{name: "confirmOrderContractが呼び出される(約定)",
			arg1: &marginOrder{SymbolCode: "1234", OrderStatus: OrderStatusInOrder, Side: SideBuy, ExecutionCondition: StockExecutionConditionMO, OrderQuantity: 1},
			arg2: &symbolPrice{SymbolCode: "1234", Ask: 1000, AskTime: time.Date(2021, 8, 13, 9, 0, 0, 0, time.Local), kind: PriceKindRegular},
			arg3: time.Date(2021, 8, 13, 9, 0, 0, 0, time.Local),
			want: &confirmContractResult{isContracted: true, price: 1000, contractedAt: time.Date(2021, 8, 13, 9, 0, 0, 0, time.Local)}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := &stockContractComponent{}
			got := component.confirmMarginOrderContract(test.arg1, test.arg2, test.arg3)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
