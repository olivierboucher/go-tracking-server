package processor

import (
  "encoding/json"

  "github.com/streadway/amqp"

  "github.com/OlivierBoucher/go-tracking-server/ctx"
  "github.com/OlivierBoucher/go-tracking-server/models"
)
//ProcessMessage processes an amqp.Delivery within a context
func ProcessMessage(c *ctx.Context, m *amqp.Delivery) {
  //Decode the payload
  var payload models.EventTrackingPayload
  err := json.Unmarshal(m.Body, &payload)
  if err != nil {
    c.Logger.Errorf("Impossible to decode payload from message - Error: %s", err.Error())
    //We can ignore the err from Nack because auto-ack is false
    m.Nack(false, true)
    return
  }
  err = c.StorageDb.StoreBatchEvents(&payload)
  if err != nil {
    c.Logger.Errorf("Impossible to store payload from message - Error: %s", err.Error())
    //We can ignore the err from Nack because auto-ack is false
    m.Nack(false, true)
    return
  }
  //ACK that the message has been processed sucessfully
  c.Logger.Infof("Sucessfully processed message: %s",m.MessageId)
  err = m.Ack(false)
  if err != nil {
    c.Logger.Infof("Could not ack message: %s", err.Error())
  }
  return
}
