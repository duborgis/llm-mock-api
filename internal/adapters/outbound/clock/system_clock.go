package clock

import "time"

// SystemClock implements ports.Clock with the real wall clock.
type SystemClock struct{}

func New() SystemClock { return SystemClock{} }

func (SystemClock) Now() time.Time { return time.Now() }

func (SystemClock) Sleep(d time.Duration) { time.Sleep(d) }
