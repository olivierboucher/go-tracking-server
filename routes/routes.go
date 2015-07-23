package routes

import (
  "net"
  "net/http"
  "time"
  "log"

  "github.com/gorilla/mux"
  "github.com/gorilla/websocket"
  "github.com/xeipuuv/gojsonschema"

  "github.com/OlivierBoucher/go-tracking-server/middlewares"
  "github.com/OlivierBoucher/go-tracking-server/ctx"
)

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
      //TODO: Perform checks on the request and deny/accept
      return true
    },
}
//Handler for /track route
//From here we have a valid authentified json request with correct schema
func handleTrack(c *ctx.Context, p []byte, w http.ResponseWriter, r *http.Request) {
  err := c.Queue.PublishEventsTrackingTask(p)
  if err != nil {
    //TODO : Could we handle this a little better?
    c.Logger.Errorf("Error publishing to queue: %s\nPayload: %s", err.Error(), string(p))
    http.Error(w, http.StatusText(500), 500)
  }
}
//Handler for /connected route
//Allows websocket connection
//From here we have an authentified request
func handleConnected(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
    const pongWait time.Duration = 10 //seconds

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        //TODO : Could we handle this a little better?
        http.Error(w, http.StatusText(500), 500)
        return
    }
    //We defer the closing of the connection
    defer conn.Close()
    //Custom settings for timeouts
    conn.SetReadDeadline(time.Now().Add(pongWait))
    conn.SetPongHandler(func(string) error {
      conn.SetReadDeadline(time.Now().Add(pongWait))
      return nil
    })

    for {
        messageType, payload, err := conn.ReadMessage()
        if err != nil {
          if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
              //Timeout error
              return
          }
          //Unhandled error
          //Maybe log something?
          return
        }

        if messageType != websocket.TextMessage {
            //The message type is not handled
            return
        }
        //We need to validate the payload and then send it
        payloadLoader := gojsonschema.NewStringLoader(string(payload))
        result, err  := c.JSONTrackingEventValidator.Schema.Validate(payloadLoader)
        if err != nil {
            log.Printf("Json validation error: %s", err.Error())
            //TODO : Send an ack with the error
        }

        if ! result.Valid() {
          log.Printf("Invalid payload")
          //TODO : Send an ack with the error
        }
        err = c.Queue.PublishEventsTrackingTask(payload)
        if err != nil {
          //TODO : Could we handle this a little better?
          http.Error(w, http.StatusText(500), 500)
        }
        //TODO: ack the request
    }
}
//Handlers  Returns a mux router containing all handlers for all routes
func Handlers(c *ctx.Context) *mux.Router {
  r := mux.NewRouter()
  //Each supported route is being added to the router
  trackHandler := ctx.NewFinalHandler(c, []byte(""), handleTrack)
  r.Handle("/track", middlewares.EnforceJSONHandler(middlewares.AuthHandler(middlewares.ValidateEventTrackingPayloadHandler(trackHandler)))).Methods("POST", "GET")
  connectedHandler := &ctx.Handler{c, handleConnected}
  r.Handle("/connected", middlewares.AuthHandler(connectedHandler)).Methods("GET")

  return r
}
