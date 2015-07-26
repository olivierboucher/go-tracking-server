package main

import (
  "flag"
  "log"
  "net/http"
  "database/sql"
  "strings"

  "github.com/streadway/amqp"
  "github.com/gocql/gocql"

  "github.com/OlivierBoucher/go-tracking-server/routes"
  "github.com/OlivierBoucher/go-tracking-server/ctx"
  "github.com/OlivierBoucher/go-tracking-server/datastores"
  "github.com/OlivierBoucher/go-tracking-server/queues"
  "github.com/OlivierBoucher/go-tracking-server/validators"
  "github.com/OlivierBoucher/go-tracking-server/utilities"
  "github.com/OlivierBoucher/go-tracking-server/processor"
)



func main() {
  processorFlagPtr := flag.Bool("processor", false, "by default it starts an http server instance, use this flag to start a processor instead")
  envFlagPtr := flag.String("env", "DEV", "by default the environnement is set to DEV, use PROD for production")
  flag.Parse()

  config, err := utilities.LoadJSONConfig()
  if err != nil {
    log.Fatalf("FATAL ERROR: reading json config: %s", err.Error())
  }

  queueConn, err := amqp.Dial(config.QueueConnectionUrl)
  if err != nil {
    log.Fatalf("FATAL ERROR: initializing persistent queue connection: %s", err.Error())
  }
  defer queueConn.Close()

  trackingValidator, err := validators.NewJSONEventTrackingValidator()
  if err != nil {
    log.Fatalf("FATAL ERROR: initializing tracking validator: %s", err.Error())
  }

  if *processorFlagPtr {
    cluster := gocql.NewCluster(strings.Split(config.StorageDbParams.ClusterUrls, "|")...)
    cluster.Keyspace = config.StorageDbParams.Keyspace
    cluster.Authenticator = gocql.PasswordAuthenticator{
      Username: config.StorageDbParams.Username,
      Password: config.StorageDbParams.Password,
    }
    storageConn, err := datastores.NewStorageInstance(cluster)
    if err != nil{
      log.Fatalf("FATAL ERROR: initializing storage db connection: %s", err.Error())
    }

    context := ctx.NewContext(
      nil,
      storageConn,
      queues.NewRabbitMQConnection(queueConn),
      trackingValidator,
      *envFlagPtr)

    startProcessingServer(context)
  } else {
    authDb, err := sql.Open("mysql", config.AuthDbConnectionString)
    if err != nil {
      log.Fatalf("FATAL ERROR: initializing database connection: %s", err.Error())
    }
    defer authDb.Close()

    context := ctx.NewContext(
      datastores.NewAuthInstance(authDb),
      nil,
      queues.NewRabbitMQConnection(queueConn),
      trackingValidator,
      *envFlagPtr)

    startTrackingServer(context)
  }
}

func startTrackingServer(context *ctx.Context) {
  context.Logger.Fatalf("FATAL ERROR: from server: %+v", http.ListenAndServe(":1337", routes.Handlers(context)))
}

func startProcessingServer(context *ctx.Context) {
  ch, err := context.Queue.Channel()
  if err != nil {
    context.Logger.Fatalf("FATAL ERROR: from openning channel: %s", err.Error())
  }
  defer ch.Close()
  
  msgs, err := context.Queue.ConsumeQueueWithChannel("tracking-queue", ch)
  if err != nil {
    context.Logger.Fatalf("FATAL ERROR: from queue consuming: %s", err.Error())
  }

  forever := make(chan bool)

  go func() {
    for m := range msgs {
      context.Logger.Infof("Recieved a message: %s", m.MessageId)
      processor.ProcessMessage(context, &m)
    }
  }()

  context.Logger.Info("Waiting for messages...")
  <-forever
}
