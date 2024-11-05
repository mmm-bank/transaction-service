package main

import (
	messaging "github.com/mmm-bank/infra/rabbitmq"
	"github.com/mmm-bank/transaction-service/http"
	"github.com/mmm-bank/transaction-service/storage"
	"log"
	"os"
)

func setUpRabbitMQ() {
	conn := messaging.NewConn(os.Getenv("RABBITMQ_URL"))
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to create channel")
	}
	ch.ExchangeDeclare(
		"transaction_events",
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	postgresQueue, err := ch.QueueDeclare(
		"postgres",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare PostgreSQL queue: %v", err)
	}

	mongoQueue, err := ch.QueueDeclare(
		"mongo",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare MongoDB queue: %v", err)
	}

	err = ch.QueueBind(
		postgresQueue.Name,
		"",
		"transaction_events",
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind PostgreSQL queue to exchange: %v", err)
	}

	err = ch.QueueBind(
		mongoQueue.Name,
		"",
		"transaction_events",
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind MongoDB queue to exchange: %v", err)
	}
}

func main() {
	addr := ":8080"
	s := storage.NewPostgresCards(os.Getenv("POSTGRES_URL"))
	server := http.NewCardService(s)
	setUpRabbitMQ()

	log.Printf("Transaction server is running on port %s...", addr[1:])
	if err := http.CreateAndRunServer(server, addr); err != nil {
		log.Fatalf("Failed to start server: %s", err)
	}
}
