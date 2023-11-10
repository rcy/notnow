package yikes

import "time"

func (t *Task) Duration() time.Duration {
	return 15 * time.Minute
}
