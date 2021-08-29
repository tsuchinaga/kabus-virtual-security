package virtual_security

import (
	"errors"
	"reflect"
	"testing"
)

type testMarginPositionStore struct {
	getAll1                []*marginPosition
	getByCode1             *marginPosition
	getByCode2             error
	getByCodeHistory       []string
	getBySymbolCode1       []*marginPosition
	getBySymbolCode2       error
	getBySymbolCodeHistory []string
	saveHistory            []*marginPosition
	removeByCodeHistory    []string
}

func (t *testMarginPositionStore) getAll() []*marginPosition { return t.getAll1 }
func (t *testMarginPositionStore) getByCode(code string) (*marginPosition, error) {
	t.getByCodeHistory = append(t.getByCodeHistory, code)
	return t.getByCode1, t.getByCode2
}
func (t *testMarginPositionStore) getBySymbolCode(symbolCode string) ([]*marginPosition, error) {
	t.getBySymbolCodeHistory = append(t.getBySymbolCodeHistory, symbolCode)
	return t.getBySymbolCode1, t.getBySymbolCode2
}
func (t *testMarginPositionStore) save(marginPosition *marginPosition) {
	t.saveHistory = append(t.saveHistory, marginPosition)
}
func (t *testMarginPositionStore) removeByCode(code string) {
	t.removeByCodeHistory = append(t.removeByCodeHistory, code)
}

func Test_getMarginPositionStore(t *testing.T) {
	got := getMarginPositionStore()
	want := &marginPositionStore{
		store: map[string]*marginPosition{},
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_marginPositionStore_GetAll(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *marginPositionStore
		want  []*marginPosition
	}{
		{name: "storeが空なら空配列を返す",
			store: &marginPositionStore{store: map[string]*marginPosition{}},
			want:  []*marginPosition{}},
		{name: "storeが空でないなら配列にして返す",
			store: &marginPositionStore{store: map[string]*marginPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
			}},
			want: []*marginPosition{
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

func Test_marginPositionStore_GetByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *marginPositionStore
		arg   string
		want1 *marginPosition
		want2 error
	}{
		{name: "引数にマッチするデータがあればデータを返す",
			store: &marginPositionStore{store: map[string]*marginPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
			}},
			arg:   "pos_2345",
			want1: &marginPosition{Code: "pos_2345"},
			want2: nil,
		},
		{name: "引数にマッチするデータがなければエラーを返す",
			store: &marginPositionStore{store: map[string]*marginPosition{
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

func Test_marginPositionStore_Save(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *marginPositionStore
		arg   *marginPosition
		want  map[string]*marginPosition
	}{
		{name: "引数がnilなら何もしない",
			store: &marginPositionStore{
				store: map[string]*marginPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: nil,
			want: map[string]*marginPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
			}},
		{name: "キーが重複しなければ新しいデータが追加される",
			store: &marginPositionStore{
				store: map[string]*marginPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: &marginPosition{Code: "pos_9999"},
			want: map[string]*marginPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
				"pos_9999": {Code: "pos_9999"},
			}},
		{name: "キーが重複したら新しいデータを上書きする",
			store: &marginPositionStore{
				store: map[string]*marginPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: &marginPosition{Code: "pos_2345", OrderCode: "ord_5555"},
			want: map[string]*marginPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345", OrderCode: "ord_5555"},
				"pos_3456": {Code: "pos_3456"},
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.store.save(test.arg)
			got := test.store.store
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_marginPositionStore_RemoveByCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store *marginPositionStore
		arg   string
		want  map[string]*marginPosition
	}{
		{name: "指定したコードがなければ何もしない",
			store: &marginPositionStore{
				store: map[string]*marginPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: "pos_0000",
			want: map[string]*marginPosition{
				"pos_1234": {Code: "pos_1234"},
				"pos_2345": {Code: "pos_2345"},
				"pos_3456": {Code: "pos_3456"},
			}},
		{name: "指定したコードがあればstoreから消す",
			store: &marginPositionStore{
				store: map[string]*marginPosition{
					"pos_1234": {Code: "pos_1234"},
					"pos_2345": {Code: "pos_2345"},
					"pos_3456": {Code: "pos_3456"},
				}},
			arg: "pos_2345",
			want: map[string]*marginPosition{
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

func Test_marginPositionStore_getBySymbolCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store iMarginPositionStore
		arg   string
		want1 []*marginPosition
		want2 error
	}{
		{name: "storeにデータがないなら空配列が返される",
			store: &marginPositionStore{store: map[string]*marginPosition{}},
			arg:   "1234",
			want1: []*marginPosition{}},
		{name: "storeに指定した銘柄コードと一致するデータがないなら空配列が返される",
			store: &marginPositionStore{store: map[string]*marginPosition{
				"spo-1": {Code: "spo-1", SymbolCode: "0001"},
				"spo-2": {Code: "spo-2", SymbolCode: "0002"},
				"spo-3": {Code: "spo-3", SymbolCode: "0003"},
			}},
			arg:   "1234",
			want1: []*marginPosition{}},
		{name: "storeに指定した銘柄コードと一致するデータがあればポジションコード順に並べて返される",
			store: &marginPositionStore{store: map[string]*marginPosition{
				"spo-1": {Code: "spo-1", SymbolCode: "1234"},
				"spo-2": {Code: "spo-2", SymbolCode: "0002"},
				"spo-3": {Code: "spo-3", SymbolCode: "1234"},
			}},
			arg: "1234",
			want1: []*marginPosition{
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
