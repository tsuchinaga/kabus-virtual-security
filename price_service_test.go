package virtual_security

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

type testPriceService struct {
	getBySymbolCode1 *symbolPrice
	getBySymbolCode2 error
	set1             error
	validation1      error
	toSymbolPrice1   *symbolPrice
	toSymbolPrice2   error
}

func (t *testPriceService) getBySymbolCode(string) (*symbolPrice, error) {
	return t.getBySymbolCode1, t.getBySymbolCode2
}
func (t *testPriceService) set(*symbolPrice) error { return t.set1 }

func (t *testPriceService) validation(RegisterPriceRequest) error { return t.validation1 }
func (t *testPriceService) toSymbolPrice(RegisterPriceRequest) (*symbolPrice, error) {
	return t.toSymbolPrice1, t.toSymbolPrice2
}

func Test_priceService_validation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg  RegisterPriceRequest
		want error
	}{
		{name: "ExchangeTypeが不明ならエラー", arg: RegisterPriceRequest{ExchangeType: ExchangeTypeUnspecified}, want: InvalidExchangeTypeError},
		{name: "銘柄コードが不明ならエラー", arg: RegisterPriceRequest{ExchangeType: ExchangeTypeStock, SymbolCode: ""}, want: InvalidSymbolCodeError},
		{name: "いずれの時刻もなかったらエラー", arg: RegisterPriceRequest{ExchangeType: ExchangeTypeStock, SymbolCode: "1234"}, want: InvalidTimeError},
		{name: "上記をパスしていればnil",
			arg:  RegisterPriceRequest{ExchangeType: ExchangeTypeStock, SymbolCode: "1234", PriceTime: time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local)},
			want: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &priceService{}
			got := service.validation(test.arg)
			if !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_priceService_toSymbolPrice(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		clock      *testClock
		priceStore *testPriceStore
		arg        RegisterPriceRequest
		want1      *symbolPrice
		want2      error
	}{
		{name: "storeがエラーを吐いたらエラーを返す",
			clock: &testClock{
				getSession1:     SessionMorning,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{
				getBySymbolCode1: nil,
				getBySymbolCode2: NilArgumentError,
			},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
			},
			want2: NilArgumentError},
		{name: "前回の価格がなければ寄りになる",
			clock: &testClock{
				getSession1:     SessionMorning,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
			},
			want1: &symbolPrice{
				ExchangeType:     ExchangeTypeStock,
				SymbolCode:       "1234",
				Price:            1000,
				PriceTime:        time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local),
				Ask:              1010,
				AskTime:          time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
				Bid:              990,
				BidTime:          time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
				kind:             PriceKindOpening,
				session:          SessionMorning,
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			}},
		{name: "前回の価格と営業日が違えば寄り付きになる",
			clock: &testClock{
				getSession1:     SessionMorning,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				priceBusinessDay: time.Date(2021, 6, 29, 0, 0, 0, 0, time.Local),
			}},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
			},
			want1: &symbolPrice{
				ExchangeType:     ExchangeTypeStock,
				SymbolCode:       "1234",
				Price:            1000,
				PriceTime:        time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local),
				Ask:              1010,
				AskTime:          time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
				Bid:              990,
				BidTime:          time.Date(2021, 6, 30, 10, 0, 1, 0, time.Local),
				kind:             PriceKindOpening,
				session:          SessionMorning,
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			}},
		{name: "前回の価格とセッションが違えば寄り付きになる",
			clock: &testClock{
				getSession1:     SessionAfternoon,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
				session:          SessionMorning,
			}},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 14, 0, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 14, 0, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 14, 0, 1, 0, time.Local),
			},
			want1: &symbolPrice{
				ExchangeType:     ExchangeTypeStock,
				SymbolCode:       "1234",
				Price:            1000,
				PriceTime:        time.Date(2021, 6, 30, 14, 0, 0, 0, time.Local),
				Ask:              1010,
				AskTime:          time.Date(2021, 6, 30, 14, 0, 1, 0, time.Local),
				Bid:              990,
				BidTime:          time.Date(2021, 6, 30, 14, 0, 1, 0, time.Local),
				kind:             PriceKindOpening,
				session:          SessionAfternoon,
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			}},
		{name: "前場引後の時間なら引けになる",
			clock: &testClock{
				getSession1:     SessionMorning,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
				session:          SessionMorning,
			}},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 11, 30, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
			},
			want1: &symbolPrice{
				ExchangeType:     ExchangeTypeStock,
				SymbolCode:       "1234",
				Price:            1000,
				PriceTime:        time.Date(2021, 6, 30, 11, 30, 0, 0, time.Local),
				Ask:              1010,
				AskTime:          time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				Bid:              990,
				BidTime:          time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				kind:             PriceKindClosing,
				session:          SessionMorning,
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			}},
		{name: "後場引後の時間なら引けになる",
			clock: &testClock{
				getSession1:     SessionAfternoon,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
				session:          SessionAfternoon,
			}},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 11, 30, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
			},
			want1: &symbolPrice{
				ExchangeType:     ExchangeTypeStock,
				SymbolCode:       "1234",
				Price:            1000,
				PriceTime:        time.Date(2021, 6, 30, 11, 30, 0, 0, time.Local),
				Ask:              1010,
				AskTime:          time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				Bid:              990,
				BidTime:          time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				kind:             PriceKindClosing,
				session:          SessionAfternoon,
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			}},
		{name: "前回の価格がない状態で前場引後の時間なら寄りかつ引けになる",
			clock: &testClock{
				getSession1:     SessionMorning,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{getBySymbolCode1: nil},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 11, 30, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
			},
			want1: &symbolPrice{
				ExchangeType:     ExchangeTypeStock,
				SymbolCode:       "1234",
				Price:            1000,
				PriceTime:        time.Date(2021, 6, 30, 11, 30, 0, 0, time.Local),
				Ask:              1010,
				AskTime:          time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				Bid:              990,
				BidTime:          time.Date(2021, 6, 30, 11, 30, 1, 0, time.Local),
				kind:             PriceKindOpeningAndClosing,
				session:          SessionMorning,
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			}},
		{name: "前回の価格がない状態で後場引後の時間なら寄りかつ引けになる",
			clock: &testClock{
				getSession1:     SessionAfternoon,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{getBySymbolCode1: nil},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 15, 0, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 15, 0, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 15, 0, 1, 0, time.Local),
			},
			want1: &symbolPrice{
				ExchangeType:     ExchangeTypeStock,
				SymbolCode:       "1234",
				Price:            1000,
				PriceTime:        time.Date(2021, 6, 30, 15, 0, 0, 0, time.Local),
				Ask:              1010,
				AskTime:          time.Date(2021, 6, 30, 15, 0, 1, 0, time.Local),
				Bid:              990,
				BidTime:          time.Date(2021, 6, 30, 15, 0, 1, 0, time.Local),
				kind:             PriceKindOpeningAndClosing,
				session:          SessionAfternoon,
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			}},
		{name: "ザラバ中ならザラバになる",
			clock: &testClock{
				getSession1:     SessionAfternoon,
				getBusinessDay1: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			},
			priceStore: &testPriceStore{getBySymbolCode1: &symbolPrice{
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
				session:          SessionAfternoon,
			}},
			arg: RegisterPriceRequest{
				ExchangeType: ExchangeTypeStock,
				SymbolCode:   "1234",
				Price:        1000,
				PriceTime:    time.Date(2021, 6, 30, 14, 0, 0, 0, time.Local),
				Ask:          1010,
				AskTime:      time.Date(2021, 6, 30, 14, 0, 1, 0, time.Local),
				Bid:          990,
				BidTime:      time.Date(2021, 6, 30, 14, 0, 1, 0, time.Local),
			},
			want1: &symbolPrice{
				ExchangeType:     ExchangeTypeStock,
				SymbolCode:       "1234",
				Price:            1000,
				PriceTime:        time.Date(2021, 6, 30, 14, 0, 0, 0, time.Local),
				Ask:              1010,
				AskTime:          time.Date(2021, 6, 30, 14, 0, 1, 0, time.Local),
				Bid:              990,
				BidTime:          time.Date(2021, 6, 30, 14, 0, 1, 0, time.Local),
				kind:             PriceKindRegular,
				session:          SessionAfternoon,
				priceBusinessDay: time.Date(2021, 6, 30, 0, 0, 0, 0, time.Local),
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &priceService{clock: test.clock, priceStore: test.priceStore}
			got1, got2 := service.toSymbolPrice(test.arg)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_priceService_getBySymbolCode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store iPriceStore
		arg   string
		want1 *symbolPrice
		want2 error
	}{
		{name: "storeがerrを返したらserviceもerrを返す",
			arg:   "1234",
			store: &testPriceStore{getBySymbolCode1: nil, getBySymbolCode2: NoDataError},
			want1: nil,
			want2: NoDataError},
		{name: "storeがsymbolPriceを返したらserviceもsymbolPriceを返す",
			arg:   "1234",
			store: &testPriceStore{getBySymbolCode1: &symbolPrice{SymbolCode: "1234"}},
			want1: &symbolPrice{SymbolCode: "1234"},
			want2: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &priceService{priceStore: test.store}
			got1, got2 := service.getBySymbolCode(test.arg)
			if !reflect.DeepEqual(test.want1, got1) || !errors.Is(got2, test.want2) {
				t.Errorf("%s error\nwant: %+v, %+v\ngot: %+v, %+v\n", t.Name(), test.want1, test.want2, got1, got2)
			}
		})
	}
}

func Test_priceService_set(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		store iPriceStore
		arg   *symbolPrice
		want  error
	}{
		{name: "storeからerrがあればそのerrを返す", store: &testPriceStore{set1: NilArgumentError}, arg: nil, want: NilArgumentError},
		{name: "storeからerrがなければnilを返す",
			store: &testPriceStore{set1: nil}, arg: &symbolPrice{SymbolCode: "1234"}, want: nil},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &priceService{priceStore: test.store}
			got := service.set(test.arg)
			if !errors.Is(got, test.want) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_newPriceService(t *testing.T) {
	t.Parallel()
	clock := &testClock{}
	priceStore := &testPriceStore{}

	want := &priceService{
		clock:      clock,
		priceStore: priceStore,
	}
	got := newPriceService(clock, priceStore)

	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}
