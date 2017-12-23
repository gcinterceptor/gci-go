package gccontrol

import (
	"testing"

	"github.com/matryer/is"
)

func TestSampler_Update(t *testing.T) {
	is := is.New(t)
	s := newSampler(3)
	s.update(10)
	is.Equal(0, s.next) // Next shouldn't be updated at the first update.

	// Checking if it is based on the minimum.
	s.update(30)
	s.update(35)
	s.update(37)
	is.Equal(int64(2), s.get())

	// Checking bounds.
	s.update(100)
	s.update(200)
	s.update(300)
	is.Equal(int64(maxSampleRate), s.get())
}
