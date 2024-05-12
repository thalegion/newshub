package main

import (
	"context"
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"newshub/cmd/bot/messages"
	"newshub/internal/broker"
	"newshub/internal/broker/kafka"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

func main() {
	token := os.Getenv("TELEGRAM_TOKEN")
	authorizedUserId, _ := strconv.ParseInt(os.Getenv("TELEGRAM_USERID"), 10, 64)

	brokerManager := kafka.NewKafkaBrokerManager(os.Getenv("KAFKA_HOST"))

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
	newsConsumer := brokerManager.GetTopicChannel(ctx, "tg_news_previews")

	for {
		select {
		case update := <-telegramUpdates: // Get updates from Telegram
			if update.Message != nil {

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
			} else if update.CallbackQuery != nil {
				callbackDataParams := strings.Split(update.CallbackQuery.Data, "|")

				switch callbackDataParams[0] {
				case "process":
					previewUuid := callbackDataParams[1]

					callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "Start process of NP "+previewUuid)
					if _, err := bot.Request(callback); err != nil {
						panic(err)
					}

					// Create a map with some data
					data := map[string]interface{}{
						"uuid": previewUuid,
					}

					// Encode the map to JSON
					jsonData, err := json.Marshal(data)
					if err != nil {
						log.Fatalf("Error during marshaling process message %v", err)
					}
					brokerManager.PostMessageToTopic("tg_process_request", broker.NewMessage(previewUuid, string(jsonData)))
				}
			}
		case msg := <-newsConsumer:
			log.Println(fmt.Sprintf("News preview message consumed: %s", msg.Payload))

			var newsPreviewMsg messages.NewNewsPreview
			err := json.Unmarshal([]byte(msg.Payload), &newsPreviewMsg)
			if err != nil {
				log.Printf("Error unmarshalling news preview message: %s", err)
				continue
			}

			// Create an inline keyboard with one button
			keyboard := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonData("PROCESS", "process|"+newsPreviewMsg.Uuid), // Button text and callback data
				),
			)

			tgMsg := tgbotapi.NewMessage(authorizedUserId, newsPreviewMsg.Link)
			tgMsg.ReplyMarkup = keyboard
			if _, err := bot.Send(tgMsg); err != nil {
				log.Panic(err)
			}
		case <-ctx.Done(): // Check if the context has been canceled
			log.Println("Stopping update processing loop.")
			bot.StopReceivingUpdates()
			return // Exit the loop if context is canceled
		}
	}
}
