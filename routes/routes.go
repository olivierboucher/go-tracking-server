package routes

import (
  "net/http"
  "github.com/gorilla/mux"
  "github.com/gorilla/websocket"
  "github.com/OlivierBoucher/go-tracking-server/middlewares"
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
//From here we have a valid authentified json request
func handleTrack(w http.ResponseWriter, r *http.Request) {
  //TODO: Handle the json payload
}
//Handler for /connected route
//Allows websocket connection
func handleConnected(w http.ResponseWriter, r *http.Request) {
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
func Handlers() *mux.Router {
  r := mux.NewRouter()
  //Each supported route is being added to the router
  trackHandler := http.HandlerFunc(handleTrack)
  r.Handle("/track", middlewares.EnforceJSONHandler(middlewares.AuthHandler(trackHandler))).Methods("POST", "GET")
  connectedHandler := http.HandlerFunc(handleConnected)
  r.Handle("/connected", connectedHandler)

  return r
}
