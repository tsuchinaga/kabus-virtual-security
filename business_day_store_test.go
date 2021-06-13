package virtual_security

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	jbd "gitlab.com/tsuchinaga/jpx-business-day"
)

type testJPXBusinessDay struct {
	jbd.BusinessDay
	isBusinessDay  bool
	refresh        error
	refreshCount   int
	lastUpdateDate time.Time
}

func (t *testJPXBusinessDay) IsBusinessDay(time.Time) bool { return t.isBusinessDay }
func (t *testJPXBusinessDay) Refresh(context.Context) error {
	t.refreshCount++
	return t.refresh
}
func (t *testJPXBusinessDay) LastUpdateDate() time.Time { return t.lastUpdateDate }

func Test_getBusinessDayStore(t *testing.T) {
	t.Parallel()

	clock := &testClock{}
	got := getBusinessDayStore(clock)
	want := &businessDayStore{
		clock:       clock,
		businessDay: jbd.NewBusinessDay(),
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("%s error\nwant: %+v\ngot: %+v\n", t.Name(), want, got)
	}
}

func Test_businessDayStore_IsBusinessDay(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		businessDayStore *businessDayStore
		now              time.Time
		lastUpdateDate   time.Time
		refresh          error
		isBusinessDay    bool
		arg              time.Time
		want             bool
		hasError         bool
		refreshCount     int
	}{
		{name: "refreshを叩いたことがなければrefreshを叩いてからisBusinessDayの結果を返す",
			businessDayStore: &businessDayStore{},
			now:              time.Date(2021, 5, 7, 16, 6, 0, 0, time.Local),
			lastUpdateDate:   time.Date(2021, 1, 7, 0, 0, 0, 0, time.Local),
			isBusinessDay:    true,
			want:             true,
			hasError:         false,
			refreshCount:     1},
		{name: "refreshが前日以前ならrefreshを叩いてからisBusinessDayの結果を返す",
			businessDayStore: &businessDayStore{refreshedAt: time.Date(2021, 5, 6, 9, 0, 0, 0, time.Local)},
			now:              time.Date(2021, 5, 7, 16, 6, 0, 0, time.Local),
			lastUpdateDate:   time.Date(2021, 1, 7, 0, 0, 0, 0, time.Local),
			isBusinessDay:    false,
			want:             false,
			hasError:         false,
			refreshCount:     1},
		{name: "jpxのlastUpdateDateがゼロ値ならrefreshを叩いてからisBusinessDayの結果を返す",
			businessDayStore: &businessDayStore{refreshedAt: time.Date(2021, 5, 7, 9, 0, 0, 0, time.Local)},
			now:              time.Date(2021, 5, 7, 16, 6, 0, 0, time.Local),
			lastUpdateDate:   time.Time{},
			isBusinessDay:    true,
			want:             true,
			hasError:         false,
			refreshCount:     1},
		{name: "refreshに失敗したらエラーを返す",
			businessDayStore: &businessDayStore{refreshedAt: time.Date(2021, 5, 7, 9, 0, 0, 0, time.Local)},
			now:              time.Date(2021, 5, 7, 16, 6, 0, 0, time.Local),
			lastUpdateDate:   time.Time{},
			refresh:          errors.New("error message"),
			isBusinessDay:    true,
			want:             false,
			hasError:         true,
			refreshCount:     1},
		{name: "当日内にrefreshされていればrefreshしない",
			businessDayStore: &businessDayStore{refreshedAt: time.Date(2021, 5, 7, 9, 0, 0, 0, time.Local)},
			now:              time.Date(2021, 5, 7, 16, 6, 0, 0, time.Local),
			lastUpdateDate:   time.Date(2021, 1, 7, 0, 0, 0, 0, time.Local),
			isBusinessDay:    true,
			want:             true,
			hasError:         false,
			refreshCount:     0},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			test.businessDayStore.clock = &testClock{now: test.now}
			businessDay := &testJPXBusinessDay{lastUpdateDate: test.lastUpdateDate, isBusinessDay: test.isBusinessDay, refresh: test.refresh}
			test.businessDayStore.businessDay = businessDay
			got, err := test.businessDayStore.IsBusinessDay(test.arg)
			if !reflect.DeepEqual(test.want, got) || (err != nil) != test.hasError || test.refreshCount != businessDay.refreshCount {
				t.Errorf("%s error\nwant: %+v, %+v, %+v\ngot: %+v, %+v, %+v\n", t.Name(), test.want, test.hasError, test.refreshCount, got, err, businessDay.refreshCount)
			}
		})
	}
}
