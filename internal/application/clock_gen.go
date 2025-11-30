package application

import "time"

type ClockGen func() time.Time

func TimeNow() time.Time {
	return time.Now()
}
