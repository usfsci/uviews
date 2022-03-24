package uviews

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/usfsci/ustore"
)

func TestApiUserAdd(t *testing.T) {
	app.Router.HandleFunc("/users", app.ApiBypassAuthentication(ustore.NewUser, ApiAdd)).Methods(http.MethodPost)

	u := &ustore.User{
		Username:      "osm1608@gmail.com",
		Password:      []byte("Pass123+Q"),
		AcceptedTerms: true,
	}

	msg := map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      u,
	}

	r, err := postMsg("", "", &msg, "/users", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", r)

	fmt.Printf("ADDED USER\n")

	// Verify that the user was created

	// Get the user id from the response
	idStr := r.Data.(map[string]interface{})["id"].(string)
	id, err := ustore.SIDFromString(idStr)
	if err != nil {
		t.Error(err)
		return
	}

	// Get the user
	u1 := &ustore.User{}

	if err := u1.Get(context.Background(), &ustore.Filter{}, id); err != nil {
		t.Error(err)
		return
	}

	// Validate user
	if u1.Username != "osm1608@gmail.com" {
		t.Errorf("Expected osm1608@gmail.com, got %s\n", u1.Username)
		return
	}

	fmt.Printf("RETRIEVED TEST USER\n")

	// Erase test user
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := (&ustore.User{}).Erase(context.Background(), mt, id); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("TEST USER ERASED\n")
}

func TestApiUserGet(t *testing.T) {
	// Register path
	app.Router.HandleFunc("/users/{0}", app.ApiAuthenticate(ustore.NewUser, ApiGet, true)).Methods(http.MethodGet)

	// Create User
	u, err := createTestUser("osm1608@gmail.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	// Get it with valid creds
	r, err := getRequest(fmt.Sprintf("/users/%s", u.ID), u.Username, "Pass123+Q", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", r)
	fmt.Printf("GOT TEST USER\n")

	// Get the id from the response to verify op
	idStr := r.Data.(map[string]interface{})["id"].(string)
	id, err := ustore.SIDFromString(idStr)
	if err != nil {
		t.Error(err)
		return
	}

	// Verify that the password is nopt exposed
	pass := r.Data.(map[string]interface{})["password"]
	if pass != nil {
		t.Errorf("expected nil pass, got %v\n", pass)
		return
	}

	u1 := &ustore.User{}
	if err := u1.Get(context.Background(), &ustore.Filter{}, id); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("VALIDATED TEST USER\n")

	// Erase test user
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := (&ustore.User{}).Erase(context.Background(), mt, u1.ID); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("TEST USER ERASED\n")
}

func TestApiUserList(t *testing.T) {
	app.Router.HandleFunc("/users", app.ApiAuthenticate(ustore.NewUser, ApiList, true)).Methods(http.MethodGet)

	unames := []string{"osm1608@gmail.com", "support@useful-science.com", "admin@useful-science.com"}
	users := make([]*ustore.User, len(unames))

	// Create Users
	for i := 0; i < len(unames); i++ {
		var err error
		users[i], err = createTestUser(unames[i], "Pass123+Q")
		if err != nil {
			t.Error(err)
			return
		}
	}

	// Set the last user as super
	if err := (&ustore.User{}).UpdateAuthLevel(context.Background(), 0, users[len(unames)-1].ID); err != nil {
		t.Error(err)
		return
	}

	// List the users
	r, err := getRequest("/users", unames[len(unames)-1], "Pass123+Q", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", r)

	// Verify the user ID's
	data := r.Data.([]interface{})
	for i, m := range data {
		umap := m.(map[string]interface{})
		if umap["username"] != users[i].Username {
			t.Errorf("expected %s, got %s\n", users[i].Username, umap["username"])
			return
		}
		if umap["password"] != nil {
			t.Errorf("expected nil password, got %s\n", umap["password"])
			return
		}
	}

	if len(data) != len(users) {
		t.Errorf("expected len data %d, got %d\n", len(users), len(data))
		return
	}

	// Erase the Test Users
	for _, u := range users {
		if err := (&ustore.User{}).Erase(context.Background(), time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC), u.ID); err != nil {
			t.Error(err)
			return
		}
	}
}

func TestUserUpdateAcceptedNews(t *testing.T) {
	// Register path
	app.Router.HandleFunc("/users/{0}", app.ApiAuthenticate(ustore.NewUser, ApiUpdate, true)).Methods(http.MethodPut)

	// Create User
	u, err := createTestUser("osm1608@gmail.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	an := true
	u1 := &ustore.User{
		Base:         ustore.Base{ModificationTime: time.Now().In(time.UTC)},
		AcceptedNews: &an,
	}

	msg := map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      u1,
	}

	// Update accepted news
	r, err := putMsg("osm1608@gmail.com", "Pass123+Q", msg, fmt.Sprintf("/users/%s", u.ID), http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	// Get the id from the response to verify op
	idStr := r.Data.(map[string]interface{})["id"].(string)
	id, err := ustore.SIDFromString(idStr)
	if err != nil {
		t.Error(err)
		return
	}

	// Get the test user
	u2 := &ustore.User{}
	if err := u2.Get(context.Background(), &ustore.Filter{}, id); err != nil {
		t.Error(err)
		return
	}

	if !*u2.AcceptedNews {
		t.Errorf("expected AcceptedNews true, go %v\n", *u2.AcceptedNews)
		return
	}

}

func TestUserUpdatePassword(t *testing.T) {
	// Register path
	app.Router.HandleFunc("/users/{0}", app.ApiAuthenticate(ustore.NewUser, ApiUpdate, true)).Methods(http.MethodPut)

	// Create User
	u, err := createTestUser("osm1608@gmail.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	newPass := "Pass987+Y"
	u1 := &ustore.User{
		Base:     ustore.Base{ModificationTime: time.Now().In(time.UTC)},
		Password: []byte(newPass),
	}

	msg := map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      u1,
	}

	// Update
	r, err := putMsg("osm1608@gmail.com", "Pass123+Q", msg, fmt.Sprintf("/users/%s", u.ID), http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	// Get the id from the response to verify op
	idStr := r.Data.(map[string]interface{})["id"].(string)
	id, err := ustore.SIDFromString(idStr)
	if err != nil {
		t.Error(err)
		return
	}

	// Get the test user
	u2 := &ustore.User{}
	if err := u2.Get(context.Background(), &ustore.Filter{}, id); err != nil {
		t.Error(err)
		return
	}

	if u2.Authenticate(newPass) != nil {
		t.Errorf("expected newPass\n")
		return
	}

	// Delete test user
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := (&ustore.User{}).Erase(context.Background(), mt, u2.ID); err != nil {
		t.Error(err)
		return
	}
}

func TestUserUpdateUsername(t *testing.T) {
	// Register path
	app.Router.HandleFunc("/users/{0}", app.ApiAuthenticate(ustore.NewUser, ApiUpdate, true)).Methods(http.MethodPut)

	// Create User
	u, err := createTestUser("osm1608@gmail.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	newUsername := "admin@useful-science.com"
	u1 := &ustore.User{
		Base:     ustore.Base{ModificationTime: time.Now().In(time.UTC)},
		Username: newUsername,
	}

	msg := map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      u1,
	}

	// Update
	r, err := putMsg("osm1608@gmail.com", "Pass123+Q", msg, fmt.Sprintf("/users/%s", u.ID), http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	// Get the id from the response to verify op
	idStr := r.Data.(map[string]interface{})["id"].(string)
	id, err := ustore.SIDFromString(idStr)
	if err != nil {
		t.Error(err)
		return
	}

	// Get the test user
	u2 := &ustore.User{}
	if err := u2.Get(context.Background(), &ustore.Filter{}, id); err != nil {
		t.Error(err)
		return
	}

	if newUsername != u2.Username {
		t.Errorf("expected %s, got %s\n", newUsername, u2.Username)
		return
	}

	// Delete test user
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := (&ustore.User{}).Erase(context.Background(), mt, u2.ID); err != nil {
		t.Error(err)
		return
	}
}

func TestUserEmailValidate(t *testing.T) {
	app.Router.HandleFunc("/users/{0}/tokens", app.ApiAuthenticate(ustore.NewRawToken, ApiEmailValidate, false)).Methods(http.MethodPost)
	app.Router.HandleFunc("/users", app.ApiBypassAuthentication(ustore.NewUser, ApiAdd)).Methods(http.MethodPost)

	// Create user
	u := &ustore.User{
		Username:      "osm1608@gmail.com",
		Password:      []byte("Pass123+Q"),
		AcceptedTerms: true,
		CountryCode:   "ES",
	}

	msg := map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      u,
	}

	r, err := postMsg("", "", &msg, "/users", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	if r.Error != nil {
		t.Error(r.Error[0])
		return
	}

	fmt.Printf("%+v\n", r)
	fmt.Printf("token = %s\n", etoken)

	fmt.Printf("ADDED USER\n")

	// Verify that the user was created

	// Get the user id from the response
	idStr := r.Data.(map[string]interface{})["id"].(string)
	id, err := ustore.SIDFromString(idStr)
	if err != nil {
		t.Error(err)
		return
	}

	// Create user
	rt := &ustore.RawToken{
		Token: etoken,
	}

	msg1 := map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      rt,
	}

	_, err = postMsg("osm1608@gmail.com", "Pass123+Q", &msg1, fmt.Sprintf("/users/%s/tokens", id), http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	// Check the email is indeed conf
	u1 := &ustore.User{}
	if err := u1.Get(context.Background(), &ustore.Filter{}, id); err != nil {
		t.Error(err)
		return
	}

	if !u1.EmailConfirmed {
		t.Errorf("expected EmailConfirmed = true, got %v", u1.EmailConfirmed)
		return
	}

	fmt.Printf("CONFIRMED EMAIL\n")

	// Delete test user
	if err := (&ustore.User{}).Erase(context.Background(), time.Now().In(time.UTC), id); err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("DELETED TEST USER\n")
}

func TestUserResetPassword(t *testing.T) {
	app.Router.HandleFunc("/users/{0}/tokens", app.ApiBypassAuthentication(ustore.NewRawToken, ApiList)).Methods(http.MethodGet)
	app.Router.HandleFunc("/users/{0}/tokens", app.ApiBypassAuthentication(ustore.NewRawToken, ApiPasswordReset)).Methods(http.MethodPut)

	const uname = "osm1608@gmail.com"
	const pass = "Pass123+Q"
	// Create test user
	u, err := createTestUser(uname, pass)
	if err != nil {
		t.Error(err)
		return
	}

	url := fmt.Sprintf("/users/%s/tokens", u.ID)

	// Get a Password Reset Token
	r, err := getRequest(url, uname, pass, http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}
	if r.Error != nil {
		t.Errorf("%v\n", r.Error)
		return
	}

	newPass := "ssaP321+Y"
	// Submit token and new Pass
	rt := &ustore.RawToken{
		Token:    etoken,
		Password: newPass,
	}

	msg := map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      rt,
	}

	r, err = putMsg(uname, pass, &msg, url, http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}
	if r.Error != nil {
		t.Errorf("%v\n", r.Error)
		return
	}

	// Check that the Password was reset
	u1 := &ustore.User{}
	if err := u1.Get(context.Background(), &ustore.Filter{}, u.ID); err != nil {
		t.Error(err)
		return
	}

	if err := u1.Authenticate(newPass); err != nil {
		t.Error(err)
		return
	}

	// Erase Test usr
	if err := (&ustore.User{}).Erase(context.Background(), time.Now().In(time.UTC), u.ID); err != nil {
		t.Error(err)
		return
	}
}

func TestUserAuthorization(t *testing.T) {
	app.Router.HandleFunc("/users/{0}", app.ApiAuthenticate(ustore.NewUser, ApiGet, true)).Methods(http.MethodGet)
	app.Router.HandleFunc("/users/{0}", app.ApiAuthenticate(ustore.NewUser, ApiUpdate, true)).Methods(http.MethodPut)

	u0, err := createTestUser("osm1608@gmail.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	u1, err := createTestUser("admin@useful-science.com", "Pass123+Q")
	if err != nil {
		t.Error(err)
		return
	}

	// Attempt GET w/out authorization
	_, err = getRequest(fmt.Sprintf("/users/%s", u0.ID), u1.Username, "Pass123+Q", http.StatusOK, false)
	if err == nil {
		t.Error("expected no authorization error")
		return
	}

	// Attempt list w/out "superuser" auth
	_, err = getRequest("/users", u1.Username, "Pass123+Q", http.StatusOK, false)
	if err == nil {
		t.Error("expected no authorization error")
		return
	}

	// Attempt Update w/out auth
	an := true
	u0.AcceptedNews = &an

	msg := map[string]interface{}{
		"timestamp": time.Now().UnixNano(),
		"data":      u0,
	}

	// Update accepted news
	_, err = putMsg(u0.Username, "Pass123+Q", msg, fmt.Sprintf("/users/%s", u1.ID), http.StatusOK, false)
	if err == nil {
		t.Error("expected no authorization error")
		return
	}

	// Erase test users
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

	if err := (&ustore.User{}).Erase(context.Background(), mt, u0.ID); err != nil {
		t.Error(err)
		return
	}

	if err := (&ustore.User{}).Erase(context.Background(), mt, u1.ID); err != nil {
		t.Error(err)
		return
	}
}

func createTestUser(uname string, password string) (*ustore.User, error) {
	// Create a user for the test
	an := false
	u := &ustore.User{
		Username:      uname,
		Password:      []byte(password),
		AcceptedTerms: true,
		AcceptedNews:  &an,
		CountryCode:   "ES",
	}

	if err := u.Add(context.Background(), ""); err != nil {
		return nil, err
	}

	fmt.Printf("TEST USER CREATED\n")

	// Confirm email
	mt := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)
	u.EmailConfirmed = true
	if err := u.UpdateEmailConfirmed(context.Background(), mt, u.ID); err != nil {
		return nil, err
	}

	fmt.Printf("TEST USER EMAIL MARKED CONFIRMED\n")

	return u, nil
}

func postMsg(
	uname string,
	pass string,
	msg interface{},
	url string,
	expectedCode int,
	checkCodeOnly bool,
) (*Response, error) {

	jmsg, err := json.MarshalIndent(msg, "", "\t")
	if err != nil {
		return nil, err
	}

	fmt.Printf("\n\n**************************\n\n")
	fmt.Printf("POST %s\n", url)
	fmt.Printf("%s\n", string(jmsg))
	fmt.Printf("\n\n**************************\n\n")

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jmsg))
	defer req.Body.Close()
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(uname, pass)

	response := executeReq(req)
	fmt.Printf("Headers= %+v\n", response.HeaderMap)
	fmt.Printf("Response= %+v\n", response.Body)

	gotCode := response.Result().StatusCode
	if gotCode != expectedCode {
		return nil, fmt.Errorf("expected status code %d, got %d", expectedCode, gotCode)
	}

	if checkCodeOnly {
		return nil, nil
	}

	var r Response
	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("json error= %+v", err)
	}

	return &r, nil
}

func putMsg(
	uname string,
	pass string,
	msg interface{},
	url string,
	expectedCode int,
	checkCodeOnly bool,
) (*Response, error) {
	jmsg, err := json.MarshalIndent(msg, "", "\t")
	if err != nil {
		return nil, err
	}

	fmt.Printf("URL= %s\n", url)
	fmt.Printf("MSG= %+v\n", string(jmsg))

	req, _ := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jmsg))
	defer req.Body.Close()
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(uname, pass)

	response := executeReq(req)
	fmt.Printf("Response= %+v\n", response.Body)

	gotCode := response.Result().StatusCode
	if gotCode != expectedCode {
		return nil, fmt.Errorf("expected status code %d, got %d", expectedCode, gotCode)
	}

	if checkCodeOnly {
		return nil, nil
	}

	var r Response
	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("json error= %+v", err)
	}

	return &r, nil
}

func getRequest(url string, uname string, pass string, expectedCode int, checkCodeOnly bool) (*Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	//defer req.Body.Close()

	//req.Header.Set("Content-Type", "application/json")
	fmt.Printf("URL= %s\n", url)
	req.SetBasicAuth(uname, pass)

	response := executeReq(req)
	fmt.Printf("Response= %+v\n", response.Body)

	gotCode := response.Result().StatusCode
	if gotCode != expectedCode {
		return nil, fmt.Errorf("expected status code %d, got %d", expectedCode, gotCode)
	}

	if checkCodeOnly {
		return nil, nil
	}

	var r Response
	err = json.NewDecoder(response.Body).Decode(&r)
	if err != nil {
		return nil, fmt.Errorf("json error= %+v", err)
	}

	return &r, nil
}

func executeReq(req *http.Request) *httptest.ResponseRecorder {
	r := httptest.NewRecorder()
	app.Router.ServeHTTP(r, req)
	return r
}
