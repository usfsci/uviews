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
	// 30 sec max life for a message
	maxMessageDelayNs = 30 * 1e9
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
		e := ApiErrBadRequest
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
	const origin = "update"

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
		e := ApiErrBadRequest
		e.Debug = fmt.Sprintf("expected %d ancestors, got %d", ent.AncestorsRootLen(), len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	// Updates should have a non-zero modification time
	if ent.GetModificationTime().IsZero() {
		e := ApiErrBadRequest
		e.Debug = "got zero modification time on update"
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	// Store update
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

func ApiDelete(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "delete"

	// Get op requires 1 more ancestor than Add or List
	if len(ancestors) != (ent.AncestorsRootLen() + 1) {
		e := ApiErrBadRequest
		e.Debug = fmt.Sprintf("expected %d ancestors, got %d", ent.AncestorsRootLen(), len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	// TODO: How to handle the Time of requests without timestamp (no body)
	if err := ent.Delete(r.Context(), time.Now().In(time.UTC), ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	ApiResponseWrite(w, origin, ent, nil, http.StatusOK)
}

func ApiGet(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "get"

	// Get op requires 1 more ancestor than Add or List
	if len(ancestors) != (ent.AncestorsRootLen() + 1) {
		e := ApiErrBadRequest
		e.Debug = fmt.Sprintf("expected %d ancestors, got %d", ent.AncestorsRootLen()+1, len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	if err := ent.Get(r.Context(), &ustore.Filter{}, ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	ent.Zero()

	ApiResponseWrite(w, origin, ent, nil, http.StatusOK)
}

func ApiList(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "list"

	// Check ancestors length
	if len(ancestors) != ent.AncestorsRootLen() {
		e := ApiErrBadRequest
		e.Debug = fmt.Sprintf("expected %d ancestors, got %d", ent.AncestorsRootLen(), len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
	}

	ents := make([]ustore.Entity, 0)
	if err := ent.List(r.Context(), &ustore.Filter{}, &ents, ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	for _, e := range ents {
		e.Zero()
	}

	ApiResponseWrite(w, origin, ents, nil, http.StatusOK)
}

// ApiEmailValidate - Validates posted Token vs Email received token
// ent must be a *RawToken
func ApiEmailValidate(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "validate-email"

	// Decode the JSON message into an Entity of *User
	if err := msgDecoder(r, ent, origin); err != nil {
		apiErr := &ApiError{
			Desc:  "unable to decode request JSON",
			Debug: err.Error(),
		}
		ApiResponseWrite(w, origin, nil, []*ApiError{apiErr}, http.StatusBadRequest)
		return
	}

	// There must be 1 ancestor (the UserID)
	if len(ancestors) != 1 {
		e := ApiErrBadRequest
		e.Debug = fmt.Sprintf("expected 1 ancestor, got %d", len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Will crash on development if the passed ent is not a *RawToken
	rawTok := ent.(*ustore.RawToken)
	if rawTok.Token == "" {
		e := ApiErrBadRequest
		e.Debug = "token cannot be empty"
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Get the user
	u1 := &ustore.User{}
	if err := u1.Get(r.Context(), &ustore.Filter{}, ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	// Validate token vs. the authenticated user token
	if err := u1.ValidateToken(rawTok.Token); err != nil {
		e := ApiErrBadRequest
		e.Debug = err.Error()
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Mark as valid
	u1.EmailConfirmed = true
	if err := u1.UpdateEmailConfirmed(r.Context(), time.Now().In(time.UTC), ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	// Response with good status and no body
	ApiResponseWrite(w, origin, nil, nil, http.StatusOK)
}

// ApiPasswordReset - Validates posted Token vs Email received token
// and resets password
// ent must be a *RawToken
func ApiPasswordReset(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "password-reset"

	// Decode the JSON message into an Entity of *User
	if err := msgDecoder(r, ent, origin); err != nil {
		apiErr := &ApiError{
			Desc:  "unable to decode request JSON",
			Debug: err.Error(),
		}
		ApiResponseWrite(w, origin, nil, []*ApiError{apiErr}, http.StatusBadRequest)
		return
	}

	// There must be 1 ancestor (the UserID)
	if len(ancestors) != 1 {
		e := ApiErrBadRequest
		e.Debug = fmt.Sprintf("expected 1 ancestor, got %d", len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Will crash on development if the passed ent is not a *RawToken
	rawTok := ent.(*ustore.RawToken)
	if rawTok.Token == "" {
		e := ApiErrBadRequest
		e.Debug = "token cannot be empty"
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Get the user
	u1 := &ustore.User{}
	if err := u1.Get(r.Context(), &ustore.Filter{}, ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	// The user email must have been previously validated
	if !u1.EmailConfirmed {
		e := ApiErrBadRequest
		e.Debug = "user email has not been validated yet"
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Validate token vs. the authenticated user token
	if err := u1.ValidateToken(rawTok.Token); err != nil {
		e := ApiErrBadRequest
		e.Debug = err.Error()
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Update Password
	if err := (&ustore.User{
		Base:     ustore.Base{ModificationTime: time.Now().In(time.UTC)},
		Password: []byte(rawTok.Password),
	}).Update(r.Context(), ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	// Response with good status and no body
	ApiResponseWrite(w, origin, nil, nil, http.StatusOK)
}

func ApiGetToken(w http.ResponseWriter, r *http.Request, ent ustore.Entity, u *ustore.User, ancestors []ustore.SIDType) {
	const origin = "password-reset"

	// Decode the JSON message into an Entity of *User
	if err := msgDecoder(r, ent, origin); err != nil {
		apiErr := &ApiError{
			Desc:  "unable to decode request JSON",
			Debug: err.Error(),
		}
		ApiResponseWrite(w, origin, nil, []*ApiError{apiErr}, http.StatusBadRequest)
		return
	}

	// There must be 1 ancestor (the UserID)
	if len(ancestors) != 1 {
		e := ApiErrBadRequest
		e.Debug = fmt.Sprintf("expected 1 ancestor, got %d", len(ancestors))
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Will crash on development if the passed ent is not a *RawToken
	rawTok := ent.(*ustore.RawToken)
	if rawTok.Token == "" {
		e := ApiErrBadRequest
		e.Debug = "token cannot be empty"
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Get the user
	u1 := &ustore.User{}
	if err := u1.Get(r.Context(), &ustore.Filter{}, ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	// The user email must have been previously validated
	if !u1.EmailConfirmed {
		e := ApiErrBadRequest
		e.Debug = "user email has not been validated yet"
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Validate token vs. the authenticated user token
	if err := u1.ValidateToken(rawTok.Token); err != nil {
		e := ApiErrBadRequest
		e.Debug = err.Error()
		ApiResponseWrite(w, origin, nil, []*ApiError{e}, http.StatusBadRequest)
		return
	}

	// Update Password
	u1.Password = []byte(rawTok.Password)
	if err := u1.Update(r.Context(), ancestors...); err != nil {
		ApiResponseStoreError(w, origin, err)
		return
	}

	// Response with good status and no body
	ApiResponseWrite(w, origin, nil, nil, http.StatusOK)
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
