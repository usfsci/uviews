package uviews

import (
	"context"
	"net/http"

	"github.com/usfsci/ustore"
)

type View interface {
	Get(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
	SetSession(s *ustore.Session)
	GetSession() *ustore.Session
	SetKind(kind string)
	GetKind() string
	IsKind(kind string) bool
	SetUser(u *ustore.User)
	GetUser() *ustore.User
	Update(oldView View)
	Load(w http.ResponseWriter, r *http.Request) error
	CanRead(ctx context.Context, vars map[string]string) (bool, error)
	CanWrite(ctx context.Context, vars map[string]string) (bool, error)
}

type DefaultView struct {
	Kind string
	*ustore.Session
	*ustore.User
}

func (view *DefaultView) Post(w http.ResponseWriter, r *http.Request) {
}

func (view *DefaultView) SetSession(s *ustore.Session) {
	view.Session = s
}

func (view *DefaultView) GetSession() *ustore.Session {
	return view.Session
}

func (view *DefaultView) SetKind(kind string) {
	view.Kind = kind
}

func (view *DefaultView) GetKind() string {
	return view.Kind
}

func (view *DefaultView) IsKind(kind string) bool {
	return view.Kind == kind
}

func (view *DefaultView) SetUser(u *ustore.User) {
	view.User = u
}

func (view *DefaultView) GetUser() *ustore.User {
	return view.User
}

func (view *DefaultView) Update(oldView View) {}

// Load - Loads the Session & User into the view
func (view *DefaultView) Load(w http.ResponseWriter, r *http.Request) error {
	var session *ustore.Session
	var err error

	session, err = LoadSession(w, r)
	if err != nil {
		return err
	}

	if session == nil {
		// nil session means that there was no session cookie
		// Init a session with user nil
		session, err = InitSession(w, r, nil)
		if err != nil {
			return err
		}
	}

	view.Session = session

	if session.UserID != nil {
		u := &ustore.User{Base: ustore.Base{ID: session.UserID}}
		if err := u.Get(r.Context(), nil); err != nil {
			handleStoreError(w, err)
			return err
		}

		view.User = u
	}

	return nil
}
