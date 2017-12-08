package is

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
)

type mockT struct {
	failed bool
}

func (m *mockT) FailNow() {
	m.failed = true
}
func (m *mockT) Fail() {
	m.failed = true
}

var tests = []struct {
	N    string
	F    func(is *I)
	Fail string
}{
	// Equal
	{
		N: "Equal(1, 1)",
		F: func(is *I) {
			is.Equal(1, 1) // 1 doesn't equal 2
		},
		Fail: ``,
	},

	{
		N: "Equal(1, 2)",
		F: func(is *I) {
			is.Equal(1, 2) // 1 doesn't equal 2
		},
		Fail: `1 != 2 // 1 doesn't equal 2`,
	},
	{
		N: "Equal(1, nil)",
		F: func(is *I) {
			is.Equal(1, nil) // 1 doesn't equal nil
		},
		Fail: `int(1) != <nil> // 1 doesn't equal nil`,
	},
	{
		N: "Equal(nil, 2)",
		F: func(is *I) {
			is.Equal(nil, 2) // nil doesn't equal 2
		},
		Fail: `<nil> != int(2) // nil doesn't equal 2`,
	},
	{
		N: "Equal(false, false)",
		F: func(is *I) {
			is.Equal(false, false) // nil doesn't equal 2
		},
		Fail: ``,
	},
	{
		N: "Equal(int32(1), int64(1))",
		F: func(is *I) {
			is.Equal(int32(1), int64(1)) // nope
		},
		Fail: `int32(1) != int64(1) // nope`,
	},
	{
		N: "Equal(map1, map2)",
		F: func(is *I) {
			m1 := map[string]interface{}{"value": 1}
			m2 := map[string]interface{}{"value": 2}
			is.Equal(m1, m2) // maps
		},
		Fail: `map[value:1] != map[value:2] // maps`,
	},
	{
		N: "Equal(true, map2)",
		F: func(is *I) {
			m1 := map[string]interface{}{"value": 1}
			m2 := map[string]interface{}{"value": 2}
			is.Equal(m1, m2) // maps
		},
		Fail: `map[value:1] != map[value:2] // maps`,
	},
	{
		N: "Equal(slice1, slice2)",
		F: func(is *I) {
			s1 := []string{"one", "two", "three"}
			s2 := []string{"one", "two", "three", "four"}
			is.Equal(s1, s2) // slices
		},
		Fail: `[one two three] != [one two three four] // slices`,
	},
	{
		N: "Equal(nil, chan)",
		F: func(is *I) {
			var a chan string
			b := make(chan string)
			is.Equal(a, b) // channels
		},
		Fail: ` // channels`,
	},
	{
		N: "Equal(nil, slice)",
		F: func(is *I) {
			var s1 []string
			s2 := []string{"one", "two", "three", "four"}
			is.Equal(s1, s2) // nil slice
		},
		Fail: ` // nil slice`,
	},

	// Fail
	{
		N: "Fail()",
		F: func(is *I) {
			is.Fail() // something went wrong
		},
		Fail: "failed // something went wrong",
	},

	// NoErr
	{
		N: "NoErr(nil)",
		F: func(is *I) {
			var err error
			is.NoErr(err) // method shouldn't return error
		},
		Fail: "",
	},
	{
		N: "NoErr(error)",
		F: func(is *I) {
			err := errors.New("nope")
			is.NoErr(err) // method shouldn't return error
		},
		Fail: "err: nope // method shouldn't return error",
	},

	// OK
	{
		N: "True(1 == 2)",
		F: func(is *I) {
			is.True(1 == 2)
		},
		Fail: "not true: 1 == 2",
	},
}

func TestFailures(t *testing.T) {
	colorful, notColorful := true, false
	testFailures(t, colorful)
	testFailures(t, notColorful)
}

func testFailures(t *testing.T, colorful bool) {
	for _, test := range tests {
		tt := &mockT{}
		is := New(tt)
		var buf bytes.Buffer
		is.out = &buf
		is.colorful = colorful
		test.F(is)
		if len(test.Fail) == 0 && tt.failed {
			t.Errorf("shouldn't fail: %s", test.N)
			continue
		}
		if len(test.Fail) > 0 && !tt.failed {
			t.Errorf("didn't fail: %s", test.N)
		}
		if colorful {
			// if colorful, we won't check the messages
			// this test is run twice, one without colorful
			// statements.
			// see TestFailures
			fmt.Print(buf.String())
			continue
		}
		output := buf.String()
		output = strings.TrimSpace(output)
		if !strings.HasSuffix(output, test.Fail) {
			t.Errorf("expected `%s` to end with `%s`", output, test.Fail)
		}
	}
}

func TestRelaxed(t *testing.T) {
	tt := &mockT{}
	is := NewRelaxed(tt)
	var buf bytes.Buffer
	is.out = &buf
	is.colorful = false
	is.NoErr(errors.New("oops"))
	is.True(1 == 2)
	actual := buf.String()
	if !strings.Contains(actual, `oops`) {
		t.Errorf("missing: oops")
	}
	if !strings.Contains(actual, `1 == 2`) {
		t.Errorf("missing: 1 == 2")
	}
	if !tt.failed {
		t.Errorf("didn't fail")
	}
}

func TestLoadComment(t *testing.T) {
	comment, ok := loadComment("./testdata/example_test.go", 12)
	if !ok {
		t.Errorf("loadComment: not ok")
	}
	if comment != `this comment will be extracted` {
		t.Errorf("loadComment: bad comment %s", comment)
	}
}

func TestLoadArguments(t *testing.T) {
	arguments, ok := loadArguments("./testdata/example_test.go", 20)
	if !ok {
		t.Errorf("loadArguments: not ok")
	}
	if arguments != `a == getB()` {
		t.Errorf("loadArguments: bad arguments %s", arguments)
	}

	arguments, ok = loadArguments("./testdata/example_test.go", 28)
	if !ok {
		t.Errorf("loadArguments: not ok")
	}
	if arguments != `a == getB()` {
		t.Errorf("loadArguments: bad arguments %s", arguments)
	}

	arguments, _ = loadArguments("./testdata/example_test.go", 26)
	if len(arguments) > 0 {
		t.Errorf("should be no arguments: %s", arguments)
	}
}

// TestSubtests ensures subtests work as expected.
// https://github.com/matryer/is/issues/1
func TestSubtests(t *testing.T) {
	is := New(t)
	t.Run("sub1", func(t *testing.T) {
		is := is.New(t)
		is.Equal(1+1, 2)
	})
}
