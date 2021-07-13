package virtual_security

import (
	"errors"
	"reflect"
	"testing"
)

type testStockPositionStore struct {
	getAll1                []*stockPosition
	getByCode1             *stockPosition
	getByCode2             error
	getByCodeHistory       []string
	getBySymbolCode1       []*stockPosition
	getBySymbolCode2       error
	getBySymbolCodeHistory []string
	addHistory             []*stockPosition
	removeByCodeHistory    []string
}

func (t *testStockPositionStore) getAll() []*stockPosition { return t.getAll1 }
func (t *testStockPositionStore) getByCode(code string) (*stockPosition, error) {
	t.getByCodeHistory = append(t.getByCodeHistory, code)
	return t.getByCode1, t.getByCode2
}
func (t *testStockPositionStore) getBySymbolCode(symbolCode string) ([]*stockPosition, error) {
	t.getBySymbolCodeHistory = append(t.getBySymbolCodeHistory, symbolCode)
	return t.getBySymbolCode1, t.getBySymbolCode2
}
func (t *testStockPositionStore) add(stockPosition *stockPosition) {
	t.addHistory = append(t.addHistory, stockPosition)
}
func (t *testStockPositionStore) removeByCode(code string) {
	t.removeByCodeHistory = append(t.removeByCodeHistory, code)
}

func Test_getStockPositionStore(t *testing.T) {
	got := getStockPositionStore()
	want := &stockPositionStore{
		store: map[string]*stockPosition{},
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_stockPositionStore_GetAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockPositionStore
		want  []*stockPosition
	}{
		{name: "storeが空なら空配列を返す",
			store: &stockPositionStore{store: map[string]*stockPosition{}},
			want:  []*stockPosition{}},
		{name: "storeが空でないなら配列にして返す",
			store: &stockPositionStore{store: map[string]*stockPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
			}},
			want: []*stockPosition{
				{Code: "pos_1234"},
				{Code: "pos_2345"},
				{Code: "pos_3456"},
			}},
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

func Test_stockPositionStore_GetByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockPositionStore
		arg   string
		want1 *stockPosition
		want2 error
	}{
		{name: "引数にマッチするデータがあればデータを返す",
			store: &stockPositionStore{store: map[string]*stockPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
			}},
			arg:   "pos_2345",
			want1: &stockPosition{Code: "pos_2345"},
			want2: nil,
		},
		{name: "引数にマッチするデータがなければエラーを返す",
			store: &stockPositionStore{store: map[string]*stockPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
			}},
			arg:   "pos_0000",
			want1: nil,
			want2: NoDataError,
		},
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

func Test_stockPositionStore_Add(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockPositionStore
		arg   *stockPosition
		want  map[string]*stockPosition
	}{
		{name: "キーが重複しなければ新しいデータが追加される",
			store: &stockPositionStore{
				store: map[string]*stockPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: &stockPosition{Code: "pos_9999"},
			want: map[string]*stockPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
				"pos_9999": {Code: "pos_9999"},
			}},
		{name: "キーが重複したら新しいデータを上書きする",
			store: &stockPositionStore{
				store: map[string]*stockPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: &stockPosition{Code: "pos_2345", OrderCode: "ord_5555"},
			want: map[string]*stockPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345", OrderCode: "ord_5555"},
				"pos_3456": {Code: "pos_3456"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.store.add(test.arg)
			got := test.store.store
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_stockPositionStore_RemoveByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *stockPositionStore
		arg   string
		want  map[string]*stockPosition
	}{
		{name: "指定したコードがなければ何もしない",
			store: &stockPositionStore{
				store: map[string]*stockPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: "pos_0000",
			want: map[string]*stockPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
			}},
		{name: "指定したコードがあればstoreから消す",
			store: &stockPositionStore{
				store: map[string]*stockPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: "pos_2345",
			want: map[string]*stockPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_3456": {Code: "pos_3456"},
			}},
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

func Test_stockPositionStore_getBySymbolCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store iStockPositionStore
		arg   string
		want1 []*stockPosition
		want2 error
	}{
		{name: "storeにデータがないなら空配列が返される",
			store: &stockPositionStore{store: map[string]*stockPosition{}},
			arg:   "1234",
			want1: []*stockPosition{}},
		{name: "storeに指定した銘柄コードと一致するデータがないなら空配列が返される",
			store: &stockPositionStore{store: map[string]*stockPosition{
				"spo-1": {Code: "spo-1", SymbolCode: "0001"},
				"spo-2": {Code: "spo-2", SymbolCode: "0002"},
				"spo-3": {Code: "spo-3", SymbolCode: "0003"},
			}},
			arg:   "1234",
			want1: []*stockPosition{}},
		{name: "storeに指定した銘柄コードと一致するデータがあればポジションコード順に並べて返される",
			store: &stockPositionStore{store: map[string]*stockPosition{
				"spo-1": {Code: "spo-1", SymbolCode: "1234"},
				"spo-2": {Code: "spo-2", SymbolCode: "0002"},
				"spo-3": {Code: "spo-3", SymbolCode: "1234"},
			}},
			arg: "1234",
			want1: []*stockPosition{
				{Code: "spo-1", SymbolCode: "1234"},
				{Code: "spo-3", SymbolCode: "1234"}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got1, got2 := test.store.getBySymbolCode(test.arg)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}
