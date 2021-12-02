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
	Update(oldView View)
	CanRead(ctx context.Context, vars map[string]string) (bool, *ustore.StoreError)
	CanWrite(ctx context.Context, vars map[string]string) (bool, *ustore.StoreError)
}

type DefaultView struct {
	Kind string
	*ustore.Session
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

func (view *DefaultView) Update(oldView View) {}

func (view *DefaultView) GetUser() *ustore.User {
	return view.User
}

/*func (view *DefaultView) canRead(ctx context.Context, vars map[string]string) (bool, *store.StoreError) {
	return true, nil
}

func (view *DefaultView) canWrite(ctx context.Context, vars map[string]string) (bool, *store.StoreError) {
	return true, nil
}*/
