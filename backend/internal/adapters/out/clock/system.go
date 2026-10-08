// Package clock provides the wall clock behind ports.Clock.
package clock

import "time"

// System reads the real time.
type System struct{}

// Now returns the current UTC time.
func (System) Now() time.Time { return time.Now().UTC() }
