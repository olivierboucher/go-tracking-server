package middlewares_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gorilla/mux"

	"github.com/OlivierBoucher/go-tracking-server/ctx"
	"github.com/OlivierBoucher/go-tracking-server/datastores"
	"github.com/OlivierBoucher/go-tracking-server/middlewares"
	"github.com/OlivierBoucher/go-tracking-server/validators"
)

var (
	server                                     *httptest.Server
	reader                                     io.Reader
	router                                     *mux.Router
	testEnforceJSONUrl                         string
	testAuthUrl                                string
	testValidateEventTrackingPayloadHandlerUrl string
)

func init() {
	mockedDb, err := sqlmock.New()
	if err != nil {
		log.Fatalf("An error: %s, was not expected when opening a stub database connection", err.Error())
	}

	columns := []string{"exists"}
	sqlmock.ExpectQuery("SELECT EXISTS\\(SELECT id FROM api_tokens WHERE token = (.+)\\)").
		WithArgs("123").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(false))

	sqlmock.ExpectQuery("SELECT EXISTS\\(SELECT id FROM api_tokens WHERE token = (.+)\\)").
		WithArgs("456").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(true))

	trackingValidator, err := validators.NewJSONEventTrackingValidator()
	if err != nil {
		log.Fatalf("An error: %s, was not expected when initializing tracking validator", err.Error())
	}

	router = mux.NewRouter()
	context := ctx.NewContext(datastores.NewAuthInstance(mockedDb), nil, nil, trackingValidator, "DEV")
	testHandler := ctx.NewHandler(context, testHandle)
	testFinalHander := ctx.NewFinalHandler(context, []byte(""), testFinalHandle)

	router.Handle("/testEnforceJSON", middlewares.EnforceJSONHandler(testHandler))
	router.Handle("/testAuth", middlewares.AuthHandler(testHandler))
	router.Handle("/testValidateEventTrackingPayloadHandler", middlewares.ValidateEventTrackingPayloadHandler(testFinalHander))

	server = httptest.NewServer(router)

	testEnforceJSONUrl = fmt.Sprintf("%s/testEnforceJSON", server.URL)
	testAuthUrl = fmt.Sprintf("%s/testAuth", server.URL)
	testValidateEventTrackingPayloadHandlerUrl = fmt.Sprintf("%s/testValidateEventTrackingPayloadHandler", server.URL)
}
func testFinalHandle(c *ctx.Context, p []byte, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("MIDDLEWARE PASSED TO NEXT HANDLER"))
}
func testHandle(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("MIDDLEWARE PASSED TO NEXT HANDLER"))
}

func TestAuthHandler(t *testing.T) {
	//TESTING WITHOUT HEADER
	reader = strings.NewReader("")
	request, err := http.NewRequest("POST", testAuthUrl, reader)

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 400 {
		t.Errorf("Bad request expected on missing header, got status %d", res.StatusCode)
	}
	log.Print("Missing header conditional passed")
	// TESTING UNAUTHORIZED
	reader = strings.NewReader("")
	request, err = http.NewRequest("POST", testAuthUrl, reader)
	request.Header.Set("Tracking-Token", "123")

	res, err = http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 401 {
		t.Errorf("Unauthorized expected, got : %d", res.StatusCode)
	}
	log.Print("Unauthorized from database passed")
	// TESTING AUTHORIZED
	reader = strings.NewReader("")
	request, err = http.NewRequest("POST", testAuthUrl, reader)
	request.Header.Set("Tracking-Token", "456")

	res, err = http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)

	resBody := string(buf.Bytes()[:])

	if resBody != "MIDDLEWARE PASSED TO NEXT HANDLER" {
		t.Errorf("Request should be passed to next handler on valid JSON. Got msg: %s", resBody)
	}
	log.Print("Authorized from database passed")
}

func TestEnforceJSONHandler(t *testing.T) {
	//TESTING EMPTY PAYLOAD
	reader = strings.NewReader("")
	request, err := http.NewRequest("POST", testEnforceJSONUrl, reader)

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 400 {
		t.Errorf("Bad request expected on empty payload, got status %d", res.StatusCode)
	}
	log.Print("Empty payload conditional passed")

	// TESTING WRONG CONTENT TYPE
	reader = strings.NewReader("This is a test")
	request, err = http.NewRequest("POST", testEnforceJSONUrl, reader)

	res, err = http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 415 {
		t.Errorf("Wrong media type expected on anything else than JSON got status %d and content-type %s", res.StatusCode, res.Header.Get("Content-Type"))
	}
	log.Print("Wrong content type conditional passed")

	// TESTING JSON PASS
	reader = bytes.NewBuffer([]byte(`{"testing": "is cool"}`))
	request, err = http.NewRequest("POST", testEnforceJSONUrl, reader)
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	res, err = http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)

	resBody := string(buf.Bytes()[:])

	if resBody != "MIDDLEWARE PASSED TO NEXT HANDLER" {
		t.Errorf("Request should be passed to next handler on valid JSON. Got msg: %s", resBody)
	}
	log.Print("Valid payload conditional passed")
}

func TestValidateEventTrackingPayloadHandler(t *testing.T) {
	// TESTING WRONG CONTENT TYPE
	reader = bytes.NewBuffer([]byte(`{"testing": "is cool"}`))
	request, err := http.NewRequest("POST", testValidateEventTrackingPayloadHandlerUrl, reader)

	res, err := http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 400 {
		t.Errorf("Wrong schema passed validation, status code : %d", res.StatusCode)
	}
	log.Print("Invalid json conditional passed")

	//TESTING VALID JSON
	reader = bytes.NewBuffer([]byte(`{"token":"TOKEN","events":[{"name":"TEST","date":"2015-07-26T15:35:49+00:00","properties":[{"name":"PROP1","value":"string value"},{"name":"PROP2","value":"123"},{"name":"PROP3","value":"12.567"}]}]}`))
	request, err = http.NewRequest("POST", testValidateEventTrackingPayloadHandlerUrl, reader)

	res, err = http.DefaultClient.Do(request)

	if err != nil {
		t.Error(err)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)

	resBody := string(buf.Bytes()[:])

	if resBody != "MIDDLEWARE PASSED TO NEXT HANDLER" {
		t.Errorf("Request should be passed to next handler on valid JSON. Got msg: %s", resBody)
	}
	log.Print("Valid conditional passed")
}
