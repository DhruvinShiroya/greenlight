package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	// initialize new http router
	router := httprouter.New()

	// convert notFoundResponse() to be used with httprouter http.Handler adapter
	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// register the relevant methods , url patter and handler functions for our
	// endpoints using handlefunc() method
	// http.MethodGet and http.MethodPost is constant
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// register user routes
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)

	// return the httprouter instance
	return app.recoverPanic(app.rateLimit(router))
}
