package gccontrol

import (
	"testing"

	"github.com/matryer/is"
)

func TestSampler_Update(t *testing.T) {
	is := is.New(t)
	s := newSampler(3)

	// Checking if it is based on the minimum.
	s.update(30)
	s.update(35)
	s.update(37)
	is.Equal(int64(30), s.get())

	// Checking bounds.
	s.update(maxSampleRate)
	s.update(maxSampleRate + 100)
	s.update(maxSampleRate + 1000)
	is.Equal(int64(maxSampleRate), s.get())

	// When zero, curr must not be updated.
	s.update(0)
	s.update(0)
	s.update(0)
	is.Equal(int64(maxSampleRate), s.get())
}
