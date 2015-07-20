package middlewares_test

import (
  "net/http"
  "net/http/httptest"
  "testing"
  "io"
  "fmt"
  "strings"
  "bytes"
  "log"

  "github.com/gorilla/mux"
  "github.com/DATA-DOG/go-sqlmock"

  "github.com/OlivierBoucher/go-tracking-server/middlewares"
  "github.com/OlivierBoucher/go-tracking-server/ctx"
  "github.com/OlivierBoucher/go-tracking-server/datastores"
)

var (
  server *httptest.Server
  reader io.Reader
  router *mux.Router
  testEnforceJSONUrl string
  testAuthUrl string
)

func init() {
  mockedDb, err := sqlmock.New()
  if err != nil {
        //t.Errorf("An error '%s' was not expected when opening a stub database connection", err)
  }
  columns := []string{"exists"}
  sqlmock.ExpectQuery("SELECT EXISTS\\(SELECT id FROM (.+) WHERE token = (.+)\\)").
        WithArgs("api_tokens", "123").
        WillReturnRows(sqlmock.NewRows(columns).AddRow(false))

  sqlmock.ExpectQuery("SELECT EXISTS\\(SELECT id FROM (.+) WHERE token = (.+)\\)").
        WithArgs("api_tokens", "456").
        WillReturnRows(sqlmock.NewRows(columns).AddRow(true))

  router = mux.NewRouter()
  context := &ctx.Context{AuthDb: datastores.NewAuthInstance(mockedDb)}
  testHandler := &ctx.Handler{context, testHandle}

  router.Handle("/testEnforceJSON", middlewares.EnforceJSONHandler(testHandler))
  router.Handle("/testAuth", middlewares.AuthHandler(testHandler))

  server = httptest.NewServer(router)

  testEnforceJSONUrl = fmt.Sprintf("%s/testEnforceJSON", server.URL)
  testAuthUrl = fmt.Sprintf("%s/testAuth", server.URL)
}

func testHandle(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("MIDDLEWARE PASSED TO NEXT HANDLER"))
}

func TestAuthHandler(t *testing.T) {
  //TESTING WITHOUT HEADER
  reader = strings.NewReader("");
  request, err := http.NewRequest("POST", testAuthUrl, reader)

  res, err := http.DefaultClient.Do(request)

  if err != nil  {
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

  if err != nil  {
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

  if err != nil  {
    t.Error(err)
  }
  buf := new(bytes.Buffer)
  buf.ReadFrom(res.Body)

  resBody := string(buf.Bytes()[:]);

  if resBody != "MIDDLEWARE PASSED TO NEXT HANDLER" {
    t.Errorf("Request should be passed to next handler on valid JSON. Got msg: %s", resBody)
  }
  log.Print("Authorized from database passed")
}

func TestEnforceJSONHandler(t *testing.T) {
  //TESTING EMPTY PAYLOAD
  reader = strings.NewReader("");
  request, err := http.NewRequest("POST", testEnforceJSONUrl, reader)

  res, err := http.DefaultClient.Do(request)

  if err != nil  {
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

  if err != nil  {
    t.Error(err)
  }

  if res.StatusCode != 415 {
    t.Errorf("Wrong media type expected on anything else than JSON got status %d and content-type %s", res.StatusCode, res.Header.Get("Content-Type"))
  }
  log.Print("Wrong content type conditional passed")

  // TESTING JSON PASS
  reader = bytes.NewBuffer([]byte(`{"testing": "is cool"}`))
  request, err = http.NewRequest("POST", testEnforceJSONUrl, reader)
  request.Header.Set("Content-Type", "application/json; charset=utf-8")

  res, err = http.DefaultClient.Do(request)

  if err != nil  {
    t.Error(err)
  }
  buf := new(bytes.Buffer)
  buf.ReadFrom(res.Body)

  resBody := string(buf.Bytes()[:]);

  if resBody != "MIDDLEWARE PASSED TO NEXT HANDLER" {
    t.Errorf("Request should be passed to next handler on valid JSON. Got msg: %s", resBody)
  }
  log.Print("Valid payload conditional passed")
}
