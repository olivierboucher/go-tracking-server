package main

import (
  "log"
  "net/http"
  "database/sql"
  "github.com/OlivierBoucher/go-tracking-server/routes"
  "github.com/OlivierBoucher/go-tracking-server/ctx"
  "github.com/OlivierBoucher/go-tracking-server/datastores"
)
func main() {
  authDb, err := sql.Open("mysql", "")
  if err != nil {
        log.Fatalf("Error on initializing database connection: %s", err.Error())
  }

  context := &ctx.Context{AuthDb: datastores.NewAuthInstance(authDb)}

  log.Fatal(http.ListenAndServe(":1337", routes.Handlers(context)))
}
