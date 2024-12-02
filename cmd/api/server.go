package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// use httprouter instance return from app.routes() as server handler
	// declare the http server with some timeout setting , which listen on provided port
	// httpserver has it's own logger which write logs and if we want it to use our custom
	// ErrorLog logger uncomment the line
	// declare a HTTP server using settings from main() func
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
		// ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	// create shutDownError channle for handling any error returned
	// from graceful shutdown
	shutDownError := make(chan error)
	// start background goroutine
	go func() {
		// create channel for os.signal values
		quit := make(chan os.Signal, 1)

		// use signal.Notify() to listen for incoming SIGINT, SIGTERM signals
		// relay them to the quit channel . any other signal will be ignored
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		// read the quit signal from the channel,
		// this code block until a signal is received
		s := <-quit
		// log message to say that signal is caught
		// get the signal name and include it in the log entry properties
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})
		// create context with a 5 second timeout
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutDownError <- err
		}

		// log message for finishing background goroutines
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})
		app.wg.Wait()
		shutDownError <- nil
	}()

	// starts the HTTP server
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  app.config.env,
	})

	// calling Shutdown() on server will caus ListenAndServe() to immediatly return
	// http.ErrServerClosed error. so if we see this error, it is actually a good
	// indicaiton that gracefull shutdown  has been initiated
	// if the error is not http.ErrServerClosed()
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// otherwise wait to received the return value from shutdown channel
	err = <-shutDownError
	if err != nil {
		return err
	}

	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
