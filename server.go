package main

import (
  "log"
  "net/http"
  _"database/sql"
  "github.com/OlivierBoucher/go-tracking-server/routes"
  "github.com/OlivierBoucher/go-tracking-server/ctx"
  _"github.com/OlivierBoucher/go-tracking-server/datastores"
)
func main() {
  /*authDb, err := sql.Open("mysql", "")
  if err != nil {
        log.Fatalf("Error on initializing database connection: %s", err.Error())
  }
  defer authDb.Close()*/

  context := &ctx.Context{/*AuthDb: datastores.NewAuthInstance(authDb)*/}

  log.Fatal(http.ListenAndServe(":1337", routes.Handlers(context)))
}
