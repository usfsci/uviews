package uviews

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/usfsci/ustore"
)

type Response struct {
	Timestamp int64       `json:"timestamp,omitempty"`
	Origin    string      `json:"origin,omitempty"`
	Status    int         `json:"status,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     []*ApiError `json:"error,omitempty"`
}

func HandleStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, ustore.ErrNotFound) {
		// The requested entity was not found
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if errors.Is(err, ustore.ErrConstraint) {
		// A foreign key is required and was not provided or the one
		// provided does not exist
	} else if errors.Is(err, ustore.ErrInternal) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ApiResponseWrite(w http.ResponseWriter, origin string, data interface{}, errors []*ApiError, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	//w.Header().Add("Time", time.Now().UTC().Format(time.RFC3339))
	w.WriteHeader(statusCode)

	r := &Response{
		Timestamp: time.Now().In(time.UTC).Unix(),
		Status:    statusCode,
		Origin:    origin,
		Data:      data,
		Error:     errors,
	}

	if err := json.NewEncoder(w).Encode(r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ApiResponseStoreError(w http.ResponseWriter, origin string, err error) {
	/*var code int
	e := &ApiError{}

	if errors.Is(err, ustore.ErrNotFound) {
		// The requested entity was not found
		code = http.StatusNotAcceptable
		e.Desc = "the requested resource was not found"
	} else if errors.Is(err, ustore.ErrConstraint) {
		// A foreign key is required and was not provided or the one
		// provided does not exist
		code = http.StatusBadRequest
		e.Desc = "key missing or unexisting"
	} else if errors.Is(err, ustore.ErrDuplicatedKey) {
		code = http.StatusBadRequest
		e.Desc = "duplicated entry"
	} else if errors.Is(err, ustore.ErrInternal) {
		code = http.StatusInternalServerError
		e.Desc = "internal server error"
	} else {
		code = http.StatusInternalServerError
		e.Desc = "unknown error"
	}

	if debugMode {
		e.Debug = err.Error()
	}*/

	code, e := ApiErrFromStoreErr(err)
	ApiResponseWrite(w, origin, nil, []*ApiError{e}, code)
}

func ApiErrFromStoreErr(err error) (int, *ApiError) {
	var code int
	e := &ApiError{}

	if errors.Is(err, ustore.ErrNotFound) {
		// The requested entity was not found
		code = http.StatusNotAcceptable
		e.Desc = "the requested resource was not found"
	} else if errors.Is(err, ustore.ErrConstraint) {
		// A foreign key is required and was not provided or the one
		// provided does not exist
		code = http.StatusBadRequest
		e.Desc = "key missing or unexisting"
	} else if errors.Is(err, ustore.ErrDuplicatedKey) {
		code = http.StatusBadRequest
		e.Desc = "duplicated entry"
	} else if errors.Is(err, ustore.ErrInternal) {
		code = http.StatusInternalServerError
		e.Desc = "internal server error"
	} else if errors.Is(err, ustore.ErrTermsNotAccepted) {
		code = http.StatusBadRequest
		e.Desc = "terms not accepted"
	} else {
		code = http.StatusInternalServerError
		e.Desc = "unknown error"
	}

	if debugMode {
		e.Debug = err.Error()
	}

	return code, e
}
