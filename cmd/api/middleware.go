package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/DhruvinShiroya/greenlight/internal/data"
	"github.com/DhruvinShiroya/greenlight/internal/validator"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// create defer go function which will write header if there was error with request
		defer func() {
			// use built in recorver function to check if there is an error
			// and close the connection and set relevant headers
			if err := recover(); err != nil {
				w.Header().Set("Connection", "closed")
				// recover return interface{} which can be used for error value
				// by converting it to fmt.Errorf() and call serverErrorResponse function
				// to write error to response
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler { // define client struct which will hold rate limiter and last seen time
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	// declare a mutex and amp to hold client ip address from the request
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// launch a background go routine which removes old entries from the
	// client map every one minute
	go func() {
		time.Sleep(time.Minute)

		// lock mutex for inactive client clean up
		mu.Lock()

		// loop though all client , if they haven't been seen within the last five minutes
		// delete corresponding entry from clients map
		for ip, client := range clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}
		// always unlock mutex after clean up or application will halt
		mu.Unlock()
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only carry out the ratelimit if enabled
		if app.config.limiter.enable {

			// extract client ip address from request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
			}
			// lock mutex
			mu.Lock()
			// check if the ip address already exists in the map. if not present
			// add it client map with new rateLimiter
			if _, found := clients[ip]; !found {
				// create a limiter which allows control over request per second
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}
			// update the lastseen for client
			clients[ip].lastSeen = time.Now()

			// call allow method on the rate limiter for the current ip address
			// if the request isn't allowed , unlock the mutex and send too many request
			// error
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			// unlock mutex before calling next handler in the chain
			// otherwise mutex will be locked till all the downstream
			// of this middleware is also returned
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// add header "Vary": "Authorization" response my vary based on authorization header

		w.Header().Set("Vary", "Authorization")
		// get the authorization header
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			fmt.Println("token are not correct")
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		// validate token
		v := validator.New()
		if data.ValidateTokenPlainText(v, token); !v.Valid() {
			app.invalidCredentialsResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}
		// set the user to response writer
		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)

	})
}

func (app *application) RequiredActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// use context get user helper that
		user := app.contextGetUser(r)

		if !user.Activated {
			app.inActiveUserAccount(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.RequiredAuthenticatedUser(fn)
}

func (app *application) RequiredAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)

	})

}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// get the user from request
		user := app.contextGetUser(r)
		// get the permission slice for user

		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		// status forbidden for users without valid permission
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}
	return app.RequiredActivatedUser(fn)
}
