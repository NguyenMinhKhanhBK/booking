package main

import (
	"learn_web/config"
	"learn_web/handlers"
	"learn_web/render"
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/sirupsen/logrus"
)

const PORT_NUMBER = ":8080"

var session *scs.SessionManager

func main() {
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
	}
	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.SetAppConfig(&app)

	logrus.Infof("Starting application at port %v", PORT_NUMBER)

	server := &http.Server{
		Addr:    PORT_NUMBER,
		Handler: routes(&app),
	}

	logrus.Info(server.ListenAndServe())

}
