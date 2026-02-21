package clock

import "time"

// Clock abstracts time to enable deterministic testing.
type Clock interface {
	Now() time.Time
}

// Real returns a Clock backed by the system clock.
func Real() Clock { return realClock{} }

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

// Fixed returns a Clock that always returns t.
// Use in tests to control time.
func Fixed(t time.Time) Clock { return fixedClock{t: t} }

type fixedClock struct{ t time.Time }

func (f fixedClock) Now() time.Time { return f.t }
