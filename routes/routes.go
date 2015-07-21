package routes

import (
  "net/http"
  "bytes"

  "github.com/gorilla/mux"
  "github.com/gorilla/websocket"

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
func handleTrack(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
  buf := new(bytes.Buffer)
  buf.ReadFrom(r.Body)
  c.Queue.PublishEventsTrackingTask(buf.Bytes())
}
//Handler for /connected route
//Allows websocket connection
func handleConnected(c *ctx.Context, w http.ResponseWriter, r *http.Request) {
  conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        //TODO: Handle the error
        panic(err)
    }
    for {
        messageType, p, err := conn.ReadMessage()
        if err != nil {
            return
        }
        //TODO: Define a protocol

        err = conn.WriteMessage(messageType, p);
        if  err != nil {
            return
        }
    }
}
//Handlers - Returns a mux router containing all handlers for all routes
func Handlers(c *ctx.Context) *mux.Router {
  r := mux.NewRouter()
  //Each supported route is being added to the router
  trackHandler := &ctx.Handler{c, handleTrack}
  r.Handle("/track", middlewares.EnforceJSONHandler(middlewares.AuthHandler(middlewares.ValidateEventTrackingPayloadHandler(trackHandler)))).Methods("POST", "GET")
  connectedHandler := &ctx.Handler{c, handleConnected}
  r.Handle("/connected", connectedHandler)

  return r
}
