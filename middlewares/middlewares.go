package middlewares

import (
  "net"
  "net/http"
  "io/ioutil"

  "github.com/xeipuuv/gojsonschema"

  "github.com/OlivierBoucher/go-tracking-server/ctx"
)

func getIP(r *http.Request) string {
    if ipProxy := r.Header.Get("X-FORWARDED-FOR"); len(ipProxy) > 0 {
        return ipProxy
    }
    ip, _, _ := net.SplitHostPort(r.RemoteAddr)
    return ip
}
//AuthHandler middleware
//This handler handles auth based on the assertion that the request is valid JSON
//Verifies for access, blocks handlers chain if access denied
func AuthHandler(next *ctx.Handler) *ctx.Handler {
  return ctx.NewHandler(next.Context, func(c *ctx.Context, w http.ResponseWriter, r *http.Request){
    //TODO: This can change for body instead ?
    token := r.Header.Get("Tracking-Token")
    //Bad request token empty or not present
    if token == "" {
      c.Logger.Infof("%s : No token header", getIP(r))
      token = r.URL.Query().Get("key")
      if token == ""{
        c.Logger.Infof("%s : No token query parameter", getIP(r))
        http.Error(w, http.StatusText(400), 400)
        return
      }
    }

    authorized, err := c.AuthDb.IsTokenAuthorized(token)

    // Internal server error TODO: Handle this
    if err != nil {
      c.Logger.Errorf("%s : Error while authorizing: %s", getIP(r), err.Error())
      http.Error(w, http.StatusText(500), 500)
      return
    }
    // Unauthorized
    if !authorized {
      c.Logger.Warnf("%s : Unauthorized: %s", getIP(r), token)
      http.Error(w, http.StatusText(401), 401)
      return
    }
    c.Logger.Infof("%s : Authorized", getIP(r))
    next.ServeHTTP(w, r)
  })
}
//EnforceJSONHandler middleware
//This handler can handle raw requests
//This handler checks for detected content type as well as content-type header
func EnforceJSONHandler(next *ctx.Handler) *ctx.Handler {
    return ctx.NewHandler(next.Context, func(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
      //Ensure that there is a body
      if r.ContentLength == 0 {
        c.Logger.Infof("%s : Recieved empty payload", getIP(r))
        http.Error(w, http.StatusText(400), 400)
        return
      }
      //Ensure that its json
      contentType := r.Header.Get("Content-Type")
      if contentType != "application/json; charset=UTF-8" {
        c.Logger.Infof("%s : Invalid content type. Got %s", getIP(r), contentType)
        http.Error(w, http.StatusText(415), 415)
        return
      }
      next.ServeHTTP(w, r);
  })
}
//ValidateEventTrackingPayloadHandler validates that the payload has a valid JSON Schema
//Uses a FinalHandler as next because it consumes the request's body
func ValidateEventTrackingPayloadHandler(next *ctx.FinalHandler) *ctx.Handler {
    return ctx.NewHandler(next.Context, func(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
      //Validate the payload
      body, err := ioutil.ReadAll(r.Body)
      if err != nil {
        c.Logger.Errorf("%s : Error reading body: %s", getIP(r), err.Error())
        http.Error(w, http.StatusText(500), 500)
        return
      }
      next.Payload = body
      requestLoader := gojsonschema.NewStringLoader(string(body))

      result, err  := c.JSONTrackingEventValidator.Schema.Validate(requestLoader)
      if err != nil {
          c.Logger.Infof("%s : Json validation error: %s", getIP(r), err.Error())
          http.Error(w, http.StatusText(400), 400)
          return
      }

      if ! result.Valid() {
        c.Logger.Infof("%s : Payload is not valid: %+v", getIP(r), result.Errors())
        http.Error(w, http.StatusText(400), 400)
        return
      }
      next.ServeHTTP(w, r);
    })
}
