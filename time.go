package virtual_security

import "time"

var (
	dayStart = time.Date(0, 1, 1, 0, 0, 0, 0, time.Local)
	dayEnd   = time.Date(0, 1, 2, 0, 0, 0, 0, time.Local).Add(-1 * time.Nanosecond)

	// 約定可能時間
	stockContractTime = newTimeRanges(
		newTimeRange(9, 0, 0, 11, 30, 5),
		newTimeRange(12, 30, 0, 15, 0, 5))

	// 約定可能現値時間
	contractableStockPriceTime = newTimeRanges(
		newTimeRange(9, 0, 0, 11, 30, 0),
		newTimeRange(12, 30, 0, 15, 0, 0))

	// 約定可能な前場時間
	contractableMorningSessionTime = newTimeRanges(
		newTimeRange(9, 0, 0, 11, 30, 5))

	// 約定可能な前場引け時間
	contractableMorningSessionCloseTime = newTimeRanges(
		newTimeRange(11, 30, 0, 11, 30, 5))

	// 約定可能な後場時間
	contractableAfternoonSessionTime = newTimeRanges(
		newTimeRange(12, 30, 0, 15, 0, 5))

	// 約定可能な後場引け時間
	contractableAfternoonSessionCloseTime = newTimeRanges(
		newTimeRange(15, 0, 0, 15, 0, 5))
)

func newTimeRanges(ranges ...*timeRange) *timeRanges {
	return &timeRanges{timeRanges: ranges}
}

type timeRanges struct {
	timeRanges []*timeRange
}

func (t *timeRanges) between(target time.Time) bool {
	for _, tr := range t.timeRanges {
		if tr.between(target) {
			return true
		}
	}
	return false
}

func newTimeRange(fromHour, fromMinute, fromSecond, toHour, toMinute, toSecond int) *timeRange {
	from := time.Date(0, 1, 1, fromHour, fromMinute, fromSecond, 0, time.Local)
	to := time.Date(0, 1, 1, toHour, toMinute, toSecond, 0, time.Local)

	return &timeRange{
		from:   from,
		to:     to,
		isBack: !from.Before(to),
	}
}

type timeRange struct {
	from   time.Time
	to     time.Time
	isBack bool
}

func (t *timeRange) between(target time.Time) bool {
	targetTime := time.Date(0, 1, 1, target.Hour(), target.Minute(), target.Second(), target.Nanosecond(), target.Location())
	if t.isBack {
		// 00:00:00 <= target < to
		// from <= target <= 23:59:59
		return (!targetTime.Before(dayStart) && targetTime.Before(t.to)) || (!targetTime.Before(t.from) && !targetTime.After(dayEnd))
	} else {
		// from <= target < to
		return !targetTime.Before(t.from) && targetTime.Before(t.to)
	}
}
