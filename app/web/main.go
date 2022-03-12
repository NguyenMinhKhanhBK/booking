package main

import (
	"booking/config"
	"booking/handlers"
	"booking/models"
	"booking/render"
	"booking/repository"
	sqldriver "booking/sql_driver"
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
	gob.Register(models.User{})
	gob.Register(models.Room{})
	gob.Register(models.Restriction{})

	app := config.AppConfig{}

	// session management
	session = scs.New()
	session.Lifetime = 24 * time.Hour
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	session.Cookie.Secure = false

	app.Session = session

	// connect to database
	logrus.Info("Connecting to database...")
	db, err := sqldriver.ConnectSQL("host=localhost port=5432 dbname=booking user=postgres password=postgres")
	if err != nil {
		logrus.WithError(err).Fatal("Cannot connect to database. Dying...")
	}
	logrus.Info("Connected to database!")

	repoDB := repository.NewPostgresRepo(&app, db)

	tc, err := render.CreateTemplateCache()
	if err != nil {
		logrus.WithError(err).Fatal("cannot create template cache")
		return err
	}
	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app, repoDB)
	handlers.NewHandlers(repo)

	render.SetAppConfig(&app)

	return nil
}
