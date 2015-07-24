package main

import (
  "flag"
  "log"
  "net/http"
  "database/sql"

  "github.com/streadway/amqp"

  "github.com/OlivierBoucher/go-tracking-server/routes"
  "github.com/OlivierBoucher/go-tracking-server/ctx"
  "github.com/OlivierBoucher/go-tracking-server/datastores"
  "github.com/OlivierBoucher/go-tracking-server/queues"
  "github.com/OlivierBoucher/go-tracking-server/validators"
  "github.com/OlivierBoucher/go-tracking-server/utilities"
)



func main() {
  processorFlagPtr := flag.Bool("processor", false, "by default it starts an http server instance, use this flag to start a processor instead")

  flag.Parse()

  config, err := utilities.LoadJSONConfig()
  if err != nil {
    log.Fatalf("FATAL ERROR: reading json config: %s", err.Error())
  }

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

  if *processorFlagPtr {
    startProcessingServer(context)
  } else {
    startTrackingServer(context)
  }
}

func startTrackingServer(context *ctx.Context) {
  context.Logger.Fatalf("FATAL ERROR: from server: %+v", http.ListenAndServe(":1337", routes.Handlers(context)))
}

func startProcessingServer(context *ctx.Context) {
  msgs, err := context.Queue.ConsumeQueue("tracking-queue")
  if err != nil {
    context.Logger.Fatalf("FATAL ERROR: from queue consuming: %s", err.Error())
  }

  forever := make(chan bool)

  go func() {
    for m := range msgs {
      context.Logger.Infof("Recieved a message: %s", m.MessageId)
      //TODO : Store the payload in database
    }
  }()

  context.Logger.Info("[*] Waiting for messages...")
  <-forever
}
