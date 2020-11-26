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
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var now = &Time{
	Time: time.Date(2020, 11, 26, 14, 10, 32, 331, time.Local),
}

func TestMain(m *testing.M) {
	cstSh, _ := time.LoadLocation("Asia/Shanghai")
	t := now.In(cstSh)
	now.Time = t
	os.Exit(m.Run())
}

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
	millis := now.CurrentMilliTime()
	nano := now.CurrentNanoTime()
	assert.Regexp(t, dateReg, FormatUnixDate(unix))
	assert.Regexp(t, dateReg, FormatMilliDate(millis))
	assert.Regexp(t, datetimeReg, FormatUnixDateTime(unix))
	assert.Regexp(t, datetimeReg, FormatMilliDateTime(millis))
	assert.Regexp(t, dateReg, FormatMilliDate(nano/1e6))
	assert.Regexp(t, datetimeReg, FormatMilliDateTime(nano/1e6))
}

func TestTime(t *testing.T) {
	beginOfYear := int64(1577808000)
	endOfYear := beginOfYear + 365*86400 - 1
	if now.Leap() {
		endOfYear += 86400
	}
	beginOfMonth := int64(1604160000)
	endOfMonth := beginOfMonth + 30*86400 - 1
	beginOfWeek := int64(1605974400)
	endOfWeek := beginOfWeek + 7*86400 - 1
	beginOfDay := int64(1606320000)
	endOfDay := beginOfDay + 86400 - 1
	beginOfHour := int64(1606370400)
	endOfHour := beginOfHour + 60*60 - 1
	beginOfMinute := int64(1606371000)
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
}
