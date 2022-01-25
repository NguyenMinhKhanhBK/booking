// Package main provides ...
package main

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		logrus.Info("khanhnguyen - custom middleware")
		next.ServeHTTP(rw, r)
	})
}
