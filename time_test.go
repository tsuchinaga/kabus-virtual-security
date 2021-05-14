package virtual_security

import (
	"reflect"
	"testing"
	"time"
)

func Test_newTimeRange(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name                               string
		arg1, arg2, arg3, arg4, arg5, arg6 int
		want                               *timeRange
	}{
		{name: "from < to", arg1: 9, arg2: 0, arg3: 0, arg4: 15, arg5: 0, arg6: 0,
			want: &timeRange{
				from:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				to:     time.Date(0, 1, 1, 15, 0, 0, 0, time.Local),
				isBack: false,
			}},
		{name: "from = to", arg1: 9, arg2: 0, arg3: 0, arg4: 9, arg5: 0, arg6: 0,
			want: &timeRange{
				from:   time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				to:     time.Date(0, 1, 1, 9, 0, 0, 0, time.Local),
				isBack: true,
			}},
		{name: "from > to", arg1: 16, arg2: 30, arg3: 0, arg4: 5, arg5: 30, arg6: 5,
			want: &timeRange{
				from:   time.Date(0, 1, 1, 16, 30, 0, 0, time.Local),
				to:     time.Date(0, 1, 1, 5, 30, 5, 0, time.Local),
				isBack: true,
			}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := newTimeRange(test.arg1, test.arg2, test.arg3, test.arg4, test.arg5, test.arg6)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_timeRange_between(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		timeRange *timeRange
		arg       time.Time
		want      bool
	}{
		{name: "from < toでdayStartと一致",
			timeRange: newTimeRange(9, 0, 0, 15, 0, 5),
			arg:       time.Date(2021, 5, 4, 0, 0, 0, 0, time.Local),
			want:      false},
		{name: "from < toでdayStartとfromの間",
			timeRange: newTimeRange(9, 0, 0, 15, 0, 5),
			arg:       time.Date(2021, 5, 4, 8, 59, 59, 999999999, time.Local),
			want:      false},
		{name: "from < toでfromと一致",
			timeRange: newTimeRange(9, 0, 0, 15, 0, 5),
			arg:       time.Date(2021, 5, 4, 9, 0, 0, 0, time.Local),
			want:      true},
		{name: "from < toでfromとtoの間",
			timeRange: newTimeRange(9, 0, 0, 15, 0, 5),
			arg:       time.Date(2021, 5, 4, 12, 0, 0, 0, time.Local),
			want:      true},
		{name: "from < toでtoと一致",
			timeRange: newTimeRange(9, 0, 0, 15, 0, 5),
			arg:       time.Date(2021, 5, 4, 15, 0, 5, 0, time.Local),
			want:      false},
		{name: "from < toでtoとdayEndの間",
			timeRange: newTimeRange(9, 0, 0, 15, 0, 5),
			arg:       time.Date(2021, 5, 4, 15, 0, 6, 0, time.Local),
			want:      false},
		{name: "from < toでdayEndと一致",
			timeRange: newTimeRange(9, 0, 0, 15, 0, 5),
			arg:       time.Date(2021, 5, 4, 23, 59, 59, 999999999, time.Local),
			want:      false},
		{name: "from = toでdayStartとfromの間",
			timeRange: newTimeRange(7, 0, 0, 7, 0, 0),
			arg:       time.Date(2021, 5, 4, 6, 59, 59, 999999999, time.Local),
			want:      true},
		{name: "from = toでfromとtoと一致",
			timeRange: newTimeRange(7, 0, 0, 7, 0, 0),
			arg:       time.Date(2021, 5, 4, 7, 0, 0, 0, time.Local),
			want:      true},
		{name: "from = toでtoとdayEndの間",
			timeRange: newTimeRange(7, 0, 0, 7, 0, 0),
			arg:       time.Date(2021, 5, 4, 23, 59, 59, 999999999, time.Local),
			want:      true},
		{name: "from > toでdayStartと一致",
			timeRange: newTimeRange(16, 30, 0, 5, 30, 5),
			arg:       time.Date(2021, 5, 4, 0, 0, 0, 0, time.Local),
			want:      true},
		{name: "from > toでdayStartとtoの間",
			timeRange: newTimeRange(16, 30, 0, 5, 30, 5),
			arg:       time.Date(2021, 5, 4, 5, 30, 4, 999999999, time.Local),
			want:      true},
		{name: "from > toでtoと一致",
			timeRange: newTimeRange(16, 30, 0, 5, 30, 5),
			arg:       time.Date(2021, 5, 4, 5, 30, 5, 0, time.Local),
			want:      false},
		{name: "from > toでfromとtoの間",
			timeRange: newTimeRange(16, 30, 0, 5, 30, 5),
			arg:       time.Date(2021, 5, 4, 5, 30, 6, 0, time.Local),
			want:      false},
		{name: "from > toでfromと一致",
			timeRange: newTimeRange(16, 30, 0, 5, 30, 5),
			arg:       time.Date(2021, 5, 4, 16, 30, 0, 0, time.Local),
			want:      true},
		{name: "from > toでfromとdayEndの間",
			timeRange: newTimeRange(16, 30, 0, 5, 30, 5),
			arg:       time.Date(2021, 5, 4, 16, 30, 1, 0, time.Local),
			want:      true},
		{name: "from > toでdayEndと一致",
			timeRange: newTimeRange(16, 30, 0, 5, 30, 5),
			arg:       time.Date(2021, 5, 4, 23, 59, 59, 999999999, time.Local),
			want:      true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.timeRange.between(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_newTimeRanges(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []*timeRange
		want *timeRanges
	}{
		{name: "引数ひとつ",
			args: []*timeRange{newTimeRange(8, 0, 0, 15, 0, 0)},
			want: &timeRanges{timeRanges: []*timeRange{newTimeRange(8, 0, 0, 15, 0, 0)}}},
		{name: "引数みっつ",
			args: []*timeRange{
				newTimeRange(9, 0, 0, 11, 30, 0),
				newTimeRange(12, 30, 0, 15, 0, 0)},
			want: &timeRanges{[]*timeRange{
				newTimeRange(9, 0, 0, 11, 30, 0),
				newTimeRange(12, 30, 0, 15, 0, 0)}}},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := newTimeRanges(test.args...)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}

func Test_timeRanges_between(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		timeRanges *timeRanges
		arg        time.Time
		want       bool
	}{
		{name: "rangesが空",
			timeRanges: &timeRanges{timeRanges: []*timeRange{}},
			arg:        time.Date(2021, 5, 4, 5, 35, 0, 0, time.Local),
			want:       false},
		{name: "rangesに要素があるけどマッチしない",
			timeRanges: &timeRanges{timeRanges: []*timeRange{
				newTimeRange(8, 45, 0, 15, 15, 0),
				newTimeRange(16, 30, 0, 5, 30, 0),
			}},
			arg:  time.Date(2021, 5, 4, 5, 35, 0, 0, time.Local),
			want: false},
		{name: "rangesに複数要素があり、マッチするものがある",
			timeRanges: &timeRanges{timeRanges: []*timeRange{
				newTimeRange(8, 45, 0, 15, 15, 0),
				newTimeRange(16, 30, 0, 5, 30, 0),
			}},
			arg:  time.Date(2021, 5, 4, 5, 0, 0, 0, time.Local),
			want: true},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.timeRanges.between(test.arg)
			if !reflect.DeepEqual(test.want, got) {
				t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), test.want, got)
			}
		})
	}
}
