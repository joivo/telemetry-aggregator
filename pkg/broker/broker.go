package broker

import (
	"github.com/streadway/amqp"
	"log"
)

var channel amqp.Channel

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func StartBroker() {
	conn, e := amqp.Dial("amqp://guest:guest@localhost:5672//")

	failOnError(e, "Failed to connect to RabbitMQ")

	defer conn.Close()

	ch, err := conn.Channel()

	failOnError(err, "Failed to open a channel")

	defer ch.Close()
}

func createQueue(ch amqp.Channel) amqp.Queue {
	q, err := ch.QueueDeclare(
		"metrics",
		false,
		false,
		false,
		false,
		nil,
		)
	failOnError(err, "Failed to declare a queue")

	return q
}

func Dispatch() {

}
