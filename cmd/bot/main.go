package main

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	authorizedUserId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_USERID"), 10, 64)

	// Set up a context with cancellation to manage Goroutines
	ctx, cancel := context.WithCancel(context.Background())

	// Set up signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM) // Listen for SIGINT (Ctrl+C) or SIGTERM

	// Goroutine to handle system signals
	go func() {
		<-signalChan // Block until a signal is received
		log.Println("Received shutdown signal, canceling context...")
		cancel() // Cancel the context to stop Goroutines
	}()

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	telegramUpdates := bot.GetUpdatesChan(u)
	newsConsumer := getNewsPreviewChannel()

	for {
		select {
		case update := <-telegramUpdates: // Get updates from Telegram
			if update.Message == nil {
				continue
			}

			chatID := update.Message.Chat.ID
			userID := update.Message.From.ID

			// Check if the user is authorized
			if userID != authorizedUserId {
				log.Printf("Unauthorized access by user ID: %d, allowed is %d", userID, authorizedUserId)
				continue // Skip unauthorized users
			}

			// If authorized, process the message
			tgMsg := tgbotapi.NewMessage(chatID, "Hello, authorized user!")
			bot.Send(tgMsg)
		case msg := <-newsConsumer.Messages():
			log.Println(fmt.Sprintf("News preview message consumed: %s", msg.Value))

			tgMsg := tgbotapi.NewMessage(authorizedUserId, string(msg.Value))
			bot.Send(tgMsg)
		case errMsg := <-newsConsumer.Errors():
			log.Println(fmt.Sprintf("Some error occured: %s", errMsg.Error()))
		case <-ctx.Done(): // Check if the context has been canceled
			log.Println("Stopping update processing loop.")
			bot.StopReceivingUpdates()
			newsConsumer.Close()
			return // Exit the loop if context is canceled
		}
	}
}

func getNewsPreviewChannel() sarama.PartitionConsumer {
	// Kafka broker address
	broker := "kafka:9092" // Kafka broker to connect to
	// Kafka topic to consume from
	topic := "tg_news_previews"

	// Create Sarama configuration
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true // Enable error handling

	// Create a Kafka consumer
	consumer, err := sarama.NewConsumer([]string{broker}, config)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}

	// Get a partition consumer (partition 0 in this example)
	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to create partition consumer: %v", err)
	}

	return partitionConsumer
}
