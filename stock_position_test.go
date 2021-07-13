package virtual_security

import (
	"errors"
	"reflect"
	"testing"
)

func Test_stockPosition_exit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		position          *stockPosition
		arg               float64
		wantOwnedQuantity float64
		wantHoldQuantity  float64
		want              error
	}{
		{name: "エグジットできるなら保有数と拘束数を減らす",
			position:          &stockPosition{OwnedQuantity: 300, HoldQuantity: 200},
			arg:               100,
			wantOwnedQuantity: 200,
			wantHoldQuantity:  100,
			want:              nil},
		{name: "保有数不足でエグジットできないなら、エラーを返す",
			position:          &stockPosition{OwnedQuantity: 300, HoldQuantity: 300},
			arg:               500,
			wantOwnedQuantity: 300,
			wantHoldQuantity:  300,
			want:              NotEnoughOwnedQuantityError},
		{name: "拘束数不足でエグジットできないなら、エラーを返す",
			position:          &stockPosition{OwnedQuantity: 300, HoldQuantity: 200},
			arg:               300,
			wantOwnedQuantity: 300,
			wantHoldQuantity:  200,
			want:              NotEnoughHoldQuantityError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.position.exit(test.arg)
			if !reflect.DeepEqual(test.wantOwnedQuantity, test.position.OwnedQuantity) ||
				!reflect.DeepEqual(test.wantHoldQuantity, test.position.HoldQuantity) ||
				!errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(),
					test.wantOwnedQuantity, test.wantHoldQuantity, test.want,
					test.position.OwnedQuantity, test.position.HoldQuantity, got)
			}
		})
	}
}

func Test_stockPosition_hold(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		position         *stockPosition
		arg              float64
		wantHoldQuantity float64
		want             error
	}{
		{name: "拘束できるなら拘束数を増やす",
			position:         &stockPosition{OwnedQuantity: 300, HoldQuantity: 200},
			arg:              100,
			wantHoldQuantity: 300,
			want:             nil},
		{name: "拘束できないなら拘束数を増やさず、エラーを返す",
			position:         &stockPosition{OwnedQuantity: 300, HoldQuantity: 100},
			arg:              300,
			wantHoldQuantity: 100,
			want:             NotEnoughOwnedQuantityError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.position.hold(test.arg)
			if !reflect.DeepEqual(test.wantHoldQuantity, test.position.HoldQuantity) || !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.wantHoldQuantity, test.want, test.position.HoldQuantity, got)
			}
		})
	}
}

func Test_stockPosition_release(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		position         *stockPosition
		arg              float64
		wantHoldQuantity float64
		want             error
	}{
		{name: "拘束を解放できるなら拘束数を減らす",
			position:         &stockPosition{HoldQuantity: 300},
			arg:              100,
			wantHoldQuantity: 200,
			want:             nil},
		{name: "拘束を解放できないなら拘束数を減らさず、エラーを返す",
			position:         &stockPosition{HoldQuantity: 100},
			arg:              200,
			wantHoldQuantity: 100,
			want:             NotEnoughHoldQuantityError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.position.release(test.arg)
			if !reflect.DeepEqual(test.wantHoldQuantity, test.position.HoldQuantity) || !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.wantHoldQuantity, test.want, test.position.HoldQuantity, got)
			}
		})
	}
}

func Test_stockPosition_isDied(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name          string
		stockPosition *stockPosition
		want          bool
	}{
		{name: "保有数がなければ死んでいる", stockPosition: &stockPosition{OwnedQuantity: 0}, want: true},
		{name: "保有数があれば生きている", stockPosition: &stockPosition{OwnedQuantity: 100}, want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.stockPosition.isDied()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockPosition_orderableQuantity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		position *stockPosition
		want     float64
	}{
		{name: "保有数と拘束数が同じなら0", position: &stockPosition{OwnedQuantity: 300, HoldQuantity: 300}, want: 0},
		{name: "拘束されていなければ保有数がそのまま出る", position: &stockPosition{OwnedQuantity: 300, HoldQuantity: 0}, want: 300},
		{name: "部分的に拘束されているなら、保有数-拘束数の値が出る", position: &stockPosition{OwnedQuantity: 300, HoldQuantity: 200}, want: 100},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.position.orderableQuantity()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
