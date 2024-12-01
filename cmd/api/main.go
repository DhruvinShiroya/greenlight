package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"os"
	"sync"
	"time"

	"github.com/DhruvinShiroya/greenlight/internal/data"
	"github.com/DhruvinShiroya/greenlight/internal/jsonlog"
	"github.com/DhruvinShiroya/greenlight/internal/mailer"
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
	// for mailer
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
}

// define the application struct to hold dependencies for our HTTP handlers , helpers
// and middleware. at the moment this only contains copy of the config struct and a logger
// , but it will grow to include a lot more as out build progresses
type application struct {
	config Config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
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

	// mailtrap credential for testing and user activation
	flag.StringVar(&config.smtp.host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&config.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&config.smtp.username, "smtp-username", "3df551409fadad", "SMTP username")
	flag.StringVar(&config.smtp.password, "smtp-password", "", "SMTP password")
	flag.StringVar(&config.smtp.sender, "smtp-sender", "Greenlight <no-reply@grd8672aa2264bb5eenlight.DhruvinShiroya.net>", "SMTP sender")
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
		mailer: mailer.New(config.smtp.host, config.smtp.port, config.smtp.username, config.smtp.password, config.smtp.sender),
	}

	// starts the HTTP server
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
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
