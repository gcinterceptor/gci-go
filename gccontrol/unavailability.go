package gccontrol

import (
	"time"
)

type unavailability struct {
}

func (u *unavailability) begin() {
}

func (u *unavailability) end() {

}

func (u *unavailability) estimate(queueSize int64) time.Duration {
	return 1 * time.Millisecond
}
