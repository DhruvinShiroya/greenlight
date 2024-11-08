package data

import (
	"database/sql"
	"errors"
	"time"

	"github.com/DhruvinShiroya/greenlight/internal/validator"
	"github.com/lib/pq"
)

type Movie struct {
	ID        int64     `json:"id"`    // for unique id for movie
	CreatedAt time.Time `json:"-"`     // Timestamp for when the movie is added to our database
	Title     string    `json:"title"` // Movie title
	Year      int32     `json:"year"`  // release year
	// Runtime has MarshalJSON method which return custom json output with "<runtime> min"
	Runtime Runtime  `json:"runtime,omitempty"` // Movie runtime (in minutes)
	Genres  []string `json:"genres,omitempty"`  // Movie genres
	Version int32    `json:"version"`           // the version number start at 1 and will be incremented each
	// time the movie is update
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	// use check method to perform validation check
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "title must not be more than 500 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")

	v.Check(movie.Runtime != 0, "runtime", "must be provided ")
	v.Check(movie.Runtime > 0, "runtime", "must be postive")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genres")
	v.Check(len(movie.Genres) <= 5, "genres", "must have less than 5 genres")
	// check if all the genres are unique or not
	v.Check(validator.Unique(movie.Genres), "genres", "must contain unique genres")
}

// define movie model

type MovieModel struct {
	DB *sql.DB
}

// add a placeholder method for inserting the new record into movie table
func (m MovieModel) Insert(movie *Movie) error {
	// define sql qeury for interting record in database
	query := `
	    INSERT INTO movies (title , year , runtime, genres)
        VALUES ($1, $2, $3, $4)
        RETURNING id, created_at, version`
	// create an args slice for the placeholder paramter
	// define empty slice interface and define slice immediately next to our SQL query
	args := []interface{}{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	// because our query returns the value use the method QueryRow() otherwise use Exec() command to execute query
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// add placeholder method for getting the record from movie table
func (m MovieModel) Get(id int64) (*Movie, error) {
	// check if id is valid positive
	// postgres bigserial starts form 1 auto-increment
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// prepare query for fetching records form database with id param
	query := `SELECT id, created_at, title , year, runtime, genres ,version FROM movies WHERE id = $1`

	// declare movie for result
	var movie Movie
	// use pq.QueryRow() method for executing query, scan the response data into field of Movie struct
	// if getting array from result, we need to convert it to pq.Array(arr) to scan target for the genres
	err := m.DB.QueryRow(query, id).Scan(&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	// if not error return movie struct
	return &movie, nil
}

// add method for update the record in the movie table
func (m MovieModel) Delete(id int64) error {
	// check if the id is positive
	if id < 1 {
		return ErrRecordNotFound
	}
	// construct qeury to delete the movie with id
	qeury := `DELETE FROM movies WHERE id = $1`

	// since the we are not returning anything we can use exec() method
	// which will return result.rowsaffected() which contains information about how many
	// rows has been affected , if 0 rows affectd means that movies with id
	// coulnd not be found and if found it will return 1 in  result.rowsaffected()
	result, err := m.DB.Exec(qeury, id)
	if err != nil {
		return err
	}
	rowsaffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsaffected == 0 {
		return ErrRecordNotFound
	}
	// otherwise result sould be one since we are using primary key to delete record
	// return nil upon movie delete
	return nil
}

// add method to update the record in the movie table
func (m MovieModel) Update(movie *Movie) error {
	// write update movie query  for title, runtime, genres, year
	// also update the version with each update
	query := `
       UPDATE movies
       SET title = $1, year = $2, runtime = $3, genres = $4, version= version + 1
       WHERE id= $5
       RETURNING version`

	args := []interface{}{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
	}
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// mock movie struct for unit testing
type MockMovieModel struct {
	DB *sql.DB
}

func (m MockMovieModel) Insert(movie *Movie) error {
	return nil
}

func (m MockMovieModel) Update(movie *Movie) error {
	return nil
}

func (m MockMovieModel) Delete(id int64) error {
	return nil
}

func (m MockMovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}
