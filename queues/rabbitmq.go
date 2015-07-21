package queues

import (
  "log"
  "github.com/streadway/amqp"
)
//RabbitMQConnection wrapper around amqp.Connection
type RabbitMQConnection struct {
  amqp.Connection
}
//NewRabbitMQConnection initializes a wrapper around an amqp.Connection
func NewRabbitMQConnection(conn *amqp.Connection) *RabbitMQConnection {
  return &RabbitMQConnection{*conn}
}
//PublishEventsTrackingTask publishes a json payload to the tracking exchange
func (c *RabbitMQConnection) PublishEventsTrackingTask(payload []byte) {
  ch, err := c.Channel()
  if err != nil {
    log.Print("Error creating channel")
  }
  defer ch.Close()

  err = ch.Publish(
    "tracking",   //Exchange
    "tracking-queue",           //Routing key
    false,        //Mandatory
    false,        //Immediate
    amqp.Publishing{
          DeliveryMode: amqp.Persistent,
          ContentType: "text/plain",
          Body: payload,
    })
    if err != nil {
      //Find a second way to persist the events
      log.Printf("Error publishing to tracking queue: %s", err.Error())
    }
}
