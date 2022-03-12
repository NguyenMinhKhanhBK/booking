package helpers

import (
	"booking/config"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/sirupsen/logrus"
)

var app *config.AppConfig

func SetAppConfig(a *config.AppConfig) {
	app = a
}

func ClientError(w http.ResponseWriter, status int) {
	logrus.Infof("Client error code %v", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	if err == nil {
		logrus.Error("error is nil")
		return
	}
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	logrus.Error(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
