// Package render provides ...
package render

import (
	"booking/config"
	"booking/models"
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

var app *config.AppConfig

func SetAppConfig(a *config.AppConfig) {
	app = a
}

var functions = template.FuncMap{}

func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, data *models.TemplateData) {
	var tc map[string]*template.Template
	// get the template cache from the app config
	if app.UseCache {
		tc = app.GetTemplateCache()
	} else {
		tc, _ = CreateTemplateCache()
	}

	t, ok := tc[tmpl]
	if !ok {
		logrus.Errorf("%v not found", tmpl)
		return
	}

	buf := new(bytes.Buffer)
	t.Execute(buf, data)

	_, err := buf.WriteTo(w)
	if err != nil {
		logrus.WithError(err).Error("fail to write template to browser")
	}
}

func CreateTemplateCache() (map[string]*template.Template, error) {
	myCache := map[string]*template.Template{}
	pages, err := filepath.Glob("./templates/*.page.tmpl")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)
		logrus.WithFields(logrus.Fields{
			"page": page,
			"name": name,
		}).Info("current page")
		ts, err := template.New(name).Funcs(functions).ParseFiles(page)
		if err != nil {
			return nil, err
		}

		matches, err := filepath.Glob("./templates/*.layout.tmpl")
		if err != nil {
			return nil, err
		}

		if len(matches) > 0 {
			ts, err = ts.ParseGlob("./templates/*.layout.tmpl")
			if err != nil {
				return nil, err
			}
		}

		myCache[name] = ts
	}

	return myCache, nil
}
