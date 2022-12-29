package main

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestPopulate(t *testing.T) {
	authConfFile = "./test/auth.yml"
	amw := authenticationMiddleware{make(map[string]string)}
	amw.Populate()
}

func dummyHandler(w http.ResponseWriter, r *http.Request) {}

func newRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic(err)
	}
	// extract the escaped original host+path from url
	// http://localhost/path/here?v=1#frag -> //localhost/path/here
	opaque := ""
	if i := len(req.URL.Scheme); i > 0 {
		opaque = url[i+1:]
	}

	if i := strings.LastIndex(opaque, "?"); i > -1 {
		opaque = opaque[:i]
	}
	if i := strings.LastIndex(opaque, "#"); i > -1 {
		opaque = opaque[:i]
	}

	// Escaped host+path workaround as detailed in https://golang.org/pkg/net/url/#URL
	// for < 1.5 client side workaround
	req.URL.Opaque = opaque

	// Simulate writing to wire
	var buff bytes.Buffer
	req.Write(&buff)
	ioreader := bufio.NewReader(&buff)

	// Parse request off of 'wire'
	req, err = http.ReadRequest(ioreader)
	if err != nil {
		panic(err)
	}
	return req
}

func TestMiddlewareHealthcheck(t *testing.T) {
	// setup the middleware
	u := map[string]string{"foo": "some_user"}
	amw := authenticationMiddleware{u}
	// build the router with the middleware
	router := mux.NewRouter()
	router.HandleFunc("/", dummyHandler).Methods("GET")
	router.Use(amw.Middleware)
	// build the request
	rw := httptest.NewRecorder()
	req := newRequest("GET", "/")
	req.Header.Set("X-Auth-Token", "foo")
	router.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Fatalf("got a bad error code. %d", rw.Code)
	}
}

func TestMiddlewareGood(t *testing.T) {
	// setup the middleware
	u := map[string]string{"foo": "some_user"}
	amw := authenticationMiddleware{u}
	// build the router with the middleware
	router := mux.NewRouter()
	router.HandleFunc("/test", dummyHandler).Methods("POST")
	router.Use(amw.Middleware)
	// build the request
	rw := httptest.NewRecorder()
	req := newRequest("POST", "/test")
	req.Header.Set("X-Auth-Token", "foo")
	router.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Fatalf("got a bad error code. %d", rw.Code)
	}
}

func TestMiddlewareDenied(t *testing.T) {
	// setup the middleware
	u := map[string]string{"foo": "some_user"}
	amw := authenticationMiddleware{u}
	// build the router with the middleware
	router := mux.NewRouter()
	router.HandleFunc("/test", dummyHandler).Methods("POST")
	router.Use(amw.Middleware)
	// build the request
	rw := httptest.NewRecorder()
	req := newRequest("POST", "/test")
	req.Header.Set("X-Auth-Token", "bar")
	router.ServeHTTP(rw, req)
	if rw.Code != 403 {
		t.Fatalf("got a bad error code. %d", rw.Code)
	}
}

func TestMiddlewareGoodBasic(t *testing.T) {
	// setup the middleware
	u := map[string]string{"foo": "some_user"}
	amw := authenticationMiddleware{u}
	// build the router with the middleware
	router := mux.NewRouter()
	router.HandleFunc("/test", dummyHandler).Methods("POST")
	router.Use(amw.Middleware)
	// build the request
	rw := httptest.NewRecorder()
	req := newRequest("POST", "/test")
	// note that this is some_user:foo base64 encoded
	req.Header.Set("Authorization", "Basic c29tZV91c2VyOmZvbw==")
	router.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Fatalf("got a bad error code. %d", rw.Code)
	}
}

func TestMiddlewareDeniedBasic(t *testing.T) {
	// setup the middleware
	u := map[string]string{"foo": "some_user"}
	amw := authenticationMiddleware{u}
	// build the router with the middleware
	router := mux.NewRouter()
	router.HandleFunc("/test", dummyHandler).Methods("POST")
	router.Use(amw.Middleware)
	// build the request
	rw := httptest.NewRecorder()
	req := newRequest("POST", "/test")
	// note that this is some_user:foo2 base64 encoded
	req.Header.Set("Authorization", "Basic c29tZV91c2VyOmZvbzI=")
	router.ServeHTTP(rw, req)
	if rw.Code != 403 {
		t.Fatalf("got a bad error code. %d", rw.Code)
	}
}

func TestMiddlewareHealthCheck(t *testing.T) {
	// setup the middleware
	u := map[string]string{"foo": "some_user"}
	amw := authenticationMiddleware{u}
	// build the router with the middleware
	router := mux.NewRouter()
	router.HandleFunc("/", dummyHandler)
	router.Use(amw.Middleware)
	// build the request
	rw := httptest.NewRecorder()
	req := newRequest("GET", "/")
	// note that this is some_user:foo2 base64 encoded
	router.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Fatalf("got a bad error code. %d", rw.Code)
	}
}
