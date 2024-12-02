package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/DhruvinShiroya/greenlight/internal/data"
	"github.com/DhruvinShiroya/greenlight/internal/validator"
)

func (app *application) listMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string
		Genres   []string
		Page     int
		PageSize int
		Sort     string
	}

	v := validator.New()

	qs := r.URL.Query()
	// Use our helpers to extract the title and genres query string values, falling back
	// to defaults of an empty string and an empty slice respectively if they are not
	// provided by the client.
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	// Get the page and page_size query string values as integers. Notice that we set
	// the default page value to 1 and default page_size to 20, and that we pass the
	// validator instance as the final argument here.
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 20, v)
	// Extract the sort query string value, falling back to "id" if it is not provided
	// by the client (which will imply a ascending sort on movie ID).
	input.Sort = app.readString(qs, "sort", "id")
	// Check the Validator instance for any errors and use the failedValidationResponse()
	// helper to send the client a response if necessary.
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// Dump the contents of the input struct in a HTTP response.
	fmt.Fprintf(w, "%+v\n", input)
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// create a struct which will hold information that we expect to be in the
	// http request boy , this struct will hold incoming request payload
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// decode request body to input struct using json.decoder
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// here movie variable is *pointer to the Movie struct
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// initialize validator
	v := validator.New()
	data.ValidateMovie(v, movie)
	// use valid method to check if there is any checks failed
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// call /data/movies Insert() method on our models , passing in the pointer
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// when sending HTTP response , we want to include the location at which
	// URL new movie is created , how to access it. we will send this information
	// in our headers with location header
	// create new header from http library
	header := make(http.Header)
	header.Set("Resource-Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// write json response with 201 resource created status code
	// send movie data in response body
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, header)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// show movie handler for "GET /v1/movies/:id" endpoint , now retrieve
// interpolated "id"
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		// use notFoundResponse() helper
		app.notFoundResponse(w, r)
		return
	}

	// get the movie with id from database
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
		}
	}

	// create varible payload to handle the incoming payload from request body
	var payload struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// read the request body and convert it to payload
	err = app.readJSON(w, r, &payload)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// now payload has the update movie value replace it with orignal movie
	movie.Title = payload.Title
	movie.Year = payload.Year
	movie.Runtime = payload.Runtime
	movie.Genres = payload.Genres
	// update movie version with each update
	// movie.Version = movie.Version + 1

	// validate updated movie and if fails return error  422 unprocessable entity
	v := validator.New()
	data.ValidateMovie(v, movie)

	// check if vadator contains any error , if present return  errors
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	// add the new movie to database
	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	// return updated movie
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// first get the id param from reqest
	id, err := app.readIDParam(r)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// delete the movie with id, if movie with id doesn't exist then
	// return record not found
	if err := app.models.Movies.Delete(id); err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	msg := fmt.Sprintf("movie id : %d deleted successfully", id)
	// upon successful movie delete return 200
	err = app.writeJSON(w, http.StatusOK, envelope{"msg": msg}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}
