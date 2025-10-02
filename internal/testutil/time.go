package testutil

import (
	"time"
)

// TimeFromUnix 从 Unix 时间戳创建 time.Time
func TimeFromUnix(sec int64) time.Time {
	return time.Unix(sec, 0)
}
