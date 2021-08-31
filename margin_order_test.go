package virtual_security

import (
	"reflect"
	"testing"
	"time"
)

func Test_marginOrder_isDied(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		marginOrder *marginOrder
		arg         time.Time
		want        bool
	}{
		{name: "未終了の注文なら生きている",
			marginOrder: &marginOrder{OrderStatus: OrderStatusInOrder},
			arg:         time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:        false},
		{name: "取消済み注文で、取消から1日以内なら生きている",
			marginOrder: &marginOrder{OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 6, 14, 11, 0, 0, 0, time.Local)},
			arg:         time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:        false},
		{name: "取消済み注文で、取消から1日丁度なら生きている",
			marginOrder: &marginOrder{OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local)},
			arg:         time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:        false},
		{name: "取消済み注文で、取消から1日以上経っていたら死んでいる",
			marginOrder: &marginOrder{OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local)},
			arg:         time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:        true},
		{name: "約定済み注文で、最後の約定から1日以内なら生きている",
			marginOrder: &marginOrder{OrderStatus: OrderStatusDone, Contracts: []*Contract{
				{ContractedAt: time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local)},
				{ContractedAt: time.Date(2021, 6, 14, 11, 0, 0, 0, time.Local)}}},
			arg:  time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want: false},
		{name: "約定済み注文で、最後の約定から1日丁度なら生きている",
			marginOrder: &marginOrder{OrderStatus: OrderStatusDone, Contracts: []*Contract{
				{ContractedAt: time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local)},
				{ContractedAt: time.Date(2021, 6, 14, 10, 0, 0, 0, time.Local)}}},
			arg:  time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want: false},
		{name: "約定済み注文で、最後の約定から1日以上経っていたら死んでいる",
			marginOrder: &marginOrder{OrderStatus: OrderStatusDone, Contracts: []*Contract{
				{ContractedAt: time.Date(2021, 6, 14, 9, 0, 0, 0, time.Local)},
				{ContractedAt: time.Date(2021, 6, 14, 9, 30, 0, 0, time.Local)}}},
			arg:  time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want: true},
		{name: "終了した注文で、取消も約定も情報が無かったら死んだものとする",
			marginOrder: &marginOrder{OrderStatus: OrderStatusDone, Contracts: []*Contract{}},
			arg:         time.Date(2021, 6, 15, 10, 0, 0, 0, time.Local),
			want:        true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.marginOrder.isDied(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginOrder_executionCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		marginOrder *marginOrder
		want        StockExecutionCondition
	}{
		{name: "逆指値で待機中でなく逆指値条件があれば、逆指値発動後の条件が返される",
			marginOrder: &marginOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO}},
			want: StockExecutionConditionMO},
		{name: "逆指値で待機中でなくても、逆指値条件がなければそのまま返す",
			marginOrder: &marginOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      nil},
			want: StockExecutionConditionStop},
		{name: "逆指値でも待機中なら、そのまま返す",
			marginOrder: &marginOrder{
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO}},
			want: StockExecutionConditionStop},
		{name: "逆指値注文でなければそのまま返す",
			marginOrder: &marginOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionLO,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO}},
			want: StockExecutionConditionLO},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.marginOrder.executionCondition()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginOrder_limitPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		marginOrder *marginOrder
		want        float64
	}{
		{name: "逆指値で待機中でなく逆指値条件があれば、逆指値発動後の指値価格が返される",
			marginOrder: &marginOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionStop,
				LimitPrice:         1000,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO, LimitPriceAfterHit: 1500}},
			want: 1500},
		{name: "逆指値で待機中でなくても、逆指値条件がなければそのまま返す",
			marginOrder: &marginOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      nil},
			want: 0},
		{name: "逆指値でも待機中なら、そのまま返す",
			marginOrder: &marginOrder{
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO, LimitPriceAfterHit: 1500}},
			want: 0},
		{name: "逆指値注文でなければそのまま返す",
			marginOrder: &marginOrder{
				OrderStatus:        OrderStatusInOrder,
				ExecutionCondition: StockExecutionConditionLO,
				StopCondition:      &StockStopCondition{ExecutionConditionAfterHit: StockExecutionConditionMO, LimitPriceAfterHit: 1500}},
			want: 0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.marginOrder.limitPrice()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginOrder_activate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		marginOrder *marginOrder
		arg1        *symbolPrice
		arg2        time.Time
		wantStatus  OrderStatus
	}{
		{name: "条件を満たせば注文中になる",
			marginOrder: &marginOrder{
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
			marginOrder: &marginOrder{
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
			marginOrder: &marginOrder{
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
			marginOrder: &marginOrder{
				SymbolCode:         "1234",
				OrderStatus:        OrderStatusWait,
				ExecutionCondition: StockExecutionConditionStop},
			arg1:       &symbolPrice{SymbolCode: "1234", Price: 1000, PriceTime: time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local)},
			arg2:       time.Date(2021, 5, 30, 20, 32, 0, 0, time.Local),
			wantStatus: OrderStatusWait},
		{name: "逆指値注文でなければ何もしない",
			marginOrder: &marginOrder{
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
			marginOrder: &marginOrder{
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
			marginOrder: &marginOrder{
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
			marginOrder: &marginOrder{
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
			test.marginOrder.activate(test.arg1, test.arg2)
			got1 := test.marginOrder.OrderStatus
			if !reflect.DeepEqual(test.wantStatus, got1) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantStatus, got1)
			}
		})
	}
}

func Test_marginOrder_expired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		marginOrder     *marginOrder
		arg             time.Time
		wantOrderStatus OrderStatus
	}{
		{name: "有効期限がゼロ値なら何もしない",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusInOrder, ExpiredAt: time.Time{}},
			arg:             time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local),
			wantOrderStatus: OrderStatusInOrder},
		{name: "有効期限が現在時刻よりも過去なら取消済みにする",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusInOrder, ExpiredAt: time.Date(2021, 6, 7, 13, 0, 0, 0, time.Local)},
			arg:             time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled},
		{name: "有効期限が現在時刻と一致しているなら状態を変えない",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusInOrder, ExpiredAt: time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local)},
			arg:             time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local),
			wantOrderStatus: OrderStatusInOrder},
		{name: "有効期限が現在時刻よりも未来なら状態を変えない",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusInOrder, ExpiredAt: time.Date(2021, 6, 7, 15, 0, 0, 0, time.Local)},
			arg:             time.Date(2021, 6, 7, 13, 24, 0, 0, time.Local),
			wantOrderStatus: OrderStatusInOrder},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.marginOrder.expired(test.arg)
			if !reflect.DeepEqual(test.wantOrderStatus, test.marginOrder.OrderStatus) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantOrderStatus, test.marginOrder.OrderStatus)
			}
		})
	}
}

func Test_marginOrder_contract(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                   string
		marginOrder            *marginOrder
		arg                    *Contract
		wantContractedQuantity float64
		wantStatus             OrderStatus
	}{
		{name: "引数がnilなら何もしない",
			marginOrder:            &marginOrder{OrderQuantity: 3, ContractedQuantity: 0, OrderStatus: OrderStatusUnspecified},
			arg:                    nil,
			wantContractedQuantity: 0,
			wantStatus:             OrderStatusUnspecified},
		{name: "約定後、約定数量が0なら注文中",
			marginOrder:            &marginOrder{OrderQuantity: 3, ContractedQuantity: 0, OrderStatus: OrderStatusUnspecified},
			arg:                    &Contract{Quantity: 0},
			wantContractedQuantity: 0,
			wantStatus:             OrderStatusInOrder},
		{name: "約定後、約定数量が注文数量未満なら部分約定",
			marginOrder:            &marginOrder{OrderQuantity: 3, ContractedQuantity: 0, OrderStatus: OrderStatusUnspecified},
			arg:                    &Contract{Quantity: 1},
			wantContractedQuantity: 1,
			wantStatus:             OrderStatusPart},
		{name: "約定後、約定数量が注文数量以上なら全約定",
			marginOrder:            &marginOrder{OrderQuantity: 3, ContractedQuantity: 1, OrderStatus: OrderStatusUnspecified},
			arg:                    &Contract{Quantity: 2},
			wantContractedQuantity: 3,
			wantStatus:             OrderStatusDone},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.marginOrder.contract(test.arg)
			got1 := test.marginOrder.ContractedQuantity
			got2 := test.marginOrder.OrderStatus
			if !reflect.DeepEqual(test.wantContractedQuantity, got1) || !reflect.DeepEqual(test.wantStatus, got2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.wantContractedQuantity, test.wantStatus, got1, got2)
			}
		})
	}
}

func Test_marginOrder_cancel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		marginOrder     *marginOrder
		arg             time.Time
		wantOrderStatus OrderStatus
		wantCanceledAt  time.Time
	}{
		{name: "ステータスがnewなら取消状態に更新",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusNew},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled,
			wantCanceledAt:  time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local)},
		{name: "ステータスがin_orderなら取消状態に更新",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusInOrder},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled,
			wantCanceledAt:  time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local)},
		{name: "ステータスがpartなら取消状態に更新",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusPart},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled,
			wantCanceledAt:  time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local)},
		{name: "ステータスがdoneなら取消状態に更新できない",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusDone},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusDone,
			wantCanceledAt:  time.Time{}},
		{name: "ステータスがcanceledなら取消状態に更新できない",
			marginOrder:     &marginOrder{OrderStatus: OrderStatusCanceled, CanceledAt: time.Date(2021, 5, 18, 10, 0, 0, 0, time.Local)},
			arg:             time.Date(2021, 5, 18, 11, 0, 0, 0, time.Local),
			wantOrderStatus: OrderStatusCanceled,
			wantCanceledAt:  time.Date(2021, 5, 18, 10, 0, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.marginOrder.cancel(test.arg)
			got1 := test.marginOrder.OrderStatus
			got2 := test.marginOrder.CanceledAt
			if !reflect.DeepEqual(test.wantOrderStatus, got1) || !reflect.DeepEqual(test.wantCanceledAt, got2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.wantOrderStatus, test.wantCanceledAt, got1, got2)
			}
		})
	}
}

func Test_marginOrder_addHoldPosition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		order *marginOrder
		arg1  string
		arg2  float64
		want  *marginOrder
	}{
		{name: "sliceがnilなら空sliceを作ってからappendする",
			order: &marginOrder{},
			arg1:  "spo-uuid-01",
			arg2:  100,
			want:  &marginOrder{HoldPositions: []*HoldPosition{{PositionCode: "spo-uuid-01", HoldQuantity: 100}}}},
		{name: "sliceに要素があったら末尾にappendする",
			order: &marginOrder{HoldPositions: []*HoldPosition{{PositionCode: "spo-uuid-01", HoldQuantity: 100}}},
			arg1:  "spo-uuid-02",
			arg2:  1000,
			want:  &marginOrder{HoldPositions: []*HoldPosition{{PositionCode: "spo-uuid-01", HoldQuantity: 100}, {PositionCode: "spo-uuid-02", HoldQuantity: 1000}}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.order.addHoldPosition(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, test.order) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, test.order)
			}
		})
	}
}

func Test_marginOrder_addExitPosition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		order     *marginOrder
		arg1      string
		arg2      float64
		wantOrder *marginOrder
	}{
		{name: "注文でHoldしているポジションがnilなら何もしない",
			order:     &marginOrder{HoldPositions: nil},
			arg1:      "spo-uuid-01",
			arg2:      50,
			wantOrder: &marginOrder{HoldPositions: nil}},
		{name: "注文でHoldしているポジションと一致しないなら何もしない",
			order:     &marginOrder{HoldPositions: []*HoldPosition{{PositionCode: "spo-uuid-02", HoldQuantity: 100, ExitQuantity: 100}, {PositionCode: "spo-uuid-03", HoldQuantity: 300, ExitQuantity: 200}}},
			arg1:      "spo-uuid-01",
			arg2:      50,
			wantOrder: &marginOrder{HoldPositions: []*HoldPosition{{PositionCode: "spo-uuid-02", HoldQuantity: 100, ExitQuantity: 100}, {PositionCode: "spo-uuid-03", HoldQuantity: 300, ExitQuantity: 200}}}},
		{name: "注文でHoldしているポジションをExitした場合、Exit数に加算しておく",
			order:     &marginOrder{HoldPositions: []*HoldPosition{{PositionCode: "spo-uuid-02", HoldQuantity: 100, ExitQuantity: 100}, {PositionCode: "spo-uuid-03", HoldQuantity: 300, ExitQuantity: 200}}},
			arg1:      "spo-uuid-03",
			arg2:      50,
			wantOrder: &marginOrder{HoldPositions: []*HoldPosition{{PositionCode: "spo-uuid-02", HoldQuantity: 100, ExitQuantity: 100}, {PositionCode: "spo-uuid-03", HoldQuantity: 300, ExitQuantity: 250}}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.order.addExitPosition(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.wantOrder, test.order) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantOrder, test.order)
			}
		})
	}
}
