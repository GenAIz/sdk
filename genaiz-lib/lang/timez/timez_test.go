package timez

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDualFormat_FormatMillis_End(t *testing.T) {
	var testEnd = time.Now().Add(2 * time.Hour)
	var testBefore = testEnd.Add(-1 * time.Minute)
	var testAfter = time.Now().Add(3 * time.Hour)
	var testFormatter = &dualFormat{
		formatLong: func(t time.Time) string {
			return t.Format(time.DateOnly)
		},
		formatShort: func(t time.Time) string {
			return t.Format(time.TimeOnly)
		},

		end: &testEnd,
	}

	assert.Equal(t, testEnd.Format(time.TimeOnly), testFormatter.FormatMillis(testEnd.UnixMilli()))
	assert.Equal(t, testBefore.Format(time.TimeOnly), testFormatter.FormatMillis(testBefore.UnixMilli()))
	assert.Equal(t, testAfter.Format(time.DateOnly), testFormatter.FormatMillis(testAfter.UnixMilli()))
}

func TestNewMidnightFormatter(t *testing.T) {
	var testFormatter = NewMidnightFormatter()
	var testLong = time.Now().Add(48 * time.Hour)
	var testShort = time.Now().Add(-1 * time.Hour)

	assert.Equal(t, testLong.Format(time.DateOnly), testFormatter.FormatMillis(testLong.UnixMilli()))
	assert.Equal(t, testShort.Format(time.TimeOnly), testFormatter.FormatSeconds(testShort.UnixMilli()/1000))
}

func TestNewTodayFormatter(t *testing.T) {
	var testFormatter = NewTodayFormatter()
	var testBefore = time.Now().Add(-48 * time.Hour)
	var testAfter = time.Now().Add(24 * time.Hour)

	assert.Equal(t, testBefore.Format(time.DateOnly), testFormatter.FormatMillis(testBefore.UnixMilli()))
	assert.Equal(t, testAfter.Format(time.TimeOnly), testFormatter.FormatSeconds(testAfter.UnixMilli()/1000))
}

func TestNewFormatter(t *testing.T) {
	var startTime = time.Now()
	var endTime = time.Now().Add(2 * time.Hour)
	var testFormatter = NewFormatter(&startTime, &endTime)
	var testTime = time.Now().Add(1 * time.Hour)

	assert.Equal(t, "1969-12-31", testFormatter.FormatSeconds(1))
	assert.Equal(t, testTime.Format(time.TimeOnly), testFormatter.FormatMillis(testTime.UnixMilli()))
}

func TestNewMonthTime(t *testing.T) {
	var testTime = NewMonthTime()

	assert.True(t, testTime.Before(time.Now()))
}

func TestNewTodayTime(t *testing.T) {
	var testTime = NewTodayTime()

	assert.True(t, testTime.Before(time.Now()))
}

func TestNewWeekTime(t *testing.T) {
	var testTime = NewWeekTime()

	assert.True(t, testTime.Before(time.Now()))
}
