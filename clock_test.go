package virtual_security

import (
	"reflect"
	"testing"
	"time"
)

type testClock struct {
	Clock
	now time.Time
}

func (t *testClock) Now() time.Time { return t.now }

func Test_NewClock(t *testing.T) {
	want := &clock{}
	got := NewClock()
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
