package virtual_security

import (
	"errors"
	"reflect"
	"testing"
)

type testStockPositionStore struct {
	getAll              []*stockPosition
	getByCode1          *stockPosition
	getByCode2          error
	getByCodeHistory    []string
	addHistory          []*stockPosition
	removeByCodeHistory []string
}

func (t *testStockPositionStore) GetAll() []*stockPosition { return t.getAll }
func (t *testStockPositionStore) GetByCode(code string) (*stockPosition, error) {
	t.getByCodeHistory = append(t.getByCodeHistory, code)
	return t.getByCode1, t.getByCode2
}
func (t *testStockPositionStore) Add(stockPosition *stockPosition) {
	t.addHistory = append(t.addHistory, stockPosition)
}
func (t testStockPositionStore) RemoveByCode(code string) {
	t.removeByCodeHistory = append(t.removeByCodeHistory, code)
}

func Test_getStockPositionStore(t *testing.T) {
	t.Parallel()
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
			got := test.store.GetAll()
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
			got1, got2 := test.store.GetByCode(test.arg)
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
			test.store.Add(test.arg)
			got := test.store.store
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_positionStockStore_RemoveByCode(t *testing.T) {
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
			test.store.RemoveByCode(test.arg)
			got := test.store.store
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
