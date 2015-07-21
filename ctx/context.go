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
//NewContext returns a new context from arguments
func NewContext(a *datastores.AuthDatastore, q *queues.RabbitMQConnection, jtv *validators.JSONEventTrackingValidator) *Context {
  return &Context{
    AuthDb: a,
    Queue:q,
    JSONTrackingEventValidator: jtv,
  }
}
