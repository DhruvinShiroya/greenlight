package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// build version
const version = "1.0.0"

// config for the http server, properties like port and env
type Config struct {
	port int
	env  string
}

// define the application struct to hold dependencies for our HTTP handlers , helpers
// and middleware. at the moment this only contains copy of the config struct and a logger
// , but it will grow to include a lot more as out build progresses
type application struct {
	config Config
	logger *log.Logger
}

func main() {
	// Declare an instance of the config struct
	var config Config

	// read the value of the port and env commandlineflags into the config struct.
	// we default to using port number and the environment "development" if no
	// flag are provided
	flag.IntVar(&config.port, "port", 4000, "api server port")
	flag.StringVar(&config.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	// initialize the new logger which writes to the out stream
	// prefixed with current date and time
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// declare the instance of the application struct
	// provide the config and logger instance
	app := &application{
		config: config,
		logger: logger,
	}

	// create a new instance of servemux and add "/v1/healthcheck" route which dispatch requests
	// to healthcheck method
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/healthcheck", app.healthcheckHandler)

	// declare the http server with some timeout setting , which listen on provided port
	// the config struct and uses athe servemux we created about as the handler
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// starts the HTTP server
	logger.Printf("Starting %s server on %s", config.env, srv.Addr)
	err := srv.ListenAndServe()
	logger.Fatal(err)

}
