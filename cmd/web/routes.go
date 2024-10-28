package main

import (
	"fmt"
	"net/http"

	"github.com/femtech-web/baker/ui"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	// initialize the router
	router := httprouter.New()

	// declare the http-router notFound handler
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Not found")
	})

	// setup the file server
	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// initialize dynamic middlewares for all endpoints
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.home))
	router.Handler(http.MethodGet, "/signup", dynamic.ThenFunc(app.signup))
	router.Handler(http.MethodGet, "/login", dynamic.ThenFunc(app.login))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
