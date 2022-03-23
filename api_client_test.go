package uviews

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/usfsci/ustore"
)

func TestClientAdd(t *testing.T) {
	// Create a user for the test
	u, err := createTestUser("osm1608@gmail.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	// Add a client
	c1 := &ustore.Client{
		Name:              "Test Client",
		NotificationToken: "notif token",
		Os:                "ios",
		Sdk:               "12",
	}

	path := fmt.Sprintf("/users/%s/clients", u.ID)

	r, err := postMsg("osm1608@gmail.com", "Pass123+Q", NewMessageSim(c1), path, http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n\n", r)
	fmt.Printf("CLIENT ADDED\n")

	// Get the client to verify op
	cidString := r.Data.(map[string]interface{})["id"].(string)
	clientID, _ := ustore.SIDFromString(cidString)

	c := &ustore.Client{}
	if err := c.Get(context.Background(), &ustore.Filter{}, u.ID, clientID); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n\n", c)
	fmt.Printf("CLIENT RETRIEVED\n")

	// Erase the test user & client
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := (&ustore.User{}).Erase(context.Background(), mt, u.ID); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("TEST USER AND CLIENT ERASED\n")
}

func TestWrongCredentialsClientAdd(t *testing.T) {
	// Create a user for the test
	u, err := createTestUser("osm1608@gmail.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	// Add a client
	c := &ustore.Client{
		Name:              "Test Client",
		NotificationToken: "notif token",
		Os:                "ios",
		Sdk:               "12",
	}

	path := fmt.Sprintf("/users/%s/clients", u.ID)

	r, err := postMsg(u.Username, "WrongPass", NewMessageSim(c), path, http.StatusUnauthorized, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n\n", r)
	fmt.Printf("CLIENT ADD WRONG PASS ATTEMPT REJECTED\n")

	r, err = postMsg("wrong uname", "Pass123+Q", NewMessageSim(c), path, http.StatusUnauthorized, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n\n", r)
	fmt.Printf("CLIENT ADD WRONG EMAIL ATTEMPT REJECTED\n")

	// Erase the test user
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := (&ustore.User{}).Erase(context.Background(), mt, u.ID); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("TEST USER ERASED\n")
}

func TestClientGet(t *testing.T) {
	// Create a user for the test
	u, err := createTestUser("osm1608@gmail.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	// Add a client
	c := &ustore.Client{
		Name:              "Test Client",
		NotificationToken: "notif token",
		Os:                "ios",
		Sdk:               "12",
	}

	path := fmt.Sprintf("/users/%s/clients", u.ID)

	r, err := postMsg(u.Username, "Pass123+Q", NewMessageSim(c), path, http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	// Get the added client
	app.Router.HandleFunc("/users/{0}/clients/{1}", app.ApiAuthenticate(ustore.NewClient, ApiGet, true)).Methods(http.MethodGet)
	cidString := r.Data.(map[string]interface{})["id"].(string)

	url := fmt.Sprintf("%s/%s", path, cidString)
	r, err = getRequest(url, "osm1608@gmail.com", "Pass123+Q", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	gotID := r.Data.(map[string]interface{})["id"].(string)
	if gotID != cidString {
		t.Errorf("expected id= %s, got %s\n", cidString, gotID)
		return
	}

	// Erase the test user
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := (&ustore.User{}).Erase(context.Background(), mt, u.ID); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("TEST USER & CLIENT ERASED\n\n")

	fmt.Printf("GOT CLIENT\n")
}
