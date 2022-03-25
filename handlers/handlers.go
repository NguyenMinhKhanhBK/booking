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
	"strings"
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

	room, err := re.DB.GetRoomByID(roomID)
	if err != nil {
		re.App.Session.Put(r.Context(), "error", "invalid data")
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
		Room:      room,
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

func (re *Repository) ShowLogin(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{Form: form.New(nil)})
}

func (re *Repository) PostLogin(w http.ResponseWriter, r *http.Request) {
	re.App.Session.RenewToken(r.Context())

	err := r.ParseForm()
	if err != nil {
		logrus.WithError(err).Error("cannot parse form")
		re.App.Session.Put(r.Context(), "error", "cannot parse form")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	form := form.New(r.PostForm)
	form.Require("email", "password")
	form.IsEmail("email")

	if !form.Valid() {
		render.RenderTemplate(w, r, "login.page.tmpl", &models.TemplateData{
			Form: form,
		})
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	id, _, err := re.DB.Authenticate(email, password)
	if err != nil {
		logrus.WithError(err).Error("failed to authenticate user")
		re.App.Session.Put(r.Context(), "error", "Invalid login credentials")
		http.Redirect(w, r, "/user/login", http.StatusSeeOther)
		return
	}

	re.App.Session.Put(r.Context(), "user_id", id)
	re.App.Session.Put(r.Context(), "flash", "Logged in successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (re *Repository) Logout(w http.ResponseWriter, r *http.Request) {
	re.App.Session.Destroy(r.Context())
	re.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}

func (re *Repository) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	render.RenderTemplate(w, r, "admin-dashboard.page.tmpl", &models.TemplateData{})
}

func (re *Repository) AdminNewReservation(w http.ResponseWriter, r *http.Request) {
	reservations, err := re.DB.AllNewReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations
	render.RenderTemplate(w, r, "admin-new-reservation.page.tmpl", &models.TemplateData{Data: data})
}

func (re *Repository) AdminAllReservation(w http.ResponseWriter, r *http.Request) {
	reservations, err := re.DB.AllReservations()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservations"] = reservations

	render.RenderTemplate(w, r, "admin-all-reservation.page.tmpl", &models.TemplateData{Data: data})
}

func (re *Repository) AdminShowReservation(w http.ResponseWriter, r *http.Request) {
	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	logrus.WithFields(logrus.Fields{
		"src": src,
		"id":  id,
	}).Info("GetURLInfo")

	stringMap := make(map[string]string)
	stringMap["src"] = src

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")
	stringMap["month"] = month
	stringMap["year"] = year

	// get reservation from database
	res, err := re.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data := make(map[string]interface{})
	data["reservation"] = res

	render.RenderTemplate(w, r, "admin-reservation-show.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		Form:      form.New(nil),
	})
}

func (re *Repository) AdminReservationCalendar(w http.ResponseWriter, r *http.Request) {
	// assume that there is no month/year specified
	now := time.Now()
	if r.URL.Query().Get("y") != "" {
		year, _ := strconv.Atoi(r.URL.Query().Get("y"))
		month, _ := strconv.Atoi(r.URL.Query().Get("m"))
		now = time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	}

	data := make(map[string]interface{})
	data["now"] = now

	next := now.AddDate(0, 1, 0)
	last := now.AddDate(0, -1, 0)

	nextMonth := next.Format("01")
	nextMonthYear := next.Format("2006")

	lastMonth := last.Format("01")
	lastMonthYear := last.Format("2006")

	stringMap := make(map[string]string)
	stringMap["next_month"] = nextMonth
	stringMap["next_month_year"] = nextMonthYear
	stringMap["last_month"] = lastMonth
	stringMap["last_month_year"] = lastMonthYear
	stringMap["this_month"] = now.Format("01")
	stringMap["this_month_year"] = now.Format("2006")

	// get the first and last days of the month
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	intMap := make(map[string]int)
	intMap["days_in_month"] = lastOfMonth.Day()

	rooms, err := re.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	data["rooms"] = rooms

	for _, room := range rooms {
		reservationMap := make(map[string]int)
		blockMap := make(map[string]int)

		for d := firstOfMonth; d.After(lastOfMonth) == false; d = d.AddDate(0, 0, 1) {
			reservationMap[d.Format("2006-01-2")] = 0
			blockMap[d.Format("2006-01-2")] = 0
		}

		restrictions, err := re.DB.GetRestrictionsForRoomByDate(room.ID, firstOfMonth, lastOfMonth)
		if err != nil {
			helpers.ServerError(w, err)
			return
		}

		for _, rs := range restrictions {
			if rs.ReservationID > 0 {
				// It is a reservation
				for d := rs.StartDate; d.After(rs.EndDate) == false; d = d.AddDate(0, 0, 1) {
					reservationMap[d.Format("2006-01-2")] = rs.ReservationID
				}
			} else {
				// It is a block
				blockMap[rs.StartDate.Format("2006-01-2")] = rs.ID
			}
		}

		data[fmt.Sprintf("reservation_map_%d", room.ID)] = reservationMap
		data[fmt.Sprintf("block_map_%d", room.ID)] = blockMap

		re.App.Session.Put(r.Context(), fmt.Sprintf("block_map_%d", room.ID), blockMap)
	}

	render.RenderTemplate(w, r, "admin-reservation-calendar.page.tmpl", &models.TemplateData{
		StringMap: stringMap,
		Data:      data,
		IntMap:    intMap,
	})
}

func (re *Repository) AdminPostReservation(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	exploded := strings.Split(r.RequestURI, "/")
	id, err := strconv.Atoi(exploded[4])
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	src := exploded[3]

	logrus.WithFields(logrus.Fields{
		"src": src,
		"id":  id,
	}).Info("GetURLInfo")

	stringMap := make(map[string]string)
	stringMap["src"] = src

	res, err := re.DB.GetReservationByID(id)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	res.FirstName = r.Form.Get("first_name")
	res.LastName = r.Form.Get("last_name")
	res.Email = r.Form.Get("email")
	res.Phone = r.Form.Get("phone")

	err = re.DB.UpdateReservation(res)
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	re.App.Session.Put(r.Context(), "flash", "Changes saved")

	month := r.Form.Get("month")
	year := r.Form.Get("year")

	if year == "" {
		http.Redirect(w, r, "/admin/reservations-"+src, http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

func (re *Repository) AdminProcessReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	_ = re.DB.UpdateProcessedForReservation(id, 1)

	re.App.Session.Put(r.Context(), "flash", "Reservation marked as processed")

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}

}

func (re *Repository) AdminDeleteReservation(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	src := chi.URLParam(r, "src")
	_ = re.DB.DeleteReservation(id)
	re.App.Session.Put(r.Context(), "flash", "Reservation deleted")

	year := r.URL.Query().Get("y")
	month := r.URL.Query().Get("m")

	if year == "" {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-%s", src), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%s&m=%s", year, month), http.StatusSeeOther)
	}
}

func (re *Repository) AdminPostReservationCalendar(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	year, _ := strconv.Atoi(r.Form.Get("y"))
	month, _ := strconv.Atoi(r.Form.Get("m"))

	rooms, err := re.DB.AllRooms()
	if err != nil {
		helpers.ServerError(w, err)
		return
	}

	form := form.New(r.PostForm)

	for _, room := range rooms {
		// Get the block map from the session. Loop through entire map, if we have an entry in the map
		// that does not exist in our posted data, and if the restriction id > 0, then it is the block we need to remove
		curMap := re.App.Session.Get(r.Context(), fmt.Sprintf("block_map_%d", room.ID)).(map[string]int)
		for name, value := range curMap {
			if value > 0 {
				if !form.Has(fmt.Sprintf("remove_block_%d_%s", room.ID, name)) {
					err := re.DB.DeleteBlockByID(value)
					if err != nil {
						logrus.WithError(err).Error("cannot delete block")
					}
				}
			}

		}
	}

	// Now handle new blocks
	for name := range r.PostForm {
		if strings.HasPrefix(name, "add_block_") {
			exploded := strings.Split(name, "_")
			roomID, _ := strconv.Atoi(exploded[2])
			t, _ := time.Parse("2006-01-2", exploded[3])
			err := re.DB.InsertBlockForRoom(roomID, t)
			if err != nil {
				logrus.WithError(err).Error("cannot insert block")
			}
		}
	}

	re.App.Session.Put(r.Context(), "flash", "Changes saved")
	http.Redirect(w, r, fmt.Sprintf("/admin/reservations-calendar?y=%d&m=%d", year, month), http.StatusSeeOther)
}
