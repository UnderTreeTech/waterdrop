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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now = &Time{Time: time.Date(2020, 11, 26, 14, 10, 32, 331, time.UTC)}

func TestNow(t *testing.T) {
	n := Now()
	assert.IsType(t, n.Time, time.Time{})
}

func TestLeap(t *testing.T) {
	leap := now.Leap()
	assert.Equal(t, true, leap)
	assert.Equal(t, true, IsLeap(now.Year()))
}

func TestFormat(t *testing.T) {
	date := now.Format(DateFormat)
	datetime := now.Format(DateTimeFormat)
	dateReg := `^((([0-9]{3}[1-9]|[0-9]{2}[1-9][0-9]{1}|[0-9]{1}[1-9][0-9]{2}|[1-9][0-9]{3})-(((0[13578]|1[02])-(0[1-9]|[12][0-9]|3[01]))|((0[469]|11)-(0[1-9]|[12][0-9]|30))|(02-(0[1-9]|[1][0-9]|2[0-8]))))|((([0-9]{2})(0[48]|[2468][048]|[13579][26])|((0[48]|[2468][048]|[3579][26])00))-02-29))$`
	datetimeReg := `^([0-9]{3}[1-9]|[0-9]{2}[1-9][0-9]{1}|[0-9]{1}[1-9][0-9]{2}|[1-9][0-9]{3})-(((0[13578]|1[02])-(0[1-9]|[12][0-9]|3[01]))|((0[469]|11)-(0[1-9]|[12][0-9]|30))|(02-(0[1-9]|[1][0-9]|2[0-8])))\s([0-1][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9])$`

	assert.Regexp(t, dateReg, date)
	assert.Regexp(t, datetimeReg, datetime)

	unix := now.CurrentUnixTime()
	assert.Regexp(t, dateReg, FormatUnixDate(unix))
	assert.Regexp(t, datetimeReg, FormatUnixDateTime(unix))
	assert.Equal(t, FormatUnix(unix, DateFormat), now.In(time.Local).Format(DateFormat))
	assert.Equal(t, FormatUnix(unix, DateTimeFormat), now.In(time.Local).Format(DateTimeFormat))
	assert.Equal(t, FormatUnix(unix, TimeFormat), now.In(time.Local).Format(TimeFormat))

	assert.Equal(t, now.Format(TimeFormat), "14:10:32")
	assert.Equal(t, now.Format(ShortDateFormat), "20201126")
	assert.Equal(t, now.Format(ShortDateTimeFormat), "20201126141032")
	assert.Equal(t, now.Format(ShortTimeFormat), "141032")
}

func TestTime(t *testing.T) {
	beginOfYear := int64(1577836800)
	endOfYear := beginOfYear + 365*86400 - 1
	if now.Leap() {
		endOfYear += 86400
	}
	beginOfMonth := int64(1604188800)
	endOfMonth := beginOfMonth + 30*86400 - 1
	beginOfWeek := int64(1606003200)
	endOfWeek := beginOfWeek + 7*86400 - 1
	beginOfDay := int64(1606348800)
	endOfDay := beginOfDay + 86400 - 1
	beginOfHour := int64(1606399200)
	endOfHour := beginOfHour + 60*60 - 1
	beginOfMinute := int64(1606399800)
	endOfMinute := beginOfMinute + 59

	assert.Equal(t, now.BeginOfYear().CurrentUnixTime(), beginOfYear)
	assert.Equal(t, now.EndOfYear().CurrentUnixTime(), endOfYear)
	assert.Equal(t, now.BeginOfMonth().CurrentUnixTime(), beginOfMonth)
	assert.Equal(t, now.EndOfMonth().CurrentUnixTime(), endOfMonth)
	assert.Equal(t, now.BeginOfWeek().CurrentUnixTime(), beginOfWeek)
	assert.Equal(t, now.EndOfWeek().CurrentUnixTime(), endOfWeek)
	assert.Equal(t, now.BeginOfDay().CurrentUnixTime(), beginOfDay)
	assert.Equal(t, now.EndOfDay().CurrentUnixTime(), endOfDay)
	assert.Equal(t, now.BeginOfHour().CurrentUnixTime(), beginOfHour)
	assert.Equal(t, now.EndOfHour().CurrentUnixTime(), endOfHour)
	assert.Equal(t, now.BeginOfMinute().CurrentUnixTime(), beginOfMinute)
	assert.Equal(t, now.EndOfMinute().CurrentUnixTime(), endOfMinute)

	assert.Equal(t, now.Yesterday().CurrentUnixTime(), now.CurrentUnixTime()-SecondsPerDay)
	assert.Equal(t, now.Tomorrow().CurrentUnixTime(), now.CurrentUnixTime()+SecondsPerDay)
	assert.Equal(t, now.DaysBefore(5).CurrentUnixTime(), now.CurrentUnixTime()-5*SecondsPerDay)
	assert.Equal(t, now.DaysAfter(5).CurrentUnixTime(), now.CurrentUnixTime()+5*SecondsPerDay)

	assert.Equal(t, Yesterday().CurrentUnixTime(), Now().CurrentUnixTime()-SecondsPerDay)
	assert.Equal(t, Tomorrow().CurrentUnixTime(), Now().CurrentUnixTime()+SecondsPerDay)
	assert.Equal(t, DaysBefore(5).CurrentUnixTime(), Now().CurrentUnixTime()-5*SecondsPerDay)
	assert.Equal(t, DaysAfter(5).CurrentUnixTime(), Now().CurrentUnixTime()+5*SecondsPerDay)

	tt, err := ParseByLayout("2020-11-26 14:10:32", DateTimeFormat, time.UTC)
	assert.Nil(t, err)
	assert.Equal(t, tt.CurrentUnixTime(), now.CurrentUnixTime())

}
