package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DhruvinShiroya/greenlight/internal/data"
	"github.com/DhruvinShiroya/greenlight/internal/jsonlog"
	_ "github.com/lib/pq"
)

// build version
const version = "1.0.0"

// config for the http server, properties like port and env
type Config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}

	// limiter struct for limiting incoming request per second and burst value and boolean for enable and disable/ disable rate limiting
	limiter struct {
		rps    float64
		burst  int
		enable bool
	}
}

// define the application struct to hold dependencies for our HTTP handlers , helpers
// and middleware. at the moment this only contains copy of the config struct and a logger
// , but it will grow to include a lot more as out build progresses
type application struct {
	config Config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	// Declare an instance of the config struct
	var config Config

	// read the value of the port and env commandlineflags into the config struct.
	// we default to using port number and the environment "development" if no
	// flag are provided
	flag.IntVar(&config.port, "port", 4000, "api server port")
	flag.StringVar(&config.env, "env", "development", "Environment (development|staging|production)")
	// Read the DSN value from the db-dsn command-line flag into the config struct.
	flag.StringVar(&config.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	// Read the db connection pool maxOpenConns, maxIdleConns , maxIdleTime
	flag.IntVar(&config.db.maxOpenConns, "db-max-open-conns", 25, "Postgres max open connection")
	flag.IntVar(&config.db.maxIdleConns, "db-max-idle-conns", 25, "Postgres max idle connection")
	flag.StringVar(&config.db.maxIdleTime, "db-max-idle-time", "15m", "Postgres max idle connection timeout")
	// command line flags to to read the setting value of config structs
	// rate limit is enable by default and need to be disable only if in development mode if needed
	flag.Float64Var(&config.limiter.rps, "limiter-rps", 4, "Rate limiter maximum request per second")
	flag.IntVar(&config.limiter.burst, "limiter-burst", 8, "Rate limiter maximum burst request")
	flag.BoolVar(&config.limiter.enable, "limiter-enable", true, "Enable rate limiter")

	flag.Parse()

	// initialize the new logger which writes to the out stream
	// prefixed with current date and time
	// logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// initialze new json logger which writes any message (at or above)
	// severity level to the standard out stream
	logger := jsonlog.NewLogger(os.Stdout, jsonlog.LevelInfo)

	// call openDB() function to create connection pool
	db, err := openDb(config)
	if err != nil {
		log.Fatal(err)
	}

	// always close the the db pool before the main function is close
	defer db.Close()

	// print if the connection was established successfully
	// update information to the new json logger
	logger.PrintInfo("database connection is established", nil)

	// declare the instance of the application struct
	// provide the config and logger instance
	app := &application{
		config: config,
		logger: logger,
		models: data.NewModel(db),
	}
	// use httprouter instance return from app.routes() as server handler
	// declare the http server with some timeout setting , which listen on provided port
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Print the server address and environment information to logger message at info level
	// pass the additional properties to the logger
	logger.PrintInfo("Starting server :", map[string]string{
		"addr": srv.Addr,
		"env":  config.env,
	})

	// starts the HTTP server
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)

}

func openDb(cfg Config) (*sql.DB, error) {
	// use sql.Open() to create connection pool
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	// set the max connection , setmaxidelconnns, setMaxIdletime
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// parse the "<time>m" value for time.Duration
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)
	// create a context with 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// use PingContext() to establish connection
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// return sql.DB
	return db, nil
}
