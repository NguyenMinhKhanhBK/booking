package render

import (
	"booking/config"
	"booking/models"
	"encoding/gob"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
)

var session *scs.SessionManager
var testApp config.AppConfig

func TestMain(m *testing.M) {
	// What to put in the session
	gob.Register(models.Reservation{})

	// session management
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	testApp.Session = session

	app = &testApp

	os.Exit(m.Run())
}

type myWriter struct{}

func (w *myWriter) Header() http.Header {
	h := http.Header{}
	return h
}

func (w *myWriter) Write(data []byte) (int, error) {
	return len(data), nil
}

func (w *myWriter) WriteHeader(statusCode int) {

}
