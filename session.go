package uviews

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/usfsci/uauth"
	"github.com/usfsci/ustore"
)

// sessionIDCookieName - An App can have only one Session Cookie Name
var sessionIDCookieName string

// SetupSessions - Sets the name of the session Cookie
// Should be called only once, on App startup
// Must be all lowercase, starting with underscore and
// only alphanum
func SetupSessions(sessionCookieName string) {
	sessionIDCookieName = sessionCookieName
}

// InitSession - Starts a DB session and sets a session id cookie
func InitSession(w http.ResponseWriter, r *http.Request, userID ustore.SIDType) (*ustore.Session, error) {
	// Build a DB session
	s := &ustore.Session{}
	if err := s.Add(r.Context(), userID, ""); err != nil {
		handleStoreError(w, err)
		return nil, err
	}

	// Set the cookie
	tokenStr, err := uauth.EncodeToken(s.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil, err
	}

	http.SetCookie(w,
		&http.Cookie{
			Name:     sessionIDCookieName,
			Value:    tokenStr,
			MaxAge:   0, // Session cookie
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

	return s, nil
}

// LoadSession - Uses the sessionid cookie to get the Session from the store
func LoadSession(w http.ResponseWriter, r *http.Request) (*ustore.Session, error) {
	// Get session id cookie
	cookie, err := r.Cookie(sessionIDCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			// No cookie, return a nil session but no error
			return nil, nil
		}
		return nil, err
	}

	var id ustore.SIDType
	if err := uauth.DecodeToken(&id, cookie.Value); err != nil {
		// Wrong or expired token, redirect to login
		return nil, err
	}

	log.Printf("sessionid= %s\n", hex.EncodeToString(id))

	// Get DB session
	session := &ustore.Session{
		Base: ustore.Base{
			ID: id,
		},
	}
	if err := session.Get(r.Context(), nil); err != nil {
		handleStoreError(w, err)
		return nil, err
	}

	return session, nil
}

func StoreDataInSession(ctx context.Context, session *ustore.Session, object interface{}) error {
	// Save the view in the session for use in the GET
	data, err := json.Marshal(object)
	if err != nil {
		return err
	}

	session.Data = data

	return session.Update(ctx)
}

// restoreFormFromSession - Restores View fields from the session
// If there is no data it does nothing
// It clears the data stored in the session
func RestoreDataFromSession(ctx context.Context, session *ustore.Session, data []byte, object interface{}) error {
	if data == nil || len(data) < 1 {
		return nil
	}

	if err := json.Unmarshal(data, object); err != nil {
		return err
	}

	session.Data = []byte("")
	return session.Update(ctx)
}

// redirect - Redirects appending a reference to the original path to the redirected URL
func redirect(w http.ResponseWriter, r *http.Request, path string, statusCode int) {
	if r.URL.RawQuery == "" {
		http.Redirect(w, r, path+pathToPars(r.URL.Path), statusCode)
	} else {
		http.Redirect(w, r, path+"?"+r.URL.RawQuery, statusCode)
	}
}

func redirectAppendQuery(w http.ResponseWriter, r *http.Request, path string, query string) {
	pq := path
	if len(query) > 0 {
		pq += "?" + query
	}

	http.Redirect(w, r, pq, http.StatusSeeOther)
}

func pathToPars(path string) string {
	// Path for redirections if needed
	return "?rp=" + base64.RawURLEncoding.EncodeToString([]byte(path))
}

func queryToPath(query string) (string, error) {
	if len(query) <= 0 {
		return "", nil
	}

	s := strings.Split(query, "=")
	if len(s) != 2 || s[0] != "rp" {
		return "", fmt.Errorf("invalid query")
	}

	// The base64 path is the last part of the array
	b, err := base64.RawURLEncoding.DecodeString(s[1])
	if err != nil {
		return "", err
	}

	return string(b), nil
}
