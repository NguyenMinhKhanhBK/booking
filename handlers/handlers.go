// Package main provides ...
package handlers

import (
	"learn_web/config"
	"learn_web/models"
	"learn_web/render"
	"net/http"

	"github.com/sirupsen/logrus"
)

var Repo *Repository

type Repository struct {
	App *config.AppConfig
}

func NewRepo(a *config.AppConfig) *Repository {
	return &Repository{
		App: a,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (re *Repository) Home(w http.ResponseWriter, r *http.Request) {
	remoteIP := r.RemoteAddr
	re.App.Session.Put(r.Context(), "remote_ip", remoteIP)

	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (re *Repository) About(w http.ResponseWriter, r *http.Request) {
	// perform some logic
	remoteIP := re.App.Session.GetString(r.Context(), "remote_ip")
	logrus.WithField("remoteIP", remoteIP).Info("GetRemoteIPFromCookie")

	stringMap := map[string]string{
		"test":      "Hello Again",
		"remote_ip": remoteIP,
	}

	// send data
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
	})
}
