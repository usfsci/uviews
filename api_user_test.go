package uviews

import (
	"bytes"
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
}

func TestApiUserGet(t *testing.T) {
	app.Router.HandleFunc("/users/{0}", app.ApiAuthenticate(ustore.NewUser, ApiGet)).Methods(http.MethodGet)

	r, err := getRequest("/users/EeylSVGC-FehrAIQrEAmUw", "support@useful-science.com", "Pass123+Q", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", r)

	r, err = getRequest("/users/EeylSWwsps-jpgIQrEAmUw", "support@useful-science.com", "Pass123+Q", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", r)
}

func TestApiAdd(t *testing.T) {
	msg := &ustore.Client{
		Name:              "Test Client 6",
		NotificationToken: "notif token 6",
		Os:                "ios",
		Sdk:               "12",
	}

	//EeykV5h1C92EEQIQrEAmUw
	//ID:EeylJ0WC0JGKYQIQrEAmUw
	r, err := postMsg("support@useful-science.com", "Pass123+Q", &msg, "/users/EeylJ0WC0JGKYQIQrEAmUw/clients", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", r)
}

func TestApiList(t *testing.T) {
	r, err := getRequest("/users/EeykV5h1C92EEQIQrEAmUw/clients", "admin@useful-science.com", "Pass123+Q", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", r)
}

func TestApiUserList(t *testing.T) {
	app.Router.HandleFunc("/users", app.ApiAuthenticate(ustore.NewUser, ApiList)).Methods(http.MethodGet)
	r, err := getRequest("/users", "admin@useful-science.com", "Pass123+Q", http.StatusOK, false)
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%+v\n", r)
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
