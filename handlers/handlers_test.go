package handlers

import (
	"booking/mocks"
	"booking/models"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type postData struct {
	key   string
	value string
}

var tests = []struct {
	name               string
	url                string
	method             string
	expectedStatusCode int
}{
	{"home", "/", http.MethodGet, http.StatusOK},
	{"about", "/about", http.MethodGet, http.StatusOK},
	{"generals-quarters", "/generals-quarters", http.MethodGet, http.StatusOK},
	{"majors-suite", "/majors-suite", http.MethodGet, http.StatusOK},
	{"search-availability", "/search-availability", http.MethodGet, http.StatusOK},
	{"contact", "/contact", http.MethodGet, http.StatusOK},

	//{"make-reservation", "/make-reservation", http.MethodGet, []postData{}, http.StatusOK},
	//{"post-search-availability", "/search-availability", http.MethodPost, []postData{
	//	{key: "start", value: "2022-01-01"},
	//	{key: "end", value: "2022-01-02"},
	//}, http.StatusOK},

	//{"post-search-availability-json", "/search-availability-json", http.MethodPost, []postData{
	//	{key: "start", value: "2022-01-01"},
	//	{key: "end", value: "2022-01-02"},
	//}, http.StatusOK},

	//{"make-reservation-post", "/make-reservation", http.MethodPost, []postData{
	//	{key: "first_name", value: "Khanh"},
	//	{key: "last_name", value: "Nguyen"},
	//	{key: "email", value: "khanhnguyen@gmail.com"},
	//	{key: "phone", value: "123456"},
	//}, http.StatusOK},
}

func TestGetHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewServer(routes)
	defer ts.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.method == http.MethodGet {
				resp, err := ts.Client().Get(ts.URL + test.url)
				assert.NoError(t, err)
				assert.Equal(t, test.expectedStatusCode, resp.StatusCode)
			} else {
				values := url.Values{}
				resp, err := ts.Client().PostForm(ts.URL+test.url, values)
				assert.NoError(t, err)
				assert.Equal(t, test.expectedStatusCode, resp.StatusCode)
			}
		})
	}
}

func TestRepository_Reservation(t *testing.T) {
	ctrl := gomock.NewController(nil)
	mockDB := mocks.NewMockDatabaseRepo(ctrl)
	Repo.DB = mockDB

	mockDB.EXPECT().GetRoomByID(gomock.Any())

	reservation := models.Reservation{
		RoomID: 1,
		Room: models.Room{
			ID:       1,
			RoomName: "General's Quarters",
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/make-reservation", nil)
	ctx := getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.Reservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("reservation handler returns wrong response code: got %v, wanted: %v", rr.Code, http.StatusOK)
	}

	// test case where reservation is not in session (reset everything)
	req = httptest.NewRequest(http.MethodGet, "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returns wrong response code: got %v, wanted: %v", rr.Code, http.StatusOK)
	}

	// test with non-existing room
	req = httptest.NewRequest(http.MethodGet, "/make-reservation", nil)
	ctx = getCtx(req)
	session.Put(ctx, "reservation", reservation)
	req = req.WithContext(ctx)
	rr = httptest.NewRecorder()

	mockDB.EXPECT().GetRoomByID(gomock.Any()).Return(models.Room{}, errors.New("room not found"))
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returns wrong response code: got %v, wanted: %v", rr.Code, http.StatusOK)
	}
}

func getCtx(req *http.Request) context.Context {
	ctx, err := session.Load(req.Context(), req.Header.Get("X-Session"))
	if err != nil {
		logrus.WithError(err).Error("cannot load context")
		return nil
	}

	return ctx
}

func TestRepository_PostReservation(t *testing.T) {
	ctrl := gomock.NewController(nil)
	mockDB := mocks.NewMockDatabaseRepo(ctrl)
	Repo.DB = mockDB

	mockDB.EXPECT().InsertReservation(gomock.Any())
	mockDB.EXPECT().InsertRoomRestriction(gomock.Any())

	reqBody := "start_date=2050-01-01"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Khanh")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Nguyen")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=khanhnguyen@gmail.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	/*
		postedData := url.Values{}
		postedData.Add("start_date", "2050-01-01")
		postedData.Add("end_date", "2050-01-02")
		postedData.Add("first_name", "Khanh")
		postedData.Add("last_name", "Nguyen")
		postedData.Add("email", "khanhnguyen@gmail.com")
		postedData.Add("phone", "123456789")
		postedData.Add("room_id", "1")
	*/

	req, _ := http.NewRequest(http.MethodPost, "/make-reservation", strings.NewReader(reqBody))
	ctx := getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Repo.PostReservation)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusSeeOther {
		t.Errorf("reservation handler returns wrong response code for successful case: got %v, wanted: %v", rr.Code, http.StatusSeeOther)
	}

	// Test for missing post body
	req, _ = http.NewRequest(http.MethodPost, "/make-reservation", nil)
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returns wrong response code for invlid form case: got %v, wanted: %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid start date
	reqBody = "start_date=invalid_format"
	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=2050-01-02")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Khanh")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Nguyen")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=khanhnguyen@gmail.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	req, _ = http.NewRequest(http.MethodPost, "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returns wrong response code for invalid start date: got %v, wanted: %v", rr.Code, http.StatusTemporaryRedirect)
	}

	// Test for invalid end date
	reqBody = "start_date=2050-01-01"

	reqBody = fmt.Sprintf("%s&%s", reqBody, "end_date=invalid_format")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "first_name=Khanh")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "last_name=Nguyen")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "email=khanhnguyen@gmail.com")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "phone=123456789")
	reqBody = fmt.Sprintf("%s&%s", reqBody, "room_id=1")

	req, _ = http.NewRequest(http.MethodPost, "/make-reservation", strings.NewReader(reqBody))
	ctx = getCtx(req)
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusTemporaryRedirect {
		t.Errorf("reservation handler returns wrong response code for invalid end date: got %v, wanted: %v", rr.Code, http.StatusTemporaryRedirect)
	}
}
