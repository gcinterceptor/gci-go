package httphandler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gcinterceptor/gci-go/gccontrol"

	"github.com/matryer/is"
)

func ExampleGCI() {
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hi")
	})
	ts := httptest.NewServer(GCI(f))
	defer ts.Close()

	// Usual serving flow.
	res, err := http.Get(ts.URL)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		panic(err)
	}
	fmt.Print(string(b))
	// Output: Hi
}

type fakeInterceptor struct {
	shouldShed bool
}

func (i *fakeInterceptor) Before() gccontrol.ShedResponse {
	if i.shouldShed {
		return gccontrol.ShedResponse{ShouldShed: true}
	}
	return gccontrol.ShedResponse{}
}

func (i *fakeInterceptor) After(sr gccontrol.ShedResponse) {}

func TestGCI(t *testing.T) {
	is := is.New(t)
	f := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hi")
	})
	gci := &fakeInterceptor{}
	ts := httptest.NewServer(newGCIHandler(gci, f))
	defer ts.Close()

	// Usual serving flow.
	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	is.Equal("Hi", string(b))

	// Shed request.
	gci.shouldShed = true
	res1, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	res1.Body.Close()
	is.Equal(http.StatusServiceUnavailable, res1.StatusCode)
}
