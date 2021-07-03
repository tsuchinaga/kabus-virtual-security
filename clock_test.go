package virtual_security

import (
	"reflect"
	"testing"
	"time"
)

type testClock struct {
	Clock
	now                    time.Time
	getStockSession        Session
	getStockSessionHistory []time.Time
	getSession             Session
	getBusinessDay         time.Time
}

func (t *testClock) Now() time.Time { return t.now }
func (t *testClock) GetStockSession(now time.Time) Session {
	t.getStockSessionHistory = append(t.getStockSessionHistory, now)
	return t.getStockSession
}
func (t *testClock) GetSession(ExchangeType, time.Time) Session       { return t.getSession }
func (t *testClock) GetBusinessDay(ExchangeType, time.Time) time.Time { return t.getBusinessDay }

func Test_newClock(t *testing.T) {
	want := &clock{}
	got := newClock()
	t.Parallel()
	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_clock_Now1(t *testing.T) {
	want := time.Now()
	got := (&clock{}).Now()

	t.Parallel()
	if got.Before(want) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_clock_Now2(t *testing.T) {
	got := (&clock{}).Now()
	want := time.Now()

	t.Parallel()
	if got.After(want) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_clock_GetStockSession(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg  time.Time
		want Session
	}{
		{name: "前場前 なら unspecified", arg: time.Date(0, 1, 1, 8, 59, 59, 0, time.Local), want: SessionUnspecified},
		{name: "前場開始時刻 なら morning", arg: time.Date(0, 1, 1, 9, 0, 0, 0, time.Local), want: SessionMorning},
		{name: "前場中 なら morning", arg: time.Date(0, 1, 1, 10, 0, 0, 0, time.Local), want: SessionMorning},
		{name: "前場終了時刻 なら morning", arg: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local), want: SessionMorning},
		{name: "前場終了後 なら morning", arg: time.Date(0, 1, 1, 11, 30, 5, 0, time.Local), want: SessionUnspecified},
		{name: "前場後・後場前 なら unspecified", arg: time.Date(0, 1, 1, 12, 0, 0, 0, time.Local), want: SessionUnspecified},
		{name: "後場開始時刻 なら afternoon", arg: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), want: SessionAfternoon},
		{name: "後場中 なら afternoon", arg: time.Date(0, 1, 1, 13, 0, 0, 0, time.Local), want: SessionAfternoon},
		{name: "後場終了時刻 なら afternoon", arg: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local), want: SessionAfternoon},
		{name: "後場終了後 なら unspecified", arg: time.Date(0, 1, 1, 15, 0, 5, 0, time.Local), want: SessionUnspecified},
		{name: "後場後 なら unspecified", arg: time.Date(0, 1, 1, 15, 0, 6, 0, time.Local), want: SessionUnspecified},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clock := &clock{}
			got := clock.GetStockSession(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_clock_GetBusinessDay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 ExchangeType
		arg2 time.Time
		want time.Time
	}{
		{name: "引数がゼロ値ならそのまま返す", arg1: ExchangeTypeUnspecified, arg2: time.Time{}, want: time.Time{}},
		{name: "現物なら年月日をそのまま営業日にして返す",
			arg1: ExchangeTypeStock,
			arg2: time.Date(2021, 6, 29, 16, 29, 0, 0, time.Local),
			want: time.Date(2021, 6, 29, 0, 0, 0, 0, time.Local)},
		{name: "信用なら年月日をそのまま営業日にして返す",
			arg1: ExchangeTypeMargin,
			arg2: time.Date(2021, 6, 29, 16, 29, 0, 0, time.Local),
			want: time.Date(2021, 6, 29, 0, 0, 0, 0, time.Local)},
		{name: "上記以外は年月日をそのまま返す",
			arg1: ExchangeTypeUnspecified,
			arg2: time.Date(2021, 6, 29, 16, 29, 0, 0, time.Local),
			want: time.Date(2021, 6, 29, 0, 0, 0, 0, time.Local)},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			c := &clock{}
			got := c.GetBusinessDay(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_clock_GetSession(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		arg1 ExchangeType
		arg2 time.Time
		want Session
	}{
		{name: "日時がゼロ値なら未指定",
			arg1: ExchangeTypeStock,
			arg2: time.Time{},
			want: SessionUnspecified},
		{name: "ExchangeTypeが未指定なら引数も未指定",
			arg1: ExchangeTypeUnspecified,
			arg2: time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local),
			want: SessionUnspecified},
		{name: "現物で前場の時間なら前場",
			arg1: ExchangeTypeStock,
			arg2: time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local),
			want: SessionMorning},
		{name: "現物で後場の時間なら後場",
			arg1: ExchangeTypeStock,
			arg2: time.Date(2021, 6, 30, 13, 0, 0, 0, time.Local),
			want: SessionAfternoon},
		{name: "現物で上記以外の時間なら未指定",
			arg1: ExchangeTypeStock,
			arg2: time.Date(2021, 6, 30, 12, 0, 0, 0, time.Local),
			want: SessionUnspecified},
		{name: "信用で前場の時間なら前場",
			arg1: ExchangeTypeMargin,
			arg2: time.Date(2021, 6, 30, 10, 0, 0, 0, time.Local),
			want: SessionMorning},
		{name: "信用で後場の時間なら後場",
			arg1: ExchangeTypeMargin,
			arg2: time.Date(2021, 6, 30, 13, 0, 0, 0, time.Local),
			want: SessionAfternoon},
		{name: "信用で上記以外の時間なら未指定",
			arg1: ExchangeTypeMargin,
			arg2: time.Date(2021, 6, 30, 12, 0, 0, 0, time.Local),
			want: SessionUnspecified},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			clock := &clock{}
			got := clock.GetSession(test.arg1, test.arg2)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
