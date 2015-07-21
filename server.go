package main

import (
  "log"
  "net/http"
  "database/sql"
  "io/ioutil"
  "encoding/json"
  "os"

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
    log.Fatalf("Error on initializing database connection: %s", err.Error())
    os.Exit(1)
  }
  defer authDb.Close()

  queueConn, err := amqp.Dial(config.QueueConnectionUrl)
  if err != nil {
    log.Fatalf("Error on initializing persistent queue connection: %s", err.Error())
    os.Exit(1)
  }
  defer queueConn.Close()

  context := ctx.NewContext(
    datastores.NewAuthInstance(authDb),
    queues.NewRabbitMQConnection(queueConn),
    validators.NewJSONEventTrackingValidator())

  log.Fatal(http.ListenAndServe(":1337", routes.Handlers(context)))
}

func loadJSONConfig() *srvConfiguration {
  file, err := ioutil.ReadFile("./config.json")
  if err != nil {
    log.Fatalf("Error on reading json config: %s", err.Error())
    os.Exit(1)
  }

  var config srvConfiguration
  json.Unmarshal(file, &config)

  return &config
}
