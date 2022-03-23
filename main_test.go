package uviews

import (
	"os"
	"testing"

	"github.com/usfsci/ustore"
)

var app *App
var etoken string

func TestMain(m *testing.M) {
	// Configure the Test Store
	ustore.ConfigMailer(func(token string, lang string, to string) error {
		etoken = token
		return nil
	})

	s := ustore.NewStore(
		"ustore_test",
		"ustore_user",
		"Cirrus-14",
		"",
		"",
	)
	defer s.Close()

	s.RegisterUserDao()

	// Start and configure the App
	app = NewApp("test_app", []byte("1234"), "11735", "", "", "")

	//app.Router.HandleFunc("/users/{0}/clients", app.ApiAuthenticate(ustore.NewClient, ApiAdd)).Methods(http.MethodPost)
	//app.Router.HandleFunc("/users/{0}/clients", app.ApiAuthenticate(ustore.NewClient, ApiList)).Methods(http.MethodGet)

	code := m.Run()
	os.Exit(code)
}
