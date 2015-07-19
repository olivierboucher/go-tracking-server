package middlewares

import (
  "bytes"
  "log"
  "net/http"
)
//Authentication middleware
//This handler handles auth based on the assertion that the request is valid JSON
func authHandler(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    //Logic for auth goes here
    //TODO : Not sure what sql database I'll be using now.. Let's just authorize everything
    next.ServeHTTP(w, r)
  })
}
//Enforce JSON middleware
//This handler can handle raw requests
func enforceJSONHandler(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    //Ensure that there is a body
    if r.ContentLength == 0 {
      http.Error(w, http.StatusText(400), 400)
      return
    }
    //Ensure that its json
    buf := new(bytes.Buffer)
    buf.ReadFrom(r.Body)
    if http.DetectContentType(buf.Bytes()) != "application/json; charset=utf8" {
      http.Error(w, http.StatusText(415), 415)
      return
    }
    next.ServeHTTP(w, r);
  })
}
