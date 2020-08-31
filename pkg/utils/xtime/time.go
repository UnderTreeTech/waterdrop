package xtime

import "time"

// get current unix time
func GetCurrentUnixTime() int64 {
	return time.Now().Unix()
}
