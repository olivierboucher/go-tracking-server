package middlewares

import (
  "bytes"
  "net/http"
  "log"
)
//AuthHandler middleware
//This handler handles auth based on the assertion that the request is valid JSON
func AuthHandler(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    //Logic for auth goes here
    //TODO : Not sure what sql database I'll be using now.. Let's just authorize everything
    next.ServeHTTP(w, r)
  })
}
//EnforceJSONHandler middleware
//This handler can handle raw requests
func EnforceJSONHandler(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    //Ensure that there is a body
    if r.ContentLength == 0 {
      http.Error(w, http.StatusText(400), 400)
      return
    }
    //Ensure that its json
    buf := new(bytes.Buffer)
    buf.ReadFrom(r.Body)
    mimeType := http.DetectContentType(buf.Bytes());
    contentType := r.Header.Get("Content-Type")

    if mimeType != "text/plain; charset=utf-8" {
      //log.Printf("ERR : Wrong mime type : %s", mimeType)
      http.Error(w, http.StatusText(415), 415)
      return
    }
    if contentType != "application/json; charset=utf-8" {
      //log.Printf("ERR : Wrong content type : %s", contentType)
      http.Error(w, http.StatusText(415), 415)
      return
    }
    next.ServeHTTP(w, r);
  })
}
