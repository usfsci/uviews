package uviews

import (
	"net/http"

	"github.com/usfsci/ustore"
)

var ApiErrNotAuthenticated = &ApiError{
	//Code:  http.StatusUnauthorized,
	Desc:  "unauthenticated",
	Debug: "no debug information",
}

func (app *App) ApiAuthenticate(
	newEntity func() ustore.Entity,
	apiHandler func(http.ResponseWriter, *http.Request, ustore.Entity, *ustore.User, []ustore.SIDType),
) http.HandlerFunc {
	const origin = "authenticate"

	return func(w http.ResponseWriter, r *http.Request) {
		uname, pass, ok := r.BasicAuth()

		if !ok || uname == "" {
			responseNotAuthenticated(w, app.name)
			return
		}

		// Check user in DB
		usr := ustore.NewUser().(*ustore.User)
		usr.Username = uname

		if err := usr.GetByName(r.Context()); err != nil {
			responseNotAuthenticated(w, app.name)
			return
		}

		if !usr.EmailConfirmed {
			responseNotAuthenticated(w, app.name)
			return
		}

		if err := usr.Authenticate(pass); err != nil {
			responseNotAuthenticated(w, app.name)
			return
		}

		// Check if user is authorized to attempt request on this entity
		ancestors, apiErr := listAncestors(r)
		if apiErr != nil {
			ApiResponseWrite(w, origin, nil, []*ApiError{apiErr}, http.StatusBadRequest)
			return
		}

		ent := newEntity()

		code, apiErr := isAuthorized(r, ent, usr, ancestors...)
		if apiErr != nil {
			ApiResponseWrite(w, origin, nil, []*ApiError{apiErr}, code)
			return
		}

		apiHandler(w, r, ent, usr, ancestors)
	}
}

func (app *App) ApiBypassAuthentication(
	newEntity func() ustore.Entity,
	apiHandler func(http.ResponseWriter, *http.Request, ustore.Entity, *ustore.User, []ustore.SIDType),
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ancestors, apiErr := listAncestors(r)
		if apiErr != nil {
			ApiResponseWrite(w, app.name, nil, []*ApiError{apiErr}, http.StatusBadRequest)
			return
		}
		apiHandler(w, r, newEntity(), nil, ancestors)
	}
}

func responseNotAuthenticated(w http.ResponseWriter, origin string) {
	ApiResponseWrite(w, origin, nil, []*ApiError{ApiErrNotAuthenticated}, http.StatusUnauthorized)
}
