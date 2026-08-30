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
	"fmt"
	"time"
)

const (
	// Common time formats
	FormatDateTime     = "2006-01-02 15:04:05"
	FormatDate         = "2006-01-02"
	FormatTime         = "15:04:05"
	FormatDateTimeNano = "2006-01-02 15:04:05.999999999"
	FormatMonth        = "2006-01"
	FormatYear         = "2006"
)

// Now returns the current time.
func Now() time.Time {
	return time.Now()
}

// NowUnix returns the current Unix timestamp in seconds.
func NowUnix() int64 {
	return time.Now().Unix()
}

// NowUnixMilli returns the current Unix timestamp in milliseconds.
func NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// NowUnixMicro returns the current Unix timestamp in microseconds.
func NowUnixMicro() int64 {
	return time.Now().UnixMicro()
}

// Today returns the current date at midnight.
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow returns tomorrow's date at midnight.
func Tomorrow() time.Time {
	return Today().Add(24 * time.Hour)
}

// Yesterday returns yesterday's date at midnight.
func Yesterday() time.Time {
	return Today().Add(-24 * time.Hour)
}

// Format formats the time using the specified layout.
func Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// FormatNow formats the current time using the specified layout.
func FormatNow(layout string) string {
	return time.Now().Format(layout)
}

// Parse parses a time string using the specified layout.
func Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// ParseInLocation parses a time string in the specified location.
func ParseInLocation(layout, value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(layout, value, loc)
}

// MustParse parses a time string and panics on error.
func MustParse(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

// UnixToDate converts Unix timestamp to date string.
func UnixToDate(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(FormatDate)
}

// UnixToDateTime converts Unix timestamp to datetime string.
func UnixToDateTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(FormatDateTime)
}

// DateToUnix converts date string to Unix timestamp.
func DateToUnix(date string) (int64, error) {
	t, err := time.Parse(FormatDate, date)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// DateTimeToUnix converts datetime string to Unix timestamp.
func DateTimeToUnix(datetime string) (int64, error) {
	t, err := time.Parse(FormatDateTime, datetime)
	if err != nil {
		return 0, err
	}
	return t.Unix(), nil
}

// Duration returns the duration between two times.
func Duration(start, end time.Time) time.Duration {
	return end.Sub(start)
}

// Age calculates the age based on birthday.
func Age(birthday time.Time) int {
	now := time.Now()
	years := now.Year() - birthday.Year()
	if now.YearDay() < birthday.YearDay() {
		years--
	}
	return years
}

// IsLeapYear checks if the given year is a leap year.
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in the specified month and year.
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// StartOfDay returns the start of the day (00:00:00).
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day (23:59:59.999999999).
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week (Monday).
func StartOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return StartOfDay(t.AddDate(0, 0, -int(weekday-time.Monday)))
}

// EndOfWeek returns the end of the week (Sunday 23:59:59.999999999).
func EndOfWeek(t time.Time) time.Time {
	return EndOfDay(StartOfWeek(t).AddDate(0, 0, 6))
}

// StartOfMonth returns the start of the month.
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month.
func EndOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month()+1, 0, 23, 59, 59, 999999999, t.Location())
}

// StartOfYear returns the start of the year.
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear returns the end of the year.
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// IsToday checks if the given time is today.
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsTomorrow checks if the given time is tomorrow.
func IsTomorrow(t *time.Time) bool {
	tomorrow := Tomorrow()
	return t.Year() == tomorrow.Year() && t.Month() == tomorrow.Month() && t.Day() == tomorrow.Day()
}

// IsYesterday checks if the given time is yesterday.
func IsYesterday(t time.Time) bool {
	yesterday := Yesterday()
	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}

// Between checks if the time is between start and end (inclusive).
func Between(t, start, end time.Time) bool {
	return !t.Before(start) && !t.After(end)
}

// Elapsed returns a human-readable string of the elapsed time.
func Elapsed(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%d seconds ago", int(d.Seconds()))
	case d < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}

// Sleep sleeps for the specified duration.
func Sleep(d time.Duration) {
	time.Sleep(d)
}

// SleepSeconds sleeps for the specified number of seconds.
func SleepSeconds(seconds float64) {
	time.Sleep(time.Duration(seconds * float64(time.Second)))
}
