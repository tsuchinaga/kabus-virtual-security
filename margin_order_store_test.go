package virtual_security

import (
	"errors"
	"reflect"
	"testing"
)

type testMarginOrderStore struct {
	getAll1             []*marginOrder
	getByCode1          *marginOrder
	getByCode2          error
	getByCodeHistory    []string
	saveHistory         []*marginOrder
	removeByCodeHistory []string
}

func (t *testMarginOrderStore) getAll() []*marginOrder { return t.getAll1 }
func (t *testMarginOrderStore) getByCode(code string) (*marginOrder, error) {
	t.getByCodeHistory = append(t.getByCodeHistory, code)
	return t.getByCode1, t.getByCode2
}
func (t *testMarginOrderStore) save(order *marginOrder) {
	if t.saveHistory == nil {
		t.saveHistory = []*marginOrder{}
	}
	t.saveHistory = append(t.saveHistory, order)
}
func (t *testMarginOrderStore) removeByCode(code string) {
	if t.removeByCodeHistory == nil {
		t.removeByCodeHistory = []string{}
	}
	t.removeByCodeHistory = append(t.removeByCodeHistory, code)
}

func Test_getMarginOrderStore(t *testing.T) {
	got := getMarginOrderStore()
	want := &marginOrderStore{store: map[string]*marginOrder{}}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_marginOrderStore_GetAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *marginOrderStore
		want  []*marginOrder
	}{
		{name: "storeが空なら空配列が返される",
			store: &marginOrderStore{store: map[string]*marginOrder{}},
			want:  []*marginOrder{}},
		{name: "storeにデータがあればコード順昇順にして返す",
			store: &marginOrderStore{store: map[string]*marginOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"},
			}},
			want: []*marginOrder{{Code: "bar"}, {Code: "baz"}, {Code: "foo"}}},
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

func Test_marginOrderStore_GetByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *marginOrderStore
		arg   string
		want1 *marginOrder
		want2 error
	}{
		{name: "storeに指定したコードがなければエラー",
			store: &marginOrderStore{store: map[string]*marginOrder{}},
			arg:   "foo",
			want1: nil,
			want2: NoDataError},
		{name: "storeに指定したコードがあればその注文を返す",
			store: &marginOrderStore{store: map[string]*marginOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:   "foo",
			want1: &marginOrder{Code: "foo"},
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

func Test_marginOrderStore_save(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		store     *marginOrderStore
		arg       *marginOrder
		wantStore map[string]*marginOrder
	}{
		{name: "引数がnilなら何もしない",
			store: &marginOrderStore{store: map[string]*marginOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:       nil,
			wantStore: map[string]*marginOrder{"foo": {Code: "foo"}, "bar": {Code: "bar"}, "baz": {Code: "baz"}}},
		{name: "storeにコードがなければ追加",
			store: &marginOrderStore{store: map[string]*marginOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:       &marginOrder{Code: "hoge"},
			wantStore: map[string]*marginOrder{"foo": {Code: "foo"}, "bar": {Code: "bar"}, "baz": {Code: "baz"}, "hoge": {Code: "hoge"}}},
		{name: "storeにコードがあれば上書き",
			store: &marginOrderStore{store: map[string]*marginOrder{
				"foo": {Code: "foo"},
				"bar": {Code: "bar"},
				"baz": {Code: "baz"}}},
			arg:       &marginOrder{Code: "foo", ExecutionCondition: StockExecutionConditionMO},
			wantStore: map[string]*marginOrder{"foo": {Code: "foo", ExecutionCondition: StockExecutionConditionMO}, "bar": {Code: "bar"}, "baz": {Code: "baz"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.store.save(test.arg)
			if !reflect.DeepEqual(test.wantStore, test.store.store) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.wantStore, test.store.store)
			}
		})
	}
}

func Test_marginOrderStore_RemoveByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *marginOrderStore
		arg   string
		want  map[string]*marginOrder
	}{
		{name: "指定したコードがなければ何もない",
			store: &marginOrderStore{store: map[string]*marginOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
			arg:   "0000",
			want:  map[string]*marginOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
		{name: "指定したコードがあれば削除する",
			store: &marginOrderStore{store: map[string]*marginOrder{"1234": {Code: "1234"}, "2345": {Code: "2345"}, "3456": {Code: "3456"}}},
			arg:   "2345",
			want:  map[string]*marginOrder{"1234": {Code: "1234"}, "3456": {Code: "3456"}}},
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
