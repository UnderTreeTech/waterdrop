/*
 *
 * Copyright 2020 waterdrop authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package xtime

import (
	"fmt"
	"time"
)

const (
	RFC3339Format       = time.RFC3339
	DateFormat          = "2006-01-02"
	DateTimeFormat      = "2006-01-02 15:04:05"
	TimeFormat          = "15:04:05"
	ShortDateTimeFormat = "20060102150405"
	ShortDateFormat     = "20060102"
	ShortTimeFormat     = "150405"

	DaysPerLeapYear       = 366   // 每闰年366天
	DaysPerNormalYear     = 365   // 每常规年365天
	MonthsPerYear         = 12    // 每年12月
	MonthsPerQuarter      = 3     // 每季度3月
	DaysPerNormalMonth    = 30    // 单月30天
	DaysPerDoubleMonth    = 31    // 双月31天
	DaysOfNormalFebruary  = 28    // 二月28天
	DaysOfLeapFebruary    = 29    // 闰年二月29天
	DaysPerWeek           = 7     // 每周7天
	HoursPerDay           = 24    // 每天24小时
	MinutesPerDay         = 1440  // 每天1440分钟
	MinutesPerHour        = 60    // 每小时60分钟
	SecondsPerDay         = 86400 // 每天86400秒
	SecondsPerHour        = 3600  // 每小时3600秒
	SecondsPerMinute      = 60    // 每分钟60秒
	MillisecondsPerSecond = 1000  // 每秒1000毫秒
)

type Time struct {
	time.Time
}

// Now return current locale time
func Now() *Time {
	return &Time{
		Time: time.Now(),
	}
}

// GetCurrentUnixTime return current unix seconds
func (t *Time) CurrentUnixTime() int64 {
	return t.Time.Unix()
}

// GetCurrentMilliTime return current milliseconds
func (t *Time) CurrentMilliTime() int64 {
	return t.Time.UnixNano() / 1e6
}

// GetCurrentNanoTime return current nano seconds
func (t *Time) CurrentNanoTime() int64 {
	return t.Time.UnixNano()
}

// Leap check current time is leap year or not
func (t *Time) Leap() bool {
	return IsLeap(t.Year())
}

// Yesterday return a yesterday Time instance of assigned t
func (t *Time) Yesterday() *Time {
	return &Time{t.Time.AddDate(0, 0, -1)}
}

// Tomorrow return a tomorrow Time instance of assigned t
func (t *Time) Tomorrow() *Time {
	return &Time{t.Time.AddDate(0, 0, 1)}
}

// DaysBefore returns a days-ago Time instance of assigned t
func (t *Time) DaysBefore(days int) *Time {
	return &Time{t.AddDate(0, 0, -days)}
}

// DaysAfter returns a days-after Time instance of assigned t
func (t *Time) DaysAfter(days int) *Time {
	return &Time{t.AddDate(0, 0, days)}
}

// Format returns a textual representation of the time value formatted according to layout
func (t *Time) Format(layout string) string {
	return t.Time.Format(layout)
}

// BeginOfYear return the beginning time of current year
func (t *Time) BeginOfYear() *Time {
	y, _, _ := t.Date()
	return &Time{time.Date(y, time.January, 1, 0, 0, 0, 0, t.Location())}
}

// EndOfYear return the end time of current year
func (t *Time) EndOfYear() *Time {
	return &Time{t.BeginOfYear().AddDate(1, 0, 0).Add(-time.Nanosecond)}
}

// BeginOfMonth return begin day time of current month
func (t *Time) BeginOfMonth() *Time {
	y, m, _ := t.Date()
	return &Time{time.Date(y, m, 1, 0, 0, 0, 0, t.Location())}
}

// EndOfMonth return end day time of current month
func (t *Time) EndOfMonth() *Time {
	return &Time{t.BeginOfMonth().AddDate(0, 1, 0).Add(-time.Nanosecond)}
}

// BeginOfWeek return begin day time of current week
// NOTE: week begin from Sunday
func (t *Time) BeginOfWeek() *Time {
	y, m, d := t.AddDate(0, 0, 0-int(t.BeginOfDay().Weekday())).Date()
	return &Time{time.Date(y, m, d, 0, 0, 0, 0, t.Location())}
}

// EndOfWeek return end day time of current week
// NOTE: week end with Saturday
func (t *Time) EndOfWeek() *Time {
	y, m, d := t.BeginOfWeek().AddDate(0, 0, 7).Add(-time.Nanosecond).Date()
	return &Time{time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())}
}

// BeginOfDay return begin time of current day
func (t *Time) BeginOfDay() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, 0, 0, 0, 0, t.Location())}
}

// EndOfDay return end time of current day
func (t *Time) EndOfDay() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())}
}

// BeginOfHour return begin time of current hour
func (t *Time) BeginOfHour() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, t.Hour(), 0, 0, 0, t.Location())}
}

// EndOfHour return end time of current hour
func (t *Time) EndOfHour() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, t.Hour(), 59, 59, int(time.Second-time.Nanosecond), t.Location())}
}

// BeginOfMinute return begin second of current minute
func (t *Time) BeginOfMinute() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, t.Hour(), t.Minute(), 0, 0, t.Location())}
}

// EndOfMinute return end second of current minute
func (t *Time) EndOfMinute() *Time {
	y, m, d := t.Date()
	return &Time{time.Date(y, m, d, t.Hour(), t.Minute(), 59, int(time.Second-time.Nanosecond), t.Location())}
}

// FormatUnix formats Unix timestamp to layout string
func FormatUnix(timestamp int64, layout string) string {
	return time.Unix(timestamp, 0).Format(layout)
}

// FormatUnixDate formats Unix timestamp to date string
func FormatUnixDate(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(DateFormat)
}

// FormatUnixDateTime formats Unix timestamp to time string
func FormatUnixDateTime(timestamp int64) string {
	return time.Unix(timestamp, 0).Format(DateTimeFormat)
}

// IsLeap check a leap year
func IsLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// Yesterday returns a yesterday Time instance of current time
func Yesterday() *Time {
	return &Time{Now().AddDate(0, 0, -1)}
}

// Tomorrow returns a tomorrow Time instance of current time
func Tomorrow() *Time {
	return &Time{Now().AddDate(0, 0, 1)}
}

// DaysBefore returns a days-ago Time instance of current time
func DaysBefore(days int) *Time {
	return &Time{Now().AddDate(0, 0, -days)}
}

// DaysAfter returns a days-after Time instance of current time
func DaysAfter(days int) *Time {
	return &Time{Now().AddDate(0, 0, days)}
}

// ParseByLayout parses a string as a Time instance by layout
func ParseByLayout(t string, layout string, loc ...*time.Location) (*Time, error) {
	l := time.Local
	if len(loc) > 0 {
		l = loc[0]
	}
	pt, err := time.ParseInLocation(layout, t, l)
	if err != nil {
		return nil, fmt.Errorf("cannot parse %s as Time by layout %s", t, layout)
	}
	return &Time{pt}, nil
}
