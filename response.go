package uviews

import (
	"errors"
	"net/http"

	"github.com/usfsci/ustore"
)

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

/* func responseError(w http.ResponseWriter, origin string, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	//r := NewResponse(nil, origin, RestStatusError, "", err)
	r := ""
	if err := json.NewEncoder(w).Encode(r); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func apiErrorResponse(w http.ResponseWriter, origin string, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	//r := NewResponse(nil, origin, RestStatusError, "", err)
	r := ""
	if e := json.NewEncoder(w).Encode(r); e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}
}

func viewErrorResponse(w http.ResponseWriter, origin string, statusCode int, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
*/
