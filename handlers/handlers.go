// Package main provides ...
package handlers

import (
	"booking/config"
	form "booking/forms"
	"booking/helpers"
	"booking/models"
	"booking/render"
	"booking/repository"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

var Repo *Repository

const (
	SEARCH_AVAIABILITY_URL = "/search-availability"
)

type Repository struct {
	App *config.AppConfig
	DB  repository.DatabaseRepo
}

func NewRepo(a *config.AppConfig, db repository.DatabaseRepo) *Repository {
	return &Repository{
		App: a,
		DB:  db,
	}
}

// NewHandlers sets the repository for the handlers
func NewHandlers(r *Repository) {
	Repo = r
}

func (re *Repository) Home(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "home.page.tmpl", &models.TemplateData{})
}

func (re *Repository) About(w http.ResponseWriter, r *http.Request) {
	// send data
	render.RenderTemplate(w, r, "about.page.tmpl", &models.TemplateData{})
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

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, start)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}
	endDate, err := time.Parse(layout, end)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	rooms, err := re.DB.SearchAvailabilityForAllRooms(startDate, endDate)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	if len(rooms) == 0 {
		re.App.Session.Put(r.Context(), "error", "No availability")
		http.Redirect(w, r, SEARCH_AVAIABILITY_URL, http.StatusSeeOther)
		return
	}

	data := make(map[string]interface{})
	data["rooms"] = rooms

	res := models.Reservation{
		StartDate: startDate,
		EndDate:   endDate,
	}

	re.App.Session.Put(r.Context(), "reservation", res)

	render.RenderTemplate(w, r, "choose-room.page.tmpl", &models.TemplateData{
		Data: data,
	})
}

type jsonResponse struct {
	OK        bool   `json:"ok"`
	Message   string `json:"message"`
	RoomID    string `json:"room_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

func (re *Repository) AvailabilityJSON(w http.ResponseWriter, r *http.Request) {
	start := r.Form.Get("start")
	end := r.Form.Get("end")
	layout := "2006-01-02"
	startDate, _ := time.Parse(layout, start)
	endDate, _ := time.Parse(layout, end)

	roomID, _ := strconv.Atoi(r.Form.Get("room_id"))

	logrus.WithFields(logrus.Fields{
		"start":   startDate,
		"end":     endDate,
		"room_id": roomID,
	}).Info("AvailabilityJSON info")

	available, err := re.DB.SearchAvailabilityByDatesByRoomID(roomID, startDate, endDate)
	if err != nil {
		resp := jsonResponse{
			OK:      false,
			Message: "Error connecting to database",
		}

		out, _ := json.MarshalIndent(resp, "", "    ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(out)
		return
	}

	resp := jsonResponse{
		OK:        available,
		Message:   "",
		StartDate: start,
		EndDate:   end,
		RoomID:    strconv.Itoa(roomID),
	}

	out, err := json.MarshalIndent(resp, "", "     ")
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(out)
}

func (re *Repository) Contact(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "contact.page.tmpl", &models.TemplateData{})
}

func (re *Repository) Reservation(w http.ResponseWriter, r *http.Request) {
	res, ok := re.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		re.App.Session.Put(r.Context(), "error", "cannot get reservation from session")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	room, err := re.DB.GetRoomByID(res.RoomID)
	if err != nil {
		re.App.Session.Put(r.Context(), "error", "room not found")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	res.Room.RoomName = room.RoomName

	re.App.Session.Put(r.Context(), "reservation", res)

	sd := res.StartDate.Format("2006-01-02")
	ed := res.EndDate.Format("2006-01-02")

	data := make(map[string]interface{})
	data["reservation"] = res
	stringMap := make(map[string]string)
	stringMap["start_date"] = sd
	stringMap["end_date"] = ed

	render.RenderTemplate(w, r, "make-reservation.page.tmpl", &models.TemplateData{
		Form:      form.New(nil),
		Data:      data,
		StringMap: stringMap,
	})
}

func (re *Repository) PostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		re.App.Session.Put(r.Context(), "error", "cannot parse form")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		logrus.Info("line 202")
		return
	}

	sd := r.Form.Get("start_date")
	ed := r.Form.Get("end_date")
	layout := "2006-01-02"

	startDate, err := time.Parse(layout, sd)
	if err != nil {
		re.App.Session.Put(r.Context(), "error", "cannot parse start date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	endDate, err := time.Parse(layout, ed)
	if err != nil {
		re.App.Session.Put(r.Context(), "error", "cannot parse end date")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomID, err := strconv.Atoi(r.Form.Get("room_id"))
	if err != nil {
		re.App.Session.Put(r.Context(), "error", "invalid room id")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	reservation := models.Reservation{
		FirstName: r.Form.Get("first_name"),
		LastName:  r.Form.Get("last_name"),
		Phone:     r.Form.Get("phone"),
		Email:     r.Form.Get("email"),
		StartDate: startDate,
		EndDate:   endDate,
		RoomID:    roomID,
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
	newReservationID, err := re.DB.InsertReservation(reservation)
	if err != nil {
		re.App.Session.Put(r.Context(), "error", "cannot insert reservation into database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	roomRestriction := models.RoomRestriction{
		StartDate:     reservation.StartDate,
		EndDate:       reservation.EndDate,
		RoomID:        reservation.RoomID,
		ReservationID: newReservationID,
		RestrictionID: 1,
	}

	_, err = re.DB.InsertRoomRestriction(roomRestriction)
	if err != nil {
		re.App.Session.Put(r.Context(), "error", "cannot insert room restriction into database")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	// send notifications
	htmlMsg := fmt.Sprintf(`
		<strong>Reservation Confirmation</strong> <br>
		Dear %s, <br>
		This email confirms your reservation from %s to %s. <br>
		Thank you for using our services! <br>
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg := models.MailData{
		To:       reservation.Email,
		From:     "me@email.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMsg,
		Template: "basic.html",
	}

	re.App.MailChan <- msg

	htmlMsg = fmt.Sprintf(`
		<strong>Reservation Notification</strong> <br>
		A reservation has been made for %s from %s to %s
	`, reservation.FirstName, reservation.StartDate.Format("2006-01-02"), reservation.EndDate.Format("2006-01-02"))

	msg = models.MailData{
		To:       "me@email.com",
		From:     "me@email.com",
		Subject:  "Reservation Confirmation",
		Content:  htmlMsg,
		Template: "basic.html",
	}

	re.App.MailChan <- msg

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

	startDate := reservation.StartDate.Format("2006-01-02")
	endDate := reservation.EndDate.Format("2006-01-02")
	stringMap := make(map[string]string)
	stringMap["start_date"] = startDate
	stringMap["end_date"] = endDate

	render.RenderTemplate(w, r, "reservation-summary.page.tmpl", &models.TemplateData{
		Data:      data,
		StringMap: stringMap,
	})
}

func (re *Repository) ChooseRoom(w http.ResponseWriter, r *http.Request) {
	roomID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res, ok := re.App.Session.Get(r.Context(), "reservation").(models.Reservation)
	if !ok {
		helpers.ServerError(w, errors.New("no session found"))
		return
	}

	res.RoomID = roomID

	re.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusSeeOther)
}

func (re *Repository) BookRoom(w http.ResponseWriter, r *http.Request) {
	roomID, _ := strconv.Atoi(r.URL.Query().Get("id"))
	layout := "2006-01-02"
	start := r.URL.Query().Get("s")
	startDate, _ := time.Parse(layout, start)
	end := r.URL.Query().Get("e")
	endDate, _ := time.Parse(layout, end)

	room, err := re.DB.GetRoomByID(roomID)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res := models.Reservation{
		RoomID:    roomID,
		StartDate: startDate,
		EndDate:   endDate,
		Room:      room,
	}

	re.App.Session.Put(r.Context(), "reservation", res)
	http.Redirect(w, r, "/make-reservation", http.StatusTemporaryRedirect)
}
