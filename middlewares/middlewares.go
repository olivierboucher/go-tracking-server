package middlewares

import (
  "bytes"
  "net/http"
  "log"

  "github.com/xeipuuv/gojsonschema"

  "github.com/OlivierBoucher/go-tracking-server/ctx"
)
//AuthHandler middleware
//This handler handles auth based on the assertion that the request is valid JSON
//Verifies for access, blocks handlers chain if access denied
func AuthHandler(next *ctx.Handler) *ctx.Handler {
  return &ctx.Handler{next.Context, func(c *ctx.Context, w http.ResponseWriter, r *http.Request){
    //TODO: This can change for body instead ?
    token := r.Header.Get("Tracking-Token")
    //Bad request token empty or not present
    if token == "" {
      http.Error(w, http.StatusText(400), 400)
      return
    }

    authorized, err := c.AuthDb.IsTokenAuthorized(token)

    // Internal server error TODO: Handle this
    if err != nil {
      log.Print(err.Error())
      http.Error(w, http.StatusText(500), 500)
      return
    }
    // Unauthorized
    if !authorized {
      http.Error(w, http.StatusText(401), 401)
      return
    }

    next.ServeHTTP(w, r)
  }}
}
//EnforceJSONHandler middleware
//This handler can handle raw requests
//This handler checks for detected content type as well as content-type header
func EnforceJSONHandler(next *ctx.Handler) *ctx.Handler {
    return &ctx.Handler{next.Context, func(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
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
        log.Printf("Invalid mime type. Got %s", mimeType)
        http.Error(w, http.StatusText(415), 415)
        return
      }
      if contentType != "application/json; charset=UTF-8" {
        log.Printf("Invalid content type. Got %s", contentType)
        http.Error(w, http.StatusText(415), 415)
        return
      }
      next.ServeHTTP(w, r);
  }}
}
//ValidateEventTrackingPayloadHandler validates that the payload has a valid JSON Schema
func ValidateEventTrackingPayloadHandler(next *ctx.Handler) *ctx.Handler {
    return &ctx.Handler{next.Context, func(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
      //Validate the payload
      buf := new(bytes.Buffer)
      buf.ReadFrom(r.Body)
      requestLoader := gojsonschema.NewStringLoader(string(buf.Bytes()))

      result, err  := c.JSONTrackingEventValidator.Schema.Validate(requestLoader)
      if err != nil {
          //TODO: Handle the error
          http.Error(w, http.StatusText(500), 500)
          return
      }

      if ! result.Valid() {
        http.Error(w, http.StatusText(400), 400)
        return
      }
      next.ServeHTTP(w, r);
    }}
}
