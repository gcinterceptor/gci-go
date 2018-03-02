package gccontrol

import (
	"testing"

	"github.com/matryer/is"
)

func TestNewST(t *testing.T) {
	is := is.New(t)
	st := newST()
	is.Equal(0.0, st.numGCs)
	is.True(st.val >= minST && st.val <= 2*minST)
}

func TestST_Update(t *testing.T) {
	t.Run("CriticalPath", func(t *testing.T) {
		is := is.New(t)
		st := &st{}
		alloc := uint64(256 * 1025 * 1024)
		st.Update(alloc, 1000, 50) // It is fine for the first run (see startMaxOverhead).
		is.True(st.Get() > alloc)
		for i := 0; i < maxGCs; i++ {
			st.Update(alloc, 1000, 50)
		}
		st.Update(alloc, 1000, 50)
		is.True(st.Get() < alloc)   // After some runs, the same amount of overhead is not fine anymore.
		is.Equal(maxGCs, st.numGCs) // numGCs should never exceed maxGCs
	})
	t.Run("Bounds", func(t *testing.T) {
		is := is.New(t)
		st := &st{}
		st.Update(10, 1000, 1000)
		is.True(st.Get() > minST)

		st.Update(maxST*2, 1000, 1000)
		is.True(st.Get() < maxST)
	})
}
