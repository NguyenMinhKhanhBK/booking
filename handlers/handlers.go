// Package main provides ...
package handlers

import (
	"booking/config"
	form "booking/forms"
	"booking/models"
	"booking/render"
	"encoding/json"
	"fmt"
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

func (re *Repository) Generals(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "generals.page.tmpl", &models.TemplateData{})
}

func (re *Repository) Majors(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "majors.page.tmpl", &models.TemplateData{})
}

func (re *Repository) Availability(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "search-availability.page.tmpl", &models.TemplateData{})
}

func (re *Repository) PostAvailability(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	res := fmt.Sprintf("PostAvailability. Start: %s - End: %s", start, end)
	w.Write([]byte(res))
}

type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (re *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	logrus.Info("khanhnguyen - AvailabilityJSON")
	resp := jsonResponse{
		OK:      true,
		Message: "Available",
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		logrus.WithError(err).Error("failed to marshal JSON")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (re *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (re *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["reservation"] = models.Reservation{}
	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form: form.New(nil),
		Data: data,
	})
}

func (re *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		logrus.WithError(err).Error("failed to parse form")
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Email:     r.Form.Get("email"),
		Phone:     r.Form.Get("phone"),
	}

	f := form.New(r.PostForm)

	f.Require("first_name", "last_name", "email", "phone")
	f.MinLength("first_name", 3)
	f.IsEmail("email")

	if !f.Valid() {
		data := make(map[string]interface{})
		data["reservation"] = reservation
		render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
			Form: f,
			Data: data,
		})
		return
	}

	re.App.Session.Put(r.Context(), "reservation", reservation)
	http.Redirect(w, r, "/reservation-summary", http.StatusSeeOther)

}

func (re *Repository) ReservationSummary(w http.ResponseWriter, r *http.Request) {
	reservation, ok := re.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		logrus.Error("cannot load item from session")
		re.App.Session.Put(r.Context(), "error", "Can't get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	re.App.Session.Remove(r.Context(), "reservation")

	data := make(map[string]interface{})
	data["reservation"] = reservation

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data: data,
	})
}
