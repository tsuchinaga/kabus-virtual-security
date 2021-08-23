package virtual_security

import (
	"errors"
	"reflect"
	"testing"
)

func Test_marginPosition_exitable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		position *marginPosition
		arg      float64
		want     error
	}{
		{name: "保有数不足でエグジットできないなら、エラーを返す",
			position: &marginPosition{OwnedQuantity: 300, HoldQuantity: 300},
			arg:      500,
			want:     NotEnoughOwnedQuantityError},
		{name: "拘束数不足でエグジットできないなら、エラーを返す",
			position: &marginPosition{OwnedQuantity: 300, HoldQuantity: 200},
			arg:      300,
			want:     NotEnoughHoldQuantityError},
		{name: "exit可能ならnilを返す",
			position: &marginPosition{OwnedQuantity: 300, HoldQuantity: 300},
			arg:      300,
			want:     nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.position.exitable(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginPosition_exit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name              string
		position          *marginPosition
		arg               float64
		wantOwnedQuantity float64
		wantHoldQuantity  float64
		want              error
	}{
		{name: "エグジットできるなら保有数と拘束数を減らす",
			position:          &marginPosition{OwnedQuantity: 300, HoldQuantity: 200},
			arg:               100,
			wantOwnedQuantity: 200,
			wantHoldQuantity:  100,
			want:              nil},
		{name: "保有数不足でエグジットできないなら、エラーを返す",
			position:          &marginPosition{OwnedQuantity: 300, HoldQuantity: 300},
			arg:               500,
			wantOwnedQuantity: 300,
			wantHoldQuantity:  300,
			want:              NotEnoughOwnedQuantityError},
		{name: "拘束数不足でエグジットできないなら、エラーを返す",
			position:          &marginPosition{OwnedQuantity: 300, HoldQuantity: 200},
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

func Test_marginPosition_holdable(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		position *marginPosition
		arg      float64
		want     error
	}{
		{name: "拘束できないなら、エラーを返す",
			position: &marginPosition{OwnedQuantity: 300, HoldQuantity: 100},
			arg:      300,
			want:     NotEnoughOwnedQuantityError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.position.holdable(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginPosition_hold(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		position         *marginPosition
		arg              float64
		wantHoldQuantity float64
		want             error
	}{
		{name: "拘束できるなら拘束数を増やす",
			position:         &marginPosition{OwnedQuantity: 300, HoldQuantity: 200},
			arg:              100,
			wantHoldQuantity: 300,
			want:             nil},
		{name: "拘束できないなら拘束数を増やさず、エラーを返す",
			position:         &marginPosition{OwnedQuantity: 300, HoldQuantity: 100},
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

func Test_marginPosition_release(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		position         *marginPosition
		arg              float64
		wantHoldQuantity float64
		want             error
	}{
		{name: "拘束を解放できるなら拘束数を減らす",
			position:         &marginPosition{HoldQuantity: 300},
			arg:              100,
			wantHoldQuantity: 200,
			want:             nil},
		{name: "拘束を解放できないなら拘束数を減らさず、エラーを返す",
			position:         &marginPosition{HoldQuantity: 100},
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

func Test_marginPosition_isDied(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		marginPosition *marginPosition
		want           bool
	}{
		{name: "保有数がなければ死んでいる", marginPosition: &marginPosition{OwnedQuantity: 0}, want: true},
		{name: "保有数があれば生きている", marginPosition: &marginPosition{OwnedQuantity: 100}, want: false},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.marginPosition.isDied()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginPosition_orderableQuantity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		position *marginPosition
		want     float64
	}{
		{name: "保有数と拘束数が同じなら0", position: &marginPosition{OwnedQuantity: 300, HoldQuantity: 300}, want: 0},
		{name: "拘束されていなければ保有数がそのまま出る", position: &marginPosition{OwnedQuantity: 300, HoldQuantity: 0}, want: 300},
		{name: "部分的に拘束されているなら、保有数-拘束数の値が出る", position: &marginPosition{OwnedQuantity: 300, HoldQuantity: 200}, want: 100},
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
