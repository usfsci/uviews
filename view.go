package uviews

import (
	"context"
	"net/http"

	"github.com/usfsci/ustore"
)

type view interface {
	get(w http.ResponseWriter, r *http.Request)
	post(w http.ResponseWriter, r *http.Request)
	setSession(s *ustore.Session)
	getSession() *ustore.Session
	setKind(kind string)
	getKind() string
	isKind(kind string) bool
	update(oldView view)
	canRead(ctx context.Context, vars map[string]string) (bool, *ustore.StoreError)
	canWrite(ctx context.Context, vars map[string]string) (bool, *ustore.StoreError)
}

type DefaultView struct {
	Kind string
	*ustore.Session
}

func (view *DefaultView) post(w http.ResponseWriter, r *http.Request) {
}

func (view *DefaultView) setSession(s *ustore.Session) {
	view.Session = s
}

func (view *DefaultView) getSession() *ustore.Session {
	return view.Session
}

func (view *DefaultView) setKind(kind string) {
	view.Kind = kind
}

func (view *DefaultView) getKind() string {
	return view.Kind
}

func (view *DefaultView) isKind(kind string) bool {
	return view.Kind == kind
}

func (view *DefaultView) update(oldView view) {}

func (view *DefaultView) setUser(u *ustore.User) {
	view.User = u
}

func (view *DefaultView) getUser() *ustore.User {
	return view.User
}

/*func (view *DefaultView) canRead(ctx context.Context, vars map[string]string) (bool, *store.StoreError) {
	return true, nil
}

func (view *DefaultView) canWrite(ctx context.Context, vars map[string]string) (bool, *store.StoreError) {
	return true, nil
}*/
