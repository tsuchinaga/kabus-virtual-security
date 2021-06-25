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
}

func (t *testClock) Now() time.Time { return t.now }
func (t *testClock) GetStockSession(now time.Time) Session {
	t.getStockSessionHistory = append(t.getStockSessionHistory, now)
	return t.getStockSession
}

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
		{name: "前場終了時刻 なら unspecified", arg: time.Date(0, 1, 1, 11, 30, 0, 0, time.Local), want: SessionUnspecified},
		{name: "前場後・後場前 なら unspecified", arg: time.Date(0, 1, 1, 12, 0, 0, 0, time.Local), want: SessionUnspecified},
		{name: "後場開始時刻 なら afternoon", arg: time.Date(0, 1, 1, 12, 30, 0, 0, time.Local), want: SessionAfternoon},
		{name: "後場中 なら afternoon", arg: time.Date(0, 1, 1, 13, 0, 0, 0, time.Local), want: SessionAfternoon},
		{name: "後場終了時刻 なら unspecified", arg: time.Date(0, 1, 1, 15, 0, 0, 0, time.Local), want: SessionUnspecified},
		{name: "後場後 なら unspecified", arg: time.Date(0, 1, 1, 15, 0, 1, 0, time.Local), want: SessionUnspecified},
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
