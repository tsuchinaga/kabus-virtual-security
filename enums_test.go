package virtual_security

import (
	"reflect"
	"testing"
)

func Test_StockExecutionCondition_IsContractableMorningSession(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		stockExecutionCondition StockExecutionCondition
		want                    bool
	}{
		{name: "未指定 は前場で約定不可能", stockExecutionCondition: StockExecutionConditionUnspecified, want: false},
		{name: "成行 は前場で約定可能", stockExecutionCondition: StockExecutionConditionMO, want: true},
		{name: "寄成(前場) は前場で約定可能", stockExecutionCondition: StockExecutionConditionMOMO, want: true},
		{name: "寄成(後場) は前場で約定不可能", stockExecutionCondition: StockExecutionConditionMOAO, want: false},
		{name: "引成(前場) は前場で約定不可能", stockExecutionCondition: StockExecutionConditionMOMC, want: false},
		{name: "引成(後場) は前場で約定不可能", stockExecutionCondition: StockExecutionConditionMOAC, want: false},
		{name: "IOC成行 は前場で約定可能", stockExecutionCondition: StockExecutionConditionIOCMO, want: true},
		{name: "指値 は前場で約定可能", stockExecutionCondition: StockExecutionConditionLO, want: true},
		{name: "寄指(前場) は前場で約定可能", stockExecutionCondition: StockExecutionConditionLOMO, want: true},
		{name: "寄指(後場) は前場で約定不可能", stockExecutionCondition: StockExecutionConditionLOAO, want: false},
		{name: "引指(前場) は前場で約定不可能", stockExecutionCondition: StockExecutionConditionLOMC, want: false},
		{name: "引指(後場) は前場で約定不可能", stockExecutionCondition: StockExecutionConditionLOAC, want: false},
		{name: "IOC指値 は前場で約定可能", stockExecutionCondition: StockExecutionConditionIOCLO, want: true},
		{name: "不成(前場) は前場で約定可能", stockExecutionCondition: StockExecutionConditionFUNARIM, want: true},
		{name: "不成(後場) は前場で約定可能", stockExecutionCondition: StockExecutionConditionFUNARIA, want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockExecutionCondition.IsContractableMorningSession()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockExecutionCondition_IsContractableMorningSessionClosing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		stockExecutionCondition StockExecutionCondition
		want                    bool
	}{
		{name: "未指定 は前場終了で約定不可能", stockExecutionCondition: StockExecutionConditionUnspecified, want: false},
		{name: "成行 は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionMO, want: true},
		{name: "寄成(前場) は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionMOMO, want: true},
		{name: "寄成(後場) は前場終了で約定不可能", stockExecutionCondition: StockExecutionConditionMOAO, want: false},
		{name: "引成(前場) は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionMOMC, want: true},
		{name: "引成(後場) は前場終了で約定不可能", stockExecutionCondition: StockExecutionConditionMOAC, want: false},
		{name: "IOC成行 は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionIOCMO, want: true},
		{name: "指値 は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionLO, want: true},
		{name: "寄指(前場) は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionLOMO, want: true},
		{name: "寄指(後場) は前場終了で約定不可能", stockExecutionCondition: StockExecutionConditionLOAO, want: false},
		{name: "引指(前場) は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionLOMC, want: true},
		{name: "引指(後場) は前場終了で約定不可能", stockExecutionCondition: StockExecutionConditionLOAC, want: false},
		{name: "IOC指値 は前場終了で約定不可能", stockExecutionCondition: StockExecutionConditionIOCLO, want: true},
		{name: "不成(前場) は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionFUNARIM, want: true},
		{name: "不成(後場) は前場終了で約定可能", stockExecutionCondition: StockExecutionConditionFUNARIA, want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockExecutionCondition.IsContractableMorningSessionClosing()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockExecutionCondition_IsContractableAfternoonSession(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		stockExecutionCondition StockExecutionCondition
		want                    bool
	}{
		{name: "未指定 は後場で約定不可能", stockExecutionCondition: StockExecutionConditionUnspecified, want: false},
		{name: "成行 は後場で約定可能", stockExecutionCondition: StockExecutionConditionMO, want: true},
		{name: "寄成(前場) は後場で約定不可能", stockExecutionCondition: StockExecutionConditionMOMO, want: false},
		{name: "寄成(後場) は後場で約定可能", stockExecutionCondition: StockExecutionConditionMOAO, want: true},
		{name: "引成(前場) は後場で約定不可能", stockExecutionCondition: StockExecutionConditionMOMC, want: false},
		{name: "引成(後場) は後場で約定不可能", stockExecutionCondition: StockExecutionConditionMOAC, want: false},
		{name: "IOC成行 は後場で約定可能", stockExecutionCondition: StockExecutionConditionIOCMO, want: true},
		{name: "指値 は後場で約定可能", stockExecutionCondition: StockExecutionConditionLO, want: true},
		{name: "寄指(前場) は後場で約定不可能", stockExecutionCondition: StockExecutionConditionLOMO, want: false},
		{name: "寄指(後場) は後場で約定可能", stockExecutionCondition: StockExecutionConditionLOAO, want: true},
		{name: "引指(前場) は後場で約定不可能", stockExecutionCondition: StockExecutionConditionLOMC, want: false},
		{name: "引指(後場) は後場で約定不可能", stockExecutionCondition: StockExecutionConditionLOAC, want: false},
		{name: "IOC指値 は後場で約定可能", stockExecutionCondition: StockExecutionConditionIOCLO, want: true},
		{name: "不成(前場) は後場で約定可能", stockExecutionCondition: StockExecutionConditionFUNARIM, want: true},
		{name: "不成(後場) は後場で約定可能", stockExecutionCondition: StockExecutionConditionFUNARIA, want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockExecutionCondition.IsContractableAfternoonSession()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockExecutionCondition_IsContractableAfternoonSessionClosing(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		stockExecutionCondition StockExecutionCondition
		want                    bool
	}{
		{name: "未指定 は後場終了で約定不可能", stockExecutionCondition: StockExecutionConditionUnspecified, want: false},
		{name: "成行 は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionMO, want: true},
		{name: "寄成(前場) は後場終了で約定不可能", stockExecutionCondition: StockExecutionConditionMOMO, want: false},
		{name: "寄成(後場) は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionMOAO, want: true},
		{name: "引成(前場) は後場終了で約定不可能", stockExecutionCondition: StockExecutionConditionMOMC, want: false},
		{name: "引成(後場) は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionMOAC, want: true},
		{name: "IOC成行 は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionIOCMO, want: true},
		{name: "指値 は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionLO, want: true},
		{name: "寄指(前場) は後場終了で約定不可能", stockExecutionCondition: StockExecutionConditionLOMO, want: false},
		{name: "寄指(後場) は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionLOAO, want: true},
		{name: "引指(前場) は後場終了で約定不可能", stockExecutionCondition: StockExecutionConditionLOMC, want: false},
		{name: "引指(後場) は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionLOAC, want: true},
		{name: "IOC指値 は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionIOCLO, want: true},
		{name: "不成(前場) は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionFUNARIM, want: true},
		{name: "不成(後場) は後場終了で約定可能", stockExecutionCondition: StockExecutionConditionFUNARIA, want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockExecutionCondition.IsContractableAfternoonSessionClosing()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_OrderStatus_IsContractable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		orderStatus OrderStatus
		want        bool
	}{
		{name: "未指定 は約定できない", orderStatus: OrderStatusUnspecified, want: false},
		{name: "新規 は約定できない", orderStatus: OrderStatusNew, want: false},
		{name: "注文中 は約定できる", orderStatus: OrderStatusInOrder, want: true},
		{name: "部分約定 は約定できる", orderStatus: OrderStatusPart, want: true},
		{name: "全約定 は約定できない", orderStatus: OrderStatusDone, want: false},
		{name: "取消中 は約定できない", orderStatus: OrderStatusInCancel, want: false},
		{name: "取消済み は約定できない", orderStatus: OrderStatusCanceled, want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.orderStatus.IsContractable()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_OrderStatus_IsFixed(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		orderStatus OrderStatus
		want        bool
	}{
		{name: "未指定 は固定されている", orderStatus: OrderStatusUnspecified, want: true},
		{name: "新規 は固定されていない", orderStatus: OrderStatusNew, want: false},
		{name: "注文中 は固定されていない", orderStatus: OrderStatusInOrder, want: false},
		{name: "部分約定 は固定されていない", orderStatus: OrderStatusPart, want: false},
		{name: "全約定 は固定されている", orderStatus: OrderStatusDone, want: true},
		{name: "取消中 は固定されていない", orderStatus: OrderStatusInCancel, want: false},
		{name: "取消済み は固定されている", orderStatus: OrderStatusCanceled, want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.orderStatus.IsFixed()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockExecutionCondition_IsMarketOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		stockExecutionCondition StockExecutionCondition
		want                    bool
	}{
		{name: "未指定 は成行注文ではない", stockExecutionCondition: StockExecutionConditionUnspecified, want: false},
		{name: "成行 は成行注文", stockExecutionCondition: StockExecutionConditionMO, want: true},
		{name: "寄成(前場) は成行注文", stockExecutionCondition: StockExecutionConditionMOMO, want: true},
		{name: "寄成(後場) は成行注文", stockExecutionCondition: StockExecutionConditionMOMC, want: true},
		{name: "引成(前場) は成行注文", stockExecutionCondition: StockExecutionConditionMOAO, want: true},
		{name: "引成(後場) は成行注文", stockExecutionCondition: StockExecutionConditionMOAC, want: true},
		{name: "IOC成行 は成行注文", stockExecutionCondition: StockExecutionConditionIOCMO, want: true},
		{name: "指値 は成行注文ではない", stockExecutionCondition: StockExecutionConditionLO, want: false},
		{name: "寄指(前場) は成行注文ではない", stockExecutionCondition: StockExecutionConditionLOMO, want: false},
		{name: "寄指(後場) は成行注文ではない", stockExecutionCondition: StockExecutionConditionLOAO, want: false},
		{name: "引指(前場) は成行注文ではない", stockExecutionCondition: StockExecutionConditionLOMC, want: false},
		{name: "引指(後場) は成行注文ではない", stockExecutionCondition: StockExecutionConditionLOAC, want: false},
		{name: "IOC指値 は成行注文ではない", stockExecutionCondition: StockExecutionConditionIOCLO, want: false},
		{name: "不成(前場) は成行注文ではない", stockExecutionCondition: StockExecutionConditionFUNARIM, want: false},
		{name: "不成(後場) は成行注文ではない", stockExecutionCondition: StockExecutionConditionFUNARIA, want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockExecutionCondition.IsMarketOrder()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockExecutionCondition_IsLimitOrder(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		stockExecutionCondition StockExecutionCondition
		want                    bool
	}{
		{name: "未指定 は指値注文ではない", stockExecutionCondition: StockExecutionConditionUnspecified, want: false},
		{name: "成行 は指値注文ではない", stockExecutionCondition: StockExecutionConditionMO, want: false},
		{name: "寄成(前場) は指値注文ではない", stockExecutionCondition: StockExecutionConditionMOMO, want: false},
		{name: "寄成(後場) は指値注文ではない", stockExecutionCondition: StockExecutionConditionMOMC, want: false},
		{name: "引成(前場) は指値注文ではない", stockExecutionCondition: StockExecutionConditionMOAO, want: false},
		{name: "引成(後場) は指値注文ではない", stockExecutionCondition: StockExecutionConditionMOAC, want: false},
		{name: "IOC成行 は指値注文ではない", stockExecutionCondition: StockExecutionConditionIOCMO, want: false},
		{name: "指値 は指値注文", stockExecutionCondition: StockExecutionConditionLO, want: true},
		{name: "寄指(前場) は指値注文", stockExecutionCondition: StockExecutionConditionLOMO, want: true},
		{name: "寄指(後場) は指値注文", stockExecutionCondition: StockExecutionConditionLOAO, want: true},
		{name: "引指(前場) は指値注文", stockExecutionCondition: StockExecutionConditionLOMC, want: true},
		{name: "引指(後場) は指値注文", stockExecutionCondition: StockExecutionConditionLOAC, want: true},
		{name: "IOC指値 は指値注文", stockExecutionCondition: StockExecutionConditionIOCLO, want: true},
		{name: "不成(前場) は指値注文ではない", stockExecutionCondition: StockExecutionConditionFUNARIM, want: false},
		{name: "不成(後場) は指値注文ではない", stockExecutionCondition: StockExecutionConditionFUNARIA, want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockExecutionCondition.IsLimitOrder()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_StockExecutionCondition_IsFunari(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                    string
		stockExecutionCondition StockExecutionCondition
		want                    bool
	}{
		{name: "未指定 は不成注文ではない", stockExecutionCondition: StockExecutionConditionUnspecified, want: false},
		{name: "成行 は不成注文ではない", stockExecutionCondition: StockExecutionConditionMO, want: false},
		{name: "寄成(前場) は不成注文ではない", stockExecutionCondition: StockExecutionConditionMOMO, want: false},
		{name: "寄成(後場) は不成注文ではない", stockExecutionCondition: StockExecutionConditionMOMC, want: false},
		{name: "引成(前場) は不成注文ではない", stockExecutionCondition: StockExecutionConditionMOAO, want: false},
		{name: "引成(後場) は不成注文ではない", stockExecutionCondition: StockExecutionConditionMOAC, want: false},
		{name: "IOC成行 は不成注文ではない", stockExecutionCondition: StockExecutionConditionIOCMO, want: false},
		{name: "指値 は不成注文ではない", stockExecutionCondition: StockExecutionConditionLO, want: false},
		{name: "寄指(前場) は不成注文ではない", stockExecutionCondition: StockExecutionConditionLOMO, want: false},
		{name: "寄指(後場) は不成注文ではない", stockExecutionCondition: StockExecutionConditionLOAO, want: false},
		{name: "引指(前場) は不成注文ではない", stockExecutionCondition: StockExecutionConditionLOMC, want: false},
		{name: "引指(後場) は不成注文ではない", stockExecutionCondition: StockExecutionConditionLOAC, want: false},
		{name: "IOC指値 は不成注文ではない", stockExecutionCondition: StockExecutionConditionIOCLO, want: false},
		{name: "不成(前場) は不成注文", stockExecutionCondition: StockExecutionConditionFUNARIM, want: true},
		{name: "不成(後場) は不成注文", stockExecutionCondition: StockExecutionConditionFUNARIA, want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockExecutionCondition.IsFunari()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
