package virtual_security

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func Test_NewPriceStore(t *testing.T) {
	t.Parallel()

	clock := &testClock{now: time.Date(2021, 5, 22, 7, 11, 0, 0, time.Local)}
	got := NewPriceStore(clock)
	want := &priceStore{
		store:      map[string]SymbolPrice{},
		clock:      clock,
		expireTime: time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local),
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_priceStore_isExpired(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		priceStore *priceStore
		arg        time.Time
		want       bool
	}{
		{name: "有効期限より引数の時刻が前",
			priceStore: &priceStore{expireTime: time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local)},
			arg:        time.Date(2021, 5, 22, 7, 0, 0, 0, time.Local),
			want:       false},
		{name: "有効期限と引数の時刻が一致",
			priceStore: &priceStore{expireTime: time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local)},
			arg:        time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local),
			want:       true},
		{name: "有効期限より引数の時刻が後",
			priceStore: &priceStore{expireTime: time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local)},
			arg:        time.Date(2021, 5, 22, 9, 0, 0, 0, time.Local),
			want:       true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.priceStore.isExpired(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_priceStore_setCalculatedExpireTime(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		priceStore *priceStore
		arg        time.Time
		want       time.Time
	}{
		{name: "現在時刻が8時以前なら当日の8時をセット",
			priceStore: &priceStore{},
			arg:        time.Date(2021, 5, 22, 7, 0, 0, 0, time.Local),
			want:       time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local)},
		{name: "現在時刻が8時なら翌日の8時をセット",
			priceStore: &priceStore{},
			arg:        time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local),
			want:       time.Date(2021, 5, 23, 8, 0, 0, 0, time.Local)},
		{name: "現在時刻が8時以降なら翌日の8時をセット",
			priceStore: &priceStore{},
			arg:        time.Date(2021, 5, 22, 9, 0, 0, 0, time.Local),
			want:       time.Date(2021, 5, 23, 8, 0, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.priceStore.setCalculatedExpireTime(test.arg)
			got := test.priceStore.expireTime
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_priceStore_GetBySymbolCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		priceStore *priceStore
		arg        string
		want1      SymbolPrice
		want2      error
	}{
		{name: "指定した銘柄が存在したらそれを返す",
			priceStore: &priceStore{
				clock:      &testClock{now: time.Date(2021, 5, 22, 7, 0, 0, 0, time.Local)},
				store:      map[string]SymbolPrice{"1234": {SymbolCode: "1234", Price: 100}, "2345": {SymbolCode: "2345", Price: 200}, "3456": {SymbolCode: "3456", Price: 300}},
				expireTime: time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local),
			},
			arg:   "2345",
			want1: SymbolPrice{SymbolCode: "2345", Price: 200},
			want2: nil},
		{name: "指定した銘柄が存在しなければエラーを返す",
			priceStore: &priceStore{
				clock:      &testClock{now: time.Date(2021, 5, 22, 7, 0, 0, 0, time.Local)},
				store:      map[string]SymbolPrice{"1234": {SymbolCode: "1234", Price: 100}, "2345": {SymbolCode: "2345", Price: 200}, "3456": {SymbolCode: "3456", Price: 300}},
				expireTime: time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local),
			},
			arg:   "0000",
			want1: SymbolPrice{},
			want2: NoDataError},
		{name: "指定した銘柄が存在しても、有効期限が切れていればstoreを空にして有効期限を更新し、エラーを返す",
			priceStore: &priceStore{
				clock:      &testClock{now: time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local)},
				store:      map[string]SymbolPrice{"1234": {SymbolCode: "1234", Price: 100}, "2345": {SymbolCode: "2345", Price: 200}, "3456": {SymbolCode: "3456", Price: 300}},
				expireTime: time.Date(2021, 5, 22, 8, 0, 0, 0, time.Local),
			},
			arg:   "2345",
			want1: SymbolPrice{},
			want2: ExpiredDataError},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.priceStore.GetBySymbolCode(test.arg)
			if !reflect.DeepEqual(test.want1, got1) && !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_priceStore_Set(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		priceStore     *priceStore
		arg            SymbolPrice
		wantStore      map[string]SymbolPrice
		wantExpireTime time.Time
	}{
		{name: "引数の情報に更新日時がまったくなければ何もしない",
			priceStore: &priceStore{
				store:      map[string]SymbolPrice{},
				expireTime: time.Date(2021, 5, 23, 8, 0, 0, 0, time.Local)},
			arg:            SymbolPrice{},
			wantStore:      map[string]SymbolPrice{},
			wantExpireTime: time.Date(2021, 5, 23, 8, 0, 0, 0, time.Local)},
		{name: "ストアの有効期限が切れていなければ、storeに追加する",
			priceStore: &priceStore{
				clock:      &testClock{now: time.Date(2021, 5, 25, 9, 0, 0, 0, time.Local)},
				store:      map[string]SymbolPrice{"1234": {SymbolCode: "1234", Price: 100}, "2345": {SymbolCode: "2345", Price: 200}, "3456": {SymbolCode: "3456", Price: 300}},
				expireTime: time.Date(2021, 5, 26, 8, 0, 0, 0, time.Local)},
			arg:            SymbolPrice{SymbolCode: "2345", Price: 400, PriceTime: time.Date(2021, 5, 25, 9, 0, 0, 0, time.Local)},
			wantStore:      map[string]SymbolPrice{"1234": {SymbolCode: "1234", Price: 100}, "2345": {SymbolCode: "2345", Price: 400, PriceTime: time.Date(2021, 5, 25, 9, 0, 0, 0, time.Local)}, "3456": {SymbolCode: "3456", Price: 300}},
			wantExpireTime: time.Date(2021, 5, 26, 8, 0, 0, 0, time.Local)},
		{name: "有効期限が切れていれば、storeをクリアし、有効期限を延長してから、storeに追加する",
			priceStore: &priceStore{
				clock:      &testClock{now: time.Date(2021, 5, 25, 9, 0, 0, 0, time.Local)},
				store:      map[string]SymbolPrice{"1234": {SymbolCode: "1234", Price: 100}, "2345": {SymbolCode: "2345", Price: 200}, "3456": {SymbolCode: "3456", Price: 300}},
				expireTime: time.Date(2021, 5, 25, 8, 0, 0, 0, time.Local)},
			arg:            SymbolPrice{SymbolCode: "2345", Price: 400, PriceTime: time.Date(2021, 5, 25, 9, 0, 0, 0, time.Local)},
			wantStore:      map[string]SymbolPrice{"2345": {SymbolCode: "2345", Price: 400, PriceTime: time.Date(2021, 5, 25, 9, 0, 0, 0, time.Local)}},
			wantExpireTime: time.Date(2021, 5, 26, 8, 0, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.priceStore.Set(test.arg)
			if !reflect.DeepEqual(test.wantStore, test.priceStore.store) || !reflect.DeepEqual(test.wantExpireTime, test.priceStore.expireTime) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.wantStore, test.wantExpireTime, test.priceStore.store, test.priceStore.expireTime)
			}
		})
	}
}
