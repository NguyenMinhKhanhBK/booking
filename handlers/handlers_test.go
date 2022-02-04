package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", http.MethodGet, []postData{}, http.StatusOK},
	{"about", "/about", http.MethodGet, []postData{}, http.StatusOK},
	{"generals-quarters", "/generals-quarters", http.MethodGet, []postData{}, http.StatusOK},
	{"majors-suite", "/majors-suite", http.MethodGet, []postData{}, http.StatusOK},
	{"search-availability", "/search-availability", http.MethodGet, []postData{}, http.StatusOK},
	{"contact", "/contact", http.MethodGet, []postData{}, http.StatusOK},
	{"make-reservation", "/make-reservation", http.MethodGet, []postData{}, http.StatusOK},

	{"post-search-availability", "/search-availability", http.MethodPost, []postData{
		{key: "start", value: "2022-01-01"},
		{key: "end", value: "2022-01-02"},
	}, http.StatusOK},

	{"post-search-availability-json", "/search-availability-json", http.MethodPost, []postData{
		{key: "start", value: "2022-01-01"},
		{key: "end", value: "2022-01-02"},
	}, http.StatusOK},

	{"make-reservation-post", "/make-reservation", http.MethodPost, []postData{
		{key: "first_name", value: "Khanh"},
		{key: "last_name", value: "Nguyen"},
		{key: "email", value: "khanhnguyen@gmail.com"},
		{key: "phone", value: "123456"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewServer(routes)
	defer ts.Close()

	for _, test := range tests {
		if test.method == http.MethodGet {
			resp, err := ts.Client().Get(ts.URL + test.url)
			assert.NoError(t, err)
			assert.Equal(t, resp.StatusCode, test.expectedStatusCode)
		} else {
			values := url.Values{}
			for _, v := range test.params {
				values.Add(v.key, v.value)
			}
			resp, err := ts.Client().PostForm(ts.URL+test.url, values)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatusCode, resp.StatusCode)
		}
	}
}
