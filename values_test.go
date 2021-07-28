package virtual_security

import (
	"reflect"
	"testing"
	"time"
)

func Test_SymbolPrice_maxTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		symbolPrice symbolPrice
		want        time.Time
	}{
		{name: "現値時刻、最良売り時刻、最良買い時刻すべてがゼロ値ならゼロ値",
			symbolPrice: symbolPrice{},
			want:        time.Time{}},
		{name: "現値時刻だけがゼロ値じゃなければ現値時刻",
			symbolPrice: symbolPrice{PriceTime: time.Date(2021, 5, 20, 23, 53, 0, 0, time.Local)},
			want:        time.Date(2021, 5, 20, 23, 53, 0, 0, time.Local)},
		{name: "最良売り時刻だけがゼロ値じゃなければ最良売り時刻",
			symbolPrice: symbolPrice{AskTime: time.Date(2021, 5, 20, 23, 54, 0, 0, time.Local)},
			want:        time.Date(2021, 5, 20, 23, 54, 0, 0, time.Local)},
		{name: "最良買い時刻だけがゼロ値じゃなければ最良買い時刻",
			symbolPrice: symbolPrice{BidTime: time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local)},
			want:        time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local)},
		{name: "現値時刻が一番新しい時刻なら現値時刻",
			symbolPrice: symbolPrice{
				PriceTime: time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local),
				AskTime:   time.Date(2021, 5, 20, 23, 54, 0, 0, time.Local),
				BidTime:   time.Date(2021, 5, 20, 23, 53, 0, 0, time.Local),
			},
			want: time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local)},
		{name: "最良売り時刻が一番新しい時刻なら最良売り時刻",
			symbolPrice: symbolPrice{
				PriceTime: time.Date(2021, 5, 20, 23, 53, 0, 0, time.Local),
				AskTime:   time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local),
				BidTime:   time.Date(2021, 5, 20, 23, 54, 0, 0, time.Local),
			},
			want: time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local)},
		{name: "最良買い時刻が一番新しい時刻なら最良買い時刻",
			symbolPrice: symbolPrice{
				PriceTime: time.Date(2021, 5, 20, 23, 54, 0, 0, time.Local),
				AskTime:   time.Date(2021, 5, 20, 23, 53, 0, 0, time.Local),
				BidTime:   time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local),
			},
			want: time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local)},
		{name: "3つとも同じ時刻なら結果も同じ",
			symbolPrice: symbolPrice{
				PriceTime: time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local),
				AskTime:   time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local),
				BidTime:   time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local),
			},
			want: time.Date(2021, 5, 20, 23, 55, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.symbolPrice.maxTime()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
