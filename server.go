package main

import (
  "log"
  "net/http"
  "github.com/OlivierBoucher/go-tracking-server/routes"
)

func main() {
  log.Fatal(http.ListenAndServe(":1337", routes.Handlers()))
}
