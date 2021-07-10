package virtual_security

import (
	"errors"
	"reflect"
	"testing"
)

type testStockOrderStore struct {
	getAll1             []*stockOrder
	getByCode1          *stockOrder
	getByCode2          error
	getByCodeHistory    []string
	add1                error
	addHistory          []*stockOrder
	removeByCodeHistory []string
}

func (t *testStockOrderStore) getAll() []*stockOrder { return t.getAll1 }
func (t *testStockOrderStore) getByCode(code string) (*stockOrder, error) {
	t.getByCodeHistory = append(t.getByCodeHistory, code)
	return t.getByCode1, t.getByCode2
}
func (t *testStockOrderStore) add(order *stockOrder) error {
	if t.addHistory == nil {
		t.addHistory = []*stockOrder{}
	}
	t.addHistory = append(t.addHistory, order)
	return t.add1
}
func (t *testStockOrderStore) removeByCode(code string) {
	if t.removeByCodeHistory == nil {
		t.removeByCodeHistory = []string{}
	}
	t.removeByCodeHistory = append(t.removeByCodeHistory, code)
}

func Test_getStockOrderStore(t *testing.T) {
	t.Parallel()
	stockOrderStoreSingleton = nil
	got := getStockOrderStore()
	want := &stockOrderStore{store: map[string]*stockOrder{}}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_stockOrderStore_GetAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockOrderStore
		want  []*stockOrder
	}{
		{name: "storeが空なら空配列が返される",
			store: &stockOrderStore{store: map[string]*stockOrder{}},
			want:  []*stockOrder{}},
		{name: "storeにデータがあればコード順昇順にして返す",
			store: &stockOrderStore{store: map[string]*stockOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"},
			}},
			want: []*stockOrder{{Code: "bar"}, {Code: "baz"}, {Code: "foo"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.store.getAll()
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
		want1 *stockOrder
		want2 error
	}{
		{name: "storeに指定したコードがなければエラー",
			store: &stockOrderStore{store: map[string]*stockOrder{}},
			arg:   "foo",
			want1: nil,
			want2: NoDataError},
		{name: "storeに指定したコードがあればその注文を返す",
			store: &stockOrderStore{store: map[string]*stockOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:   "foo",
			want1: &stockOrder{Code: "foo"},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.store.getByCode(test.arg)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_stockOrderStore_Add(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		store     *stockOrderStore
		arg       *stockOrder
		want      error
		wantStore map[string]*stockOrder
	}{
		{name: "storeにコードがなければ追加",
			store: &stockOrderStore{store: map[string]*stockOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:       &stockOrder{Code: "hoge"},
			wantStore: map[string]*stockOrder{"foo": {Code: "foo"}, "bar": {Code: "bar"}, "baz": {Code: "baz"}, "hoge": {Code: "hoge"}}},
		{name: "storeにコードがあれば上書き",
			store: &stockOrderStore{store: map[string]*stockOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:       &stockOrder{Code: "foo", ExecutionCondition: StockExecutionConditionMO},
			wantStore: map[string]*stockOrder{"foo": {Code: "foo", ExecutionCondition: StockExecutionConditionMO}, "bar": {Code: "bar"}, "baz": {Code: "baz"}}},
		{name: "引数がnilならエラーを返す",
			store:     &stockOrderStore{store: map[string]*stockOrder{}},
			arg:       nil,
			want:      NilArgumentError,
			wantStore: map[string]*stockOrder{}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.store.add(test.arg)
			if !errors.Is(got, test.want) || !reflect.DeepEqual(test.wantStore, test.store.store) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want, test.wantStore, got, test.store.store)
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
		want  map[string]*stockOrder
	}{
		{name: "指定したコードがなければ何もない",
			store: &stockOrderStore{store: map[string]*stockOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
			arg:   "0000",
			want:  map[string]*stockOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
		{name: "指定したコードがあれば削除する",
			store: &stockOrderStore{store: map[string]*stockOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
			arg:   "2345",
			want:  map[string]*stockOrder{"1234": {Code: "1234"}, "3456": {Code: "3456"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.store.removeByCode(test.arg)
			got := test.store.store
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
