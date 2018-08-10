package httphandler

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
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

var handlerMessage = "Foo"

func defaultTestHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, handlerMessage)
}

func TestGCI_checkHeader(t *testing.T) {
	handler := GCI(defaultTestHandler)
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set(gciHeader, heapCheckHeader)
	w := httptest.NewRecorder()
	handler(w, req)

	body, _ := ioutil.ReadAll(w.Result().Body)
	_, err := strconv.Atoi(string(body))
	if err != nil {
		t.Fatalf("want:number got err:%q", err)
	}
}

func TestGCI_fallthrough(t *testing.T) {
	handler := GCI(defaultTestHandler)
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest("GET", "http://example.com/foo", nil))

	body, _ := ioutil.ReadAll(w.Result().Body)
	if string(body) != handlerMessage {
		t.Fatalf("want:%s got:%s", handlerMessage, string(body))
	}
}

func TestGCI_gc(t *testing.T) {
	handler := GCI(defaultTestHandler)
	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	req.Header.Set(gciHeader, "gen1")
	w := httptest.NewRecorder()
	handler(w, req)

	body, _ := ioutil.ReadAll(w.Result().Body)
	if string(body) != "" {
		t.Fatalf("want:\"\" got:%s", string(body))
	}
}
