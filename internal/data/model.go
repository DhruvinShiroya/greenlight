package data

import (
	"database/sql"
	"errors"
)

// custom error for get method when record isn't found
var (
	ErrRecordNotFound = errors.New("record not found")
)

// you can add as many model as we want in our service
// add movie model to struct model
// for unit testing the any models we will replace the modles struct with interface
type Models struct {
	Movies interface {
		Insert(movie *Movie) error
		Update(movie *Movie) error
		Delete(id int64) error
		Get(id int64) (*Movie, error)
	}
}

func NewModel(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}

}

// for unit test models
func NewMockModel() Models {
	return Models{
		Movies: MockMovieModel{},
	}
}
