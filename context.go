package ctx

import (
  "github.com/OlivierBoucher/go-tracking-server/datastores"
  "github.com/OlivierBoucher/go-tracking-server/queues"
  "github.com/OlivierBoucher/go-tracking-server/validators"
)
//Context a context that holds database and queue connections
type Context struct {
  AuthDb *datastores.AuthDatastore
  Queue *queues.RabbitMQConnection
  JSONTrackingEventValidator *validators.JSONEventTrackingValidator
}
