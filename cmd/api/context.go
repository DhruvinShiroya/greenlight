package main

import (
	"context"
	"net/http"

	"github.com/DhruvinShiroya/greenlight/internal/data"
)

// define context type
type contextKey string

// convert the string "user" user to a usercontext
const userContextKey = contextKey("user")

// set usercontext to a given request and provide new request with 
// user struct addd to the context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
   ctx := context.WithValue(r.Context(),userContextKey,user)
   return r.WithContext(ctx)
}

// get the user from context if not present panic
func (app *application) contextGetUser(r *http.Request) *data.User{
  user, ok:= r.Context().Value(userContextKey).(*data.User)
  if !ok {
    panic("missing user value in request context")
  }

  return user 
}
