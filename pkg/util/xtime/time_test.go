// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xtime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNow(t *testing.T) {
	now := Now()
	assert.False(t, now.IsZero())
}

func TestNowUnix(t *testing.T) {
	unix := NowUnix()
	assert.Greater(t, unix, int64(0))
}

func TestToday(t *testing.T) {
	today := Today()
	assert.Equal(t, 0, today.Hour())
	assert.Equal(t, 0, today.Minute())
	assert.Equal(t, 0, today.Second())
}

func TestTomorrow(t *testing.T) {
	today := Today()
	tomorrow := Tomorrow()
	assert.Equal(t, today.Add(24*time.Hour), tomorrow)
}

func TestYesterday(t *testing.T) {
	today := Today()
	yesterday := Yesterday()
	assert.Equal(t, today.Add(-24*time.Hour), yesterday)
}

func TestFormat(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	assert.Equal(t, "2024-01-15 14:30:45", Format(t1, FormatDateTime))
	assert.Equal(t, "2024-01-15", Format(t1, FormatDate))
	assert.Equal(t, "14:30:45", Format(t1, FormatTime))
}

func TestParse(t *testing.T) {
	t1, err := Parse(FormatDateTime, "2024-01-15 14:30:45")
	assert.NoError(t, err)
	assert.Equal(t, 2024, t1.Year())
	assert.Equal(t, time.January, t1.Month())
	assert.Equal(t, 15, t1.Day())
}

func TestMustParse(t *testing.T) {
	t1 := MustParse(FormatDate, "2024-01-15")
	assert.Equal(t, 2024, t1.Year())
}

func TestMustParsePanic(t *testing.T) {
	assert.Panics(t, func() {
		MustParse(FormatDate, "invalid")
	})
}

func TestUnixToDate(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	result := UnixToDate(t1.Unix())
	assert.Equal(t, "2024-01-15", result)
}

func TestUnixToDateTime(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 14, 30, 45, 0, time.Local)
	result := UnixToDateTime(t1.Unix())
	assert.Equal(t, "2024-01-15 14:30:45", result)
}

func TestDateToUnix(t *testing.T) {
	unix, err := DateToUnix("2024-01-15")
	assert.NoError(t, err)
	assert.Greater(t, unix, int64(0))
}

func TestDateTimeToUnix(t *testing.T) {
	unix, err := DateTimeToUnix("2024-01-15 14:30:45")
	assert.NoError(t, err)
	assert.Greater(t, unix, int64(0))
}

func TestDuration(t *testing.T) {
	start := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 15, 30, 0, 0, time.UTC)
	d := Duration(start, end)
	assert.Equal(t, 90*time.Minute, d)
}

func TestAge(t *testing.T) {
	now := time.Now()
	birthday := now.AddDate(-30, 0, 0)
	age := Age(birthday)
	// Age should be 30 if birthday has passed this year, or 29 if not
	if now.YearDay() >= birthday.YearDay() {
		assert.Equal(t, 30, age)
	} else {
		assert.Equal(t, 29, age)
	}
}

func TestIsLeapYear(t *testing.T) {
	assert.True(t, IsLeapYear(2024))
	assert.False(t, IsLeapYear(2023))
	assert.False(t, IsLeapYear(1900))
	assert.True(t, IsLeapYear(2000))
}

func TestDaysInMonth(t *testing.T) {
	assert.Equal(t, 31, DaysInMonth(2024, time.January))
	assert.Equal(t, 29, DaysInMonth(2024, time.February)) // Leap year
	assert.Equal(t, 28, DaysInMonth(2023, time.February)) // Non-leap year
	assert.Equal(t, 30, DaysInMonth(2024, time.April))
}

func TestStartOfDay(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	start := StartOfDay(t1)
	assert.Equal(t, 0, start.Hour())
	assert.Equal(t, 0, start.Minute())
	assert.Equal(t, 0, start.Second())
}

func TestEndOfDay(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	end := EndOfDay(t1)
	assert.Equal(t, 23, end.Hour())
	assert.Equal(t, 59, end.Minute())
	assert.Equal(t, 59, end.Second())
}

func TestStartOfWeek(t *testing.T) {
	// 2024-01-15 is Monday
	t1 := time.Date(2024, 1, 17, 14, 30, 45, 0, time.UTC) // Wednesday
	start := StartOfWeek(t1)
	assert.Equal(t, time.Monday, start.Weekday())
}

func TestEndOfWeek(t *testing.T) {
	// 2024-01-15 is Monday
	t1 := time.Date(2024, 1, 17, 14, 30, 45, 0, time.UTC) // Wednesday
	end := EndOfWeek(t1)
	assert.Equal(t, time.Sunday, end.Weekday())
}

func TestStartOfMonth(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	start := StartOfMonth(t1)
	assert.Equal(t, 1, start.Day())
}

func TestEndOfMonth(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 14, 30, 45, 0, time.UTC)
	end := EndOfMonth(t1)
	assert.Equal(t, 31, end.Day())
}

func TestStartOfYear(t *testing.T) {
	t1 := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)
	start := StartOfYear(t1)
	assert.Equal(t, time.January, start.Month())
	assert.Equal(t, 1, start.Day())
}

func TestEndOfYear(t *testing.T) {
	t1 := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)
	end := EndOfYear(t1)
	assert.Equal(t, time.December, end.Month())
	assert.Equal(t, 31, end.Day())
}

func TestIsToday(t *testing.T) {
	now := time.Now()
	assert.True(t, IsToday(now))

	yesterday := time.Now().AddDate(0, 0, -1)
	assert.False(t, IsToday(yesterday))
}

func TestBetween(t *testing.T) {
	t1 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	start := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 15, 23, 59, 59, 0, time.UTC)

	assert.True(t, Between(t1, start, end))
	assert.False(t, Between(start.AddDate(0, 0, -1), start, end))
}

func TestElapsed(t *testing.T) {
	t1 := time.Now().Add(-30 * time.Second)
	elapsed := Elapsed(t1)
	assert.Contains(t, elapsed, "seconds ago")

	t2 := time.Now().Add(-5 * time.Minute)
	elapsed = Elapsed(t2)
	assert.Contains(t, elapsed, "minutes ago")

	t3 := time.Now().Add(-2 * time.Hour)
	elapsed = Elapsed(t3)
	assert.Contains(t, elapsed, "hours ago")
}
