package timez

import "time"

type Formatter interface {
	FormatSeconds(int64) string

	FormatMillis(int64) string
}

type dualFormat struct {
	formatLong  func(time.Time) string
	formatShort func(time.Time) string

	end   *time.Time
	start *time.Time
}

func (df dualFormat) FormatSeconds(epochSeconds int64) string {
	return df.FormatMillis(epochSeconds * 1000)
}

func (df dualFormat) FormatMillis(epochMillis int64) string {
	var epochTime = time.UnixMilli(epochMillis)

	if df.start != nil {
		if epochTime.After(*df.start) {
			return df.formatShort(epochTime)
		}
	} else if df.end != nil {
		if epochTime.Before(*df.end) {
			return df.formatShort(epochTime)
		}
	}

	return df.formatLong(epochTime)
}

func NewMidnightFormatter() Formatter {
	var year, month, day = time.Now().In(time.Local).Date()
	var midnight = time.Date(year, month, day, 0, 0, 0, 0, time.Local).
		Add(24 * time.Hour)

	return NewFormatter(nil, &midnight)
}

func NewTodayFormatter() Formatter {
	var year, month, day = time.Now().In(time.Local).Date()
	var today = time.Date(year, month, day, 0, 0, 0, 0, time.Local)

	return NewFormatter(&today, nil)
}

func NewFormatter(start, end *time.Time) Formatter {
	return &dualFormat{
		formatLong: func(t time.Time) string {
			return t.Format(time.DateOnly)
		},
		formatShort: func(t time.Time) string {
			return t.Format(time.TimeOnly)
		},
		end:   end,
		start: start,
	}
}

func NewMonthTime() *time.Time {
	var now = time.Now()
	var result = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	return &result
}

func NewTodayTime() *time.Time {
	var now = time.Now()
	var result = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return &result
}

func NewWeekTime() *time.Time {
	var now = time.Now()
	var weekDay = now.Weekday()
	var beginning = now.AddDate(0, 0, int(time.Sunday-weekDay))
	var result = time.Date(beginning.Year(), beginning.Month(), beginning.Day(), 0, 0, 0, 0, beginning.Location())

	return &result
}
