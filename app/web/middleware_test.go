// Package main provides ...
package main

import (
	"net/http"
	"testing"
)

func TestNoSurf(t *testing.T) {
	myHandler := &myHandler{}
	h := NoSurf(myHandler)

	switch v := h.(type) {
	case http.Handler:
		// do nothing
	default:
		t.Errorf("type is not http.Handler, but is %T", v)
	}
}
