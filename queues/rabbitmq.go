package queues

import (
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
func (c *RabbitMQConnection) PublishEventsTrackingTask(payload []byte) error {
	ch, err := c.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.Publish(
		"tracking",       //Exchange
		"tracking-queue", //Routing key
		false,            //Mandatory
		false,            //Immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         payload,
		})
	if err != nil {
		return err
	}
	return nil
}

//ConsumeQueueWithChannel consumes a RabbitMQConnection queue with predefined settings from a provided amqp.Channel
func (c *RabbitMQConnection) ConsumeQueueWithChannel(queue string, ch *amqp.Channel) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		queue,
		"",
		false, //Auto ACK. False because we want to make sure of data integrity
		false,
		false,
		false,
		nil,
	)
}
