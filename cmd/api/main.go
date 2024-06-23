package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DhruvinShiroya/greenlight/internal/data"
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
	// Read the DSN value from the db-dsn command-line flag into the config struct.
	flag.StringVar(&config.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	// Read the db connection pool maxOpenConns, maxIdleConns , maxIdleTime
	flag.IntVar(&config.db.maxOpenConns, "db-max-open-conns", 25, "Postgres max open connection")
	flag.IntVar(&config.db.maxIdleConns, "db-max-idle-conns", 25, "Postgres max idle connection")
	flag.StringVar(&config.db.maxIdleTime, "db-max-idle-time", "15m", "Postgres max connection idle connection")
	flag.Parse()

	// initialize the new logger which writes to the out stream
	// prefixed with current date and time
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// call openDB() function to create connection pool
	db, err := openDb(config)
	if err != nil {
		log.Fatal(err)
	}

	// always close the the db pool before the main function is close
	defer db.Close()

	// print if the connection was established successfully
	fmt.Println("database connection is established")

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
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// starts the HTTP server
	logger.Printf("Starting %s server on %s", config.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)

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
