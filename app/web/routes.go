package main

import (
	"learn_web/config"
	"learn_web/handlers"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func routes(app *config.AppConfig) http.Handler {
	/*
		mux := pat.New()
		mux.Get("/", http.HandlerFunc(handlers.Repo.Home))
		mux.Get("/about", http.HandlerFunc(handlers.Repo.About))
	*/

	mux := chi.NewRouter()

	mux.Use(middleware.Logger)
	mux.Use(LogRequest)
	mux.Use(session.LoadAndSave)

	mux.Get("/", handlers.Repo.Home)
	mux.Get("/about", handlers.Repo.About)

	return mux

}
