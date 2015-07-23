package main

import (
  "log"
  "net/http"
  "database/sql"
  "io/ioutil"
  "encoding/json"

  "github.com/streadway/amqp"

  "github.com/OlivierBoucher/go-tracking-server/routes"
  "github.com/OlivierBoucher/go-tracking-server/ctx"
  "github.com/OlivierBoucher/go-tracking-server/datastores"
  "github.com/OlivierBoucher/go-tracking-server/queues"
  "github.com/OlivierBoucher/go-tracking-server/validators"
)

type srvConfiguration struct {
  AuthDbConnectionString string `json:"authDb"`
  QueueConnectionUrl string `json:"queueUrl"`
}

func main() {
  config := loadJSONConfig()

  authDb, err := sql.Open("mysql", config.AuthDbConnectionString)
  if err != nil {
    log.Fatalf("FATAL ERROR: initializing database connection: %s", err.Error())
  }
  defer authDb.Close()

  queueConn, err := amqp.Dial(config.QueueConnectionUrl)
  if err != nil {
    log.Fatalf("FATAL ERROR: initializing persistent queue connection: %s", err.Error())
  }
  defer queueConn.Close()

  trackingValidator, err := validators.NewJSONEventTrackingValidator()
  if err != nil {
    log.Fatalf("FATAL ERROR: initializing tracking validator: %s", err.Error())
  }

  context := ctx.NewContext(
    datastores.NewAuthInstance(authDb),
    queues.NewRabbitMQConnection(queueConn),
    trackingValidator,
    "DEVELOPMENT")

  context.Logger.Fatalf("FATAL ERROR: from server: %+v", http.ListenAndServe(":1337", routes.Handlers(context)))
}

func loadJSONConfig() *srvConfiguration {
  file, err := ioutil.ReadFile("./config.json")
  if err != nil {
    log.Fatalf("FATAL ERROR: reading json config: %s", err.Error())
  }

  var config srvConfiguration
  json.Unmarshal(file, &config)

  return &config
}
