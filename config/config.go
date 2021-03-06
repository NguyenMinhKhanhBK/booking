// Package config provides ...
package config

import (
	"booking/models"
	"html/template"

	"github.com/alexedwards/scs/v2"
)

// AppConfig holds the application config
type AppConfig struct {
	// UseCache is useful during development
	UseCache      bool
	TemplateCache map[string]*template.Template
	Session       *scs.SessionManager
	MailChan      chan models.MailData
}

func (a *AppConfig) GetTemplateCache() map[string]*template.Template {
	return a.TemplateCache
}
