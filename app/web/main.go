package main

import (
	"booking/config"
	"booking/handlers"
	"booking/models"
	"booking/render"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/sirupsen/logrus"
)

const PORT_NUMBER = ":8080"

var session *scs.SessionManager
var app config.AppConfig

func main() {
	if err := run(); err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Starting application at port %v", PORT_NUMBER)

	server := &http.Server{
		Addr:    PORT_NUMBER,
		Handler: routes(&app),
	}

	logrus.Info(server.ListenAndServe())

}

func run() error {
	// What to put in the session
	gob.Register(models.Reservation{})

	app := config.AppConfig{}

	// session management
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		logrus.WithError(err).Fatal("cannot create template cache")
		return err
	}
	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.SetAppConfig(&app)

	return nil
}
