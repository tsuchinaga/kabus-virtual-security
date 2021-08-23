package virtual_security

import (
	"errors"
	"testing"
	"time"
)

type testValidationComponent struct {
	iValidatorComponent
	isValidMarginOrder1 error
}

func (t *testValidationComponent) isValidMarginOrder(*marginOrder, time.Time, []*marginPosition) error {
	return t.isValidMarginOrder1
}

func Test_validationComponent_isValidMarginOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 *marginOrder
		arg2 time.Time
		arg3 []*marginPosition
		want error
	}{
		{name: "取引種別が不明ならエラー",
			arg1: &marginOrder{
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidTradeTypeError},
		{name: "方向が不明ならエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidSideError},
		{name: "執行条件が不明ならエラー",
			arg1: &marginOrder{
				TradeType:     TradeTypeEntry,
				Side:          SideBuy,
				SymbolCode:    "1234",
				OrderQuantity: 100,
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidExecutionConditionError},
		{name: "銘柄がゼロ値ならエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				OrderQuantity:      100,
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidSymbolCodeError},
		{name: "数量がゼロ値ならエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidQuantityError},
		{name: "指値を指定して指値価格がゼロ値ならエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionLO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidLimitPriceError},
		{name: "指値を指定して指値価格があればエラーなし",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionLO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				LimitPrice:         1000,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: nil},
		{name: "有効期限が過去ならエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 18, 0, 0, 0, 0, time.Local),
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidExpiredError},
		{name: "逆指値で逆指値条件が設定されていなければエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidStopConditionError},
		{name: "逆指値でトリガー価格が設定されていなければエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				StopCondition:      &StockStopCondition{},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidStopConditionError},
		{name: "逆指値でトリガー後の執行条件が逆指値ならエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				StopCondition:      &StockStopCondition{StopPrice: 2000, ExecutionConditionAfterHit: StockExecutionConditionStop},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidStopConditionError},
		{name: "逆指値でトリガー後の執行条件が指値で指値価格が指定されていなければエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				StopCondition:      &StockStopCondition{StopPrice: 2000, ExecutionConditionAfterHit: StockExecutionConditionLO},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidStopConditionError},
		{name: "逆指値でトリガー後の執行条件が指値で指値価格が指定されていればエラーなし",
			arg1: &marginOrder{
				TradeType:          TradeTypeEntry,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionStop,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				StopCondition:      &StockStopCondition{StopPrice: 2000, ExecutionConditionAfterHit: StockExecutionConditionLO, LimitPriceAfterHit: 2100},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: nil},
		{name: "ExitでExitするポジションがnilならエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeExit,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidExitPositionError},
		{name: "ExitでExitするポジションが空配列ならエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeExit,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				ExitPositionList:   []ExitPosition{},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidExitPositionError},
		{name: "ExitでExitするポジションの数量と全数量が一致していなければエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeExit,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				ExitPositionList:   []ExitPosition{{PositionCode: "mpo-01", Quantity: 50}},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidExitQuantityError},
		{name: "ExitでExitするポジションが存在しなければエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeExit,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				ExitPositionList:   []ExitPosition{{PositionCode: "mpo-01", Quantity: 50}, {PositionCode: "mpo-02", Quantity: 30}, {PositionCode: "mpo-03", Quantity: 20}},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{},
			want: InvalidExitPositionCodeError},
		{name: "ExitでExitするポジションの保有数が足りなければエラー",
			arg1: &marginOrder{
				TradeType:          TradeTypeExit,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				ExitPositionList:   []ExitPosition{{PositionCode: "mpo-01", Quantity: 50}, {PositionCode: "mpo-02", Quantity: 30}, {PositionCode: "mpo-03", Quantity: 20}},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 0}, {Code: "mpo-02", OwnedQuantity: 100, HoldQuantity: 70}, {Code: "mpo-03", OwnedQuantity: 50, HoldQuantity: 40}},
			want: NotEnoughOwnedQuantityError},
		{name: "ExitでExitするポジションがあればエラーなし",
			arg1: &marginOrder{
				TradeType:          TradeTypeExit,
				Side:               SideBuy,
				ExecutionCondition: StockExecutionConditionMO,
				SymbolCode:         "1234",
				OrderQuantity:      100,
				ExpiredAt:          time.Date(2021, 8, 19, 0, 0, 0, 0, time.Local),
				ExitPositionList:   []ExitPosition{{PositionCode: "mpo-01", Quantity: 50}, {PositionCode: "mpo-02", Quantity: 30}, {PositionCode: "mpo-03", Quantity: 20}},
			},
			arg2: time.Date(2021, 8, 19, 14, 0, 0, 0, time.Local),
			arg3: []*marginPosition{{Code: "mpo-01", OwnedQuantity: 100, HoldQuantity: 0}, {Code: "mpo-02", OwnedQuantity: 100, HoldQuantity: 70}, {Code: "mpo-03", OwnedQuantity: 50, HoldQuantity: 0}},
			want: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := &validatorComponent{}
			got := component.isValidMarginOrder(test.arg1, test.arg2, test.arg3)
			if !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
