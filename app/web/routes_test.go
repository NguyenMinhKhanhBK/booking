// Package main provides ...
package main

import (
	"booking/config"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRoutes(t *testing.T) {
	app := config.AppConfig{}

	mux := routes(&app)
	switch v := mux.(type) {
	case *chi.Mux:
		// do nothing
	default:
		t.Errorf("type is not *chi.Mux, but is %T", v)
	}
}
