package virtual_security

import (
	"errors"
	"reflect"
	"testing"
)

func Test_GetStockOrderStore(t *testing.T) {
	t.Parallel()
	got := GetStockOrderStore()
	want := &stockOrderStore{store: map[string]*StockOrder{}}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_stockOrderStore_GetAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockOrderStore
		want  []*StockOrder
	}{
		{name: "storeが空なら空配列が返される",
			store: &stockOrderStore{store: map[string]*StockOrder{}},
			want:  []*StockOrder{}},
		{name: "storeにデータがあればコード順昇順にして返す",
			store: &stockOrderStore{store: map[string]*StockOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"},
			}},
			want: []*StockOrder{{Code: "bar"}, {Code: "baz"}, {Code: "foo"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.store.GetAll()
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockOrderStore_GetByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockOrderStore
		arg   string
		want1 *StockOrder
		want2 error
	}{
		{name: "storeに指定したコードがなければエラー",
			store: &stockOrderStore{store: map[string]*StockOrder{}},
			arg:   "foo",
			want1: nil,
			want2: NoDataError},
		{name: "storeに指定したコードがあればその注文を返す",
			store: &stockOrderStore{store: map[string]*StockOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:   "foo",
			want1: &StockOrder{Code: "foo"},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.store.GetByCode(test.arg)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_stockOrderStore_Add(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockOrderStore
		arg   *StockOrder
		want  map[string]*StockOrder
	}{
		{name: "storeにコードがなければ追加",
			store: &stockOrderStore{store: map[string]*StockOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:  &StockOrder{Code: "hoge"},
			want: map[string]*StockOrder{"foo": {Code: "foo"}, "bar": {Code: "bar"}, "baz": {Code: "baz"}, "hoge": {Code: "hoge"}}},
		{name: "storeにコードがあれば上書き",
			store: &stockOrderStore{store: map[string]*StockOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:  &StockOrder{Code: "foo", ExecutionCondition: StockExecutionConditionMO},
			want: map[string]*StockOrder{"foo": {Code: "foo", ExecutionCondition: StockExecutionConditionMO}, "bar": {Code: "bar"}, "baz": {Code: "baz"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.store.Add(test.arg)
			got := test.store.store
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockOrderStore_RemoveByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockOrderStore
		arg   string
		want  map[string]*StockOrder
	}{
		{name: "指定したコードがなければ何もない",
			store: &stockOrderStore{store: map[string]*StockOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
			arg:   "0000",
			want:  map[string]*StockOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
		{name: "指定したコードがあれば削除する",
			store: &stockOrderStore{store: map[string]*StockOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
			arg:   "2345",
			want:  map[string]*StockOrder{"1234": {Code: "1234"}, "3456": {Code: "3456"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.store.RemoveByCode(test.arg)
			got := test.store.store
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
