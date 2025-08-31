package models

import "time"

// TimeNow is a thin wrapper around time.Now to make tests deterministic via
// replacement if needed in the future.
func TimeNow() time.Time {
	return time.Now()
}
