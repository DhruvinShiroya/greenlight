package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/DhruvinShiroya/greenlight/internal/data"
	"github.com/DhruvinShiroya/greenlight/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	//validate email and passsword
	v := validator.New()

	data.ValidateEmail(v, input.Email)
	data.ValidatePassword(v, input.Password)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// get the user and check if the credential are valid
	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// match password
	match, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// if password don't match than we user invalid credential
	if !match {
		app.invalidCredentialsResponse(w, r)
		return
	}

  // delete old tokens with Authentication scope
  err = app.models.Token.DeleteAllForUser(user.ID,data.ScopeAuthentication)
  if err != nil {
    app.serverErrorResponse(w,r,err)
  }

	// generate new token with scope authentication
	token, err := app.models.Token.New(user.ID, time.Hour*24, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}
