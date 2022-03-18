package uviews

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/usfsci/ustore"
)

func ApiAdd(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "add"

	// Decode the JSON message
	if err := msgDecoder(r, ent, origin); err != nil {
		apiErr := &ApiError{
			Desc:  "unable to decode request JSON",
			Debug: err.Error(),
		}
		ApiResponseWrite(w, origin, nil, []*ApiError{apiErr}, http.StatusBadRequest)
		return
	}

	// Check ancestors length
	if len(ancestors) != ent.AncestorsRootLen() {
		e := ApiErrWrongAncestors
		e.Debug = fmt.Sprintf("expected %d ancestors, got %d", ent.AncestorsRootLen(), len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	// Store add
	if err := ent.Add(r.Context(), "", ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	ApiResponseWrite(w, origin, ent.GetID(), nil, http.StatusOK)
}

func ApiGet(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "get"

	// Get op requires 1 more ancestor than Add or List
	if len(ancestors) != ent.AncestorsRootLen() {
		e := ApiErrWrongAncestors
		e.Debug = fmt.Sprintf("expected %d ancestors, got %d", ent.AncestorsRootLen(), len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	if err := ent.Get(r.Context(), &ustore.Filter{}, ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	ApiResponseWrite(w, origin, ent, nil, http.StatusOK)
}

func ApiList(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "list"

	// Check ancestors length
	if len(ancestors) != ent.AncestorsRootLen() {
		e := ApiErrWrongAncestors
		e.Debug = fmt.Sprintf("expected %d ancestors, got %d", ent.AncestorsRootLen(), len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	ents := make([]ustore.Entity, 0)
	if err := ent.List(r.Context(), &ustore.Filter{}, &ents, ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	ApiResponseWrite(w, origin, ents, nil, http.StatusOK)
}

func msgDecoder(r *http.Request, ent ustore.Entity, origin string) error {
	defer r.Body.Close()

	err := json.NewDecoder(r.Body).Decode(ent)
	if err != nil {
		log.Printf("Rest.msgDecoder json error: %+v\n", err)
		return err
	}

	return nil
}

func listAncestors(r *http.Request) ([]ustore.SIDType, *ApiError) {
	// Extract parent info from the path
	vrs := mux.Vars(r)

	ancestors := make([]ustore.SIDType, len(vrs))
	for k, v := range vrs {
		// The ancestor position
		i, err := strconv.Atoi(k)
		if err != nil {
			return nil, &ApiError{
				Desc:  "unable to decode path",
				Debug: err.Error(),
			}
		}

		sid, err := ustore.SIDFromString(v)
		if err != nil {
			return nil, &ApiError{
				Desc:  "id not properly formatted",
				Debug: err.Error(),
			}
		}

		ancestors[i] = sid
	}

	return ancestors, nil
}

func isAuthorized(r *http.Request, ent ustore.Entity, u *ustore.User, ancestors ...ustore.SIDType) (int, *ApiError) {
	// Check autorization to try access
	can, err := ent.IsAuthorized(r.Context(), u, ancestors...)
	if err != nil {
		//ApiResponseStoreError(w, origin, err)
		return ApiErrFromStoreErr(err)
	}
	if !can {
		/*apiErr := &ApiError{
			Desc:  "user has no authority to perform request",
			Debug: "",
		}*/
		//ApiResponseWrite(w, origin, nil, []*ApiError{apiErr}, http.StatusForbidden)
		return http.StatusForbidden, &ApiError{
			Desc:  "user has no authority to perform request",
			Debug: "",
		}
	}

	return http.StatusOK, nil
}
