package pynezzentials

import (
	"strconv"
	"time"
)

// UnixNanoTimestamp returns the current Unix timestamp in nanoseconds.
func UnixNanoTimestamp() int64 {
	return time.Now().UnixNano()
}

// UnixTimestamp returns the current Unix timestamp in seconds.
func UnixTimestamp() int64 {
	return time.Now().Unix()
}

// UnixTimeToTime converts a Unix timestamp to a time.Time object.
func UnixTimeToTime(unixTimestamp int64) time.Time {
	return time.Unix(unixTimestamp, 0)
}

// UnixNanoToTime converts a Unix timestamp in nanoseconds to a time.Time object.
func UnixNanoToTime(unixNanoTimestamp int64) time.Time {
	return time.Unix(0, unixNanoTimestamp)
}

// TimeToUnixTimestamp converts a time.Time object to a Unix timestamp.
func TimeToUnixTimestamp(t time.Time) int64 {
	return t.Unix()
}

// UnixMilliTimestamp returns the current Unix timestamp in milliseconds.
func UnixMilliTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// UserSessionIdValue generates a session ID for a user based on the user's ID and the current time.
func UserSessionIdValue(userId uint64, timestamp time.Time) string {
	return strconv.FormatUint(userId, 10) + "-" + strconv.FormatInt(timestamp.UnixNano(), 10)[:5]
}

func TimestampToTime(timestamp string) time.Time {
	i, _ := strconv.ParseInt(timestamp, 10, 64)
	return time.Unix(i, 0)
}
