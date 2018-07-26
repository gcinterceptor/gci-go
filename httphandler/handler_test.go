package httphandler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

func ExampleIntercept() {
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
