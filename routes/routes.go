package routes

import (
  "net/http"
  "github.com/gorilla/mux"
  "tracking-server/middlewares"
)
//Handler for /track route
//From here we have a valid authentified json request
func handleTrack(w http.ResponseWriter, r *http.Request) {
  //TODO: Handle the json payload
}
//Handlers - Returns a mux router containing all handlers for all routes
func Handlers() *mux.Router {
  r := mux.NewRouter()
  //Each supported route is being added to the router
  trackHandler := http.HandlerFunc(handleTrack)
  r.Handle("/track", middlewares.EnforceJSONHandler(middlewares.AuthHandler(trackHandler))).Methods("POST", "GET")

  return r
}
