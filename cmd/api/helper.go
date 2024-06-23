package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// define envelop type
type envelope map[string]interface{}

func (app *application) readIDParam(r *http.Request) (int64, error) {

	// any request parameter in httprouter will be stored in request context
	params := httprouter.ParamsFromContext(r.Context())

	// now get the id data from params using ByName method
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil

}

func (app *application) writeJSON(w http.ResponseWriter, status int, data interface{}, headers http.Header) error {

	// pass the go object (data) to the json.Marshal() function return []bytes slice
	// using json.MarshalIndent will result in 65% longer to run and 30% more memory
	js, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// append the newline to the js
	js = append(js, '\n')

	// set content header for "application/json"
	// default header is "Content-Type: text/plain; charset=utf-8"
	for key, value := range headers {
		w.Header()[key] = value
	}

	// add the content type : application/json to header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// write the JSON as the http response body
	w.Write([]byte(js))

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {

	// define maximum size of payload or request body to 1MB
	maxByte := 1_04_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxByte))

	// to have better control on incoming POST request body DisallowUnknownFields() method before decoding
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	// decode the request body into destination
	err := dec.Decode(dst)
	if err != nil {
		// if there is error start the triage
		var syntaxError *json.SyntaxError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// error as used to check if the error type is syntax error type and returns error position
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		// decode function also return io.ErrUnexpectedEOF for syntax error in JSON
		// we check with error.Is() with and return generic error message.
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contain badly-formed JSON")
		// if there was error due to target type or destination , if the error relates to specific field
		// than field will be return or the position at which error ocurred
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type or field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type at character %d", unmarshalTypeError.Offset)
		// io.EOF error wil be return by decode() if the request body is empty.
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		// if the json contains invalid field name which can not be mapped to destination
		// decode will return error "json: unknown field "<name>"".
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldname := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldname)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxByte)
		// stop the program because because given error is due to non-nil pointer
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		default:
			return err
		}
	}
	// Call Decode() again, using a pointer to an empty anonymous struct as the
	// destination. If the request body only contained a single JSON value this will
	// return an io.EOF error. So if we get anything else, we know that there is
	// additional data in the request body and we return our own custom error message.:w

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}
