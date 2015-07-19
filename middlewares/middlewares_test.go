package middlewares_test

import (
  "tracking-server/middlewares"
  "net/http"
  "net/http/httptest"
  "testing"
  "io"
  "github.com/gorilla/mux"
  "fmt"
  "strings"
  "bytes"
)

var (
  server *httptest.Server
  reader io.Reader
  router *mux.Router
  testEnforceJSONUrl string
  testAuthUrl string
)

func init() {
  router = mux.NewRouter()

  router.Handle("/testEnforceJSON", enforceJSONHandler(TestHandle))
  router.Handle("/testAuth", authHandler(TestHandle))

  server = httptest.NewServer(router)

  testEnforceJSONUrl = fmt.Sprintf("%s/testEnforceJSON", server.URL)
  testAuthUrl = fmt.Sprintf("%s/testAuth", server.URL)
}

func TestHandle(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("MIDDLEWARE PASSED TO NEXT HANDLER"))
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
  // TESTING WRONG CONTENT TYPE
  reader = strings.NewReader("This is a test")
  request, err = http.NewRequest("POST", testEnforceJSONUrl, reader)
  request.Header.Add("Content-Type", "text/plain")

  res, err = http.DefaultClient.Do(request)

  if err != nil  {
    t.Error(err)
  }

  if res.StatusCode != 415 {
    t.Errorf("Wrong media type expected on anything else than JSON got status %d and content-type %s", res.StatusCode, res.Header.Get("Content-Type"))
  }
  // TESTING JSON PASS
  reader = strings.NewReader(`{"testing": "is cool"}`)
  request, err = http.NewRequest("POST", testEnforceJSONUrl, reader)

  res, err = http.DefaultClient.Do(request)

  if err != nil  {
    t.Error(err)
  }
  buf := new(bytes.Buffer)
  buf.ReadFrom(res.Body)

  if buf.Bytes() != "MIDDLEWARE PASSED TO NEXT HANDLER" {
    t.Error("Request should be passed to next handler on valid JSON")
  }
}
