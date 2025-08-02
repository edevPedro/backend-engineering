package main
import (
	"os"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main()	{
	url := os.Getenv("CLOUDAMQP_URL")
	if url == "" {
		url = "amqp://localhost"
	}

	fmt.Println("Connecting to", url)

	connection, _ := amqp.Dial(url)
	defer connection.Close()

	go func(con *amqp.Connection) {
		channel, _ := connection.Channel()
		defer channel.Close()
		durable, exclusive := false, false
		autoDelete, noWait := true, false
		q, _ := channel.QueueDeclare("jobs", durable, autoDelete, exclusive, noWait, nil)
		channel.QueueBind(q.Name, "ping", "amq.topic", false, nil)
		autoAck, exclusive, noLocal, noWait := false, false, false, false
		messages, _ := channel.Consume(q.Name, "", autoAck, exclusive, noLocal, noWait, nil)
		multiAck := false
		for msg := range messages {
			fmt.Println("Body:", string(msg.Body), "Timestamp:", msg.Timestamp)
			msg.Ack(multiAck)
		}
	}(connection)

	go func(con *amqp.Connection) {
		timer := time.NewTicker(1 * time.Second)
		channel, _ := connection.Channel()

		for t := range timer.C {
			msg := amqp.Publishing{
				DeliveryMode: 1,
				Timestamp:    t,
				ContentType: "text/plain",
				Body:        []byte("Hello World!"),
			}
				mandatory, immediate := false, false
				channel.Publish("amq.topic", "ping", mandatory, immediate, msg)
		}
	}(connection)

	select {}
}