package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/google/uuid"
	"log"
	"newshub/cmd/app/entities"
	"newshub/cmd/app/messages"
	"newshub/cmd/app/repositories"
	"newshub/internal/broker"
	"newshub/internal/broker/kafka"
	"newshub/internal/db"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

// Перефразируй следующий текст, чтобы использовать её как описание для поста в Instagram:
func main() {
	connectionProvider := db.NewConnectionProvider(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_NAME"))
	conn := connectionProvider.GetConnection()
	defer conn.Close()

	brokerManager := kafka.NewKafkaBrokerManager(os.Getenv("KAFKA_HOST"))

	repository := repositories.NewNewsPreviewRepository(conn)

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

	//parseUseCase(repository, brokerManager)

	processRequestConsumer := brokerManager.GetTopicChannel(ctx, "tg_process_request")
	for {
		select {
		case msg := <-processRequestConsumer:
			log.Printf("Accepted process request %v", msg.Payload)

			var processRequest messages.ProcessNewsPreviewMessage
			err := json.Unmarshal([]byte(msg.Payload), &processRequest)
			if err != nil {
				log.Fatalf("Error unmarshalling process request: %v", err)
			}

			newsPreview := repository.FindByUuid(processRequest.Uuid)

			if newsPreview == nil {
				log.Fatalf("Can't find news preview with uuid %v", processRequest.Uuid)
			}

			c := colly.NewCollector()
			// setting a valid User-Agent header
			c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

			var content string
			c.OnHTML("main > section[class*='_content-section_'] > div[class*='_content_'] > *", func(element *colly.HTMLElement) {
				if strings.Contains(element.Attr("class"), "_read-more_") {
					return
				}

				content += element.Text + "\n"
			})

			c.OnScraped(func(response *colly.Response) {
				if content == "" {
					log.Fatalf("Can't find content for news preview with uuid %v", processRequest.Uuid)
				}

				newsPreview.SetContent(content)
				repository.Update(newsPreview)
			})

			c.Visit(newsPreview.Link)

		case <-ctx.Done(): // Check if the context has been canceled
			log.Println("Stopping update processing loop.")
			return // Exit the loop if context is canceled
		}
	}
}

func parseUseCase(repository *repositories.NewsPreviewRepository, brokerManager broker.BrokerManager) {
	c := colly.NewCollector()
	// setting a valid User-Agent header
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	c.OnHTML("ul[data-index-news-area] li a", func(element *colly.HTMLElement) {
		title := ""

		link := element.Request.AbsoluteURL(element.Attr("href"))

		element.ForEach("span", func(i int, element *colly.HTMLElement) {
			if strings.Contains(element.Attr("class"), "_news-widget__item__title_") {
				title = strings.TrimSpace(element.Text)
			}
		})

		existedPreview := repository.FindByTitle(title)
		if existedPreview != nil {
			log.Println(fmt.Sprintf("Preview with title '%s' already existed. %v", title, existedPreview))
			return
		}

		preview := entities.NewNewsPreview(uuid.New(), title, link)

		repository.Add(preview)
		brokerManager.PostMessageToTopic("tg_news_previews", broker.NewMessage(preview.Uuid.String(), string(preview.Json())))
	})

	c.Visit("https://stopgame.ru/")
}
