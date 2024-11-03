package main

import (
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
		app.notFound(w)
	})

	// setup the file server
	fileServer := http.FileServer(http.FS(ui.Files))
	router.Handler(http.MethodGet, "/static/*filepath", fileServer)

	// initialize middlewares for all routes
	dynamic := alice.New(app.sessionManager.LoadAndSave, noSurf, app.authenticate)

	router.Handler(http.MethodGet, "/", dynamic.ThenFunc(app.getHome))
	router.Handler(http.MethodGet, "/signup", dynamic.ThenFunc(app.getSignup))
	router.Handler(http.MethodPost, "/api/signup", dynamic.ThenFunc(app.userSignup))
	router.Handler(http.MethodGet, "/login", dynamic.ThenFunc(app.getLogin))
	router.Handler(http.MethodPost, "/api/login", dynamic.ThenFunc(app.userLogin))

	// initialize middlewares for protected routes
	protected := dynamic.Append(app.requireAuthentication)

	router.Handler(http.MethodGet, "/features", protected.ThenFunc(app.getFeatures))
	router.Handler(http.MethodPost, "/api/features", protected.ThenFunc(app.addFeatures))
	router.Handler(http.MethodGet, "/import", protected.ThenFunc(app.getImport))
	router.Handler(http.MethodPost, "/api/import", protected.ThenFunc(app.savePastData))
	router.Handler(http.MethodGet, "/predict", protected.ThenFunc(app.getPredict))
	router.Handler(http.MethodPost, "/api/logout", protected.ThenFunc(app.userLogout))

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(router)
}
