package uviews

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/usfsci/ustore"
)

const (
	maxMessageDelayNs = 30 * 60 * 1e9
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

	data := map[string]interface{}{
		"id":                ent.GetID(),
		"modification_time": ent.GetModificationTime().Format(time.RFC3339),
	}
	ApiResponseWrite(w, origin, data, nil, http.StatusOK)
}

func ApiUpdate(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
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
	if len(ancestors) != (ent.AncestorsRootLen() + 1) {
		e := ApiErrWrongAncestors
		e.Debug = fmt.Sprintf("expected %d ancestors, got %d", ent.AncestorsRootLen(), len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	// TODO: Store update
	if err := ent.Update(r.Context(), ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	data := map[string]interface{}{
		"id":                ent.GetID(),
		"modification_time": ent.GetModificationTime().Format(time.RFC3339),
	}
	ApiResponseWrite(w, origin, data, nil, http.StatusOK)
}

func ApiGet(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "get"

	// Get op requires 1 more ancestor than Add or List
	if len(ancestors) != (ent.AncestorsRootLen() + 1) {
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

// msgDecoder -
func msgDecoder(r *http.Request, ent ustore.Entity, origin string) error {
	defer r.Body.Close()

	msg := &Message{}

	err := json.NewDecoder(r.Body).Decode(msg)
	if err != nil {
		log.Printf("Rest.msgDecoder json error: %+v\n", err)
		return err
	}

	// Check the timestamp of the message
	t := time.Now().In(time.UTC).UnixNano()

	if msg.Timestamp > t {
		// A message from the future! hummm...
		return fmt.Errorf("messages from the future are not allowed yet")
	}

	if (t - msg.Timestamp) > int64(maxMessageDelayNs) {
		// The message is too old
		return fmt.Errorf("message is %d minutes old", t-msg.Timestamp)
	}

	if err := json.Unmarshal(msg.Data, ent); err != nil {
		return err
	}

	// For Updates the Entity ModificationTime must be before the time stamp
	if msg.Timestamp < ent.GetModificationTime().UnixNano() {
		return fmt.Errorf("message timestamp before entity modification")
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
		return ApiErrFromStoreErr(err)
	}
	if !can {
		return http.StatusForbidden, &ApiError{
			Desc:  "user has no authority to perform request",
			Debug: "",
		}
	}

	return http.StatusOK, nil
}
