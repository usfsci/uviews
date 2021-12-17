package uviews

import (
	"html/template"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/usfsci/ustore"
)

const (
	invalidInputFlag = ""
)

var formDecoder = schema.NewDecoder()

type Form interface {
	SetCsrf(csrfField template.HTML)
	FieldState(key string) string
	GetAction() string
	SetMissing(key string)
	SetLoggedIn(state bool)
}

type DefaultForm struct {
	// Cross Site Request Forgery field for the form
	CsrfField template.HTML
	// Forms must have a submit input named "action"
	Action string `schema:"action,required"`
	// 1st missing required input
	// It will be highlighted on the form
	Missing string `schema:"-"`
	// Map of top banner menu names and links
	Menu map[string]string `schema:"-"`
	// True if user is logged in
	LoggedIn bool `schema:"-"`
	// Error message to display on page
	ViewError string `schema:"-"`
	//Labels    map[string]string
}

func (df *DefaultForm) SetCsrf(csrfField template.HTML) {
	df.CsrfField = csrfField
}

// FieldState - It is used by the template "fieldState" function to determine
// if a field should be highlighted as missing on the rendered template
func (df *DefaultForm) FieldState(key string) string {
	if key == df.Missing {
		return invalidInputFlag
	}

	return ""
}

func (df *DefaultForm) SetMissing(key string) {
	df.Missing = key
}

func (df *DefaultForm) GetAction() string {
	return df.Action
}

func (df *DefaultForm) SetLoggedIn(state bool) {
	df.LoggedIn = state
}

// DecodeForm - Decodes a form from the request Post Form. r.ParseForm should be called before.
// If there are missing required inputs flags the 1st one, stores the form in the session
// and redirects to GET.
// If there is an error the calling view should return and not attempt to execute any Action.
// The validator func should validate the form fields, returning true if all passed, false otherwise
func DecodeForm(w http.ResponseWriter, r *http.Request, session *ustore.Session, form Form, validator func(f Form) bool) error {
	pf := r.PostForm

	if err := formDecoder.Decode(form, pf); err != nil {
		// If EmptyFieldError flag and redirect to itself
		if e, ok := err.(schema.MultiError); ok {
			for _, v := range e {
				if efe, ok := v.(schema.EmptyFieldError); ok {
					// Mark the missing key
					form.SetMissing(efe.Key)
					if err := StoreDataInSession(r.Context(), session, &form); err != nil {
						HandleStoreError(w, err)
						return err
					}
					http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
					return err
				}
			}
		}

		if ok := validator(form); !ok {
			if err := StoreDataInSession(r.Context(), session, &form); err != nil {
				HandleStoreError(w, err)
				return err
			}
			http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		}

		// Any other error is considered internal
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}

	return nil
}
