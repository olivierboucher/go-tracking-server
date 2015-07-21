package main

import (
  "log"
  "net/http"
  "database/sql"

  "github.com/streadway/amqp"

  "github.com/OlivierBoucher/go-tracking-server/routes"
  "github.com/OlivierBoucher/go-tracking-server/ctx"
  "github.com/OlivierBoucher/go-tracking-server/datastores"
  "github.com/OlivierBoucher/go-tracking-server/queues"
  "github.com/OlivierBoucher/go-tracking-server/validators"
)
func main() {
  authDb, err := sql.Open("mysql", "")
  if err != nil {
    log.Fatalf("Error on initializing database connection: %s", err.Error())
  }
  defer authDb.Close()

  queueConn, err := amqp.Dial("")
  if err != nil {
    log.Fatalf("Error on initializing persistent queue connection: %s", err.Error())
  }
  defer queueConn.Close()

  context := &ctx.Context{
    AuthDb: datastores.NewAuthInstance(authDb),
    Queue:queues.NewRabbitMQConnection(queueConn),
    JSONTrackingEventValidator: validators.NewJSONEventTrackingValidator(),
  }

  log.Fatal(http.ListenAndServe(":1337", routes.Handlers(context)))
}
