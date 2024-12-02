package data

import (
	"database/sql"
	"errors"
)

// custom error for get method when record isn't found
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// you can add as many model as we want in our service
// add movie model to struct model
// for unit testing the any models we will replace the modles struct with interface
type Models struct {
	Movies      MovieModel
	Users       UserModel
	Token       TokenModel
	Permissions PermissionsModel
}

func NewModel(db *sql.DB) Models {
	return Models{
		Movies:      MovieModel{DB: db},
		Users:       UserModel{DB: db},
		Token:       TokenModel{DB: db},
		Permissions: PermissionsModel{DB: db},
	}
}

// for unit test models
/*
func NewMockModel() Models {
	return Models{
		Movies: MockMovieModel{},
	}
}
*/
