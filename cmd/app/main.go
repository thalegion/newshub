package main

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/gocolly/colly"
	"log"
	"newshub/cmd/app/structs"
	"strings"
)

func main() {
	c := colly.NewCollector()
	// setting a valid User-Agent header
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"

	c.OnHTML("ul[data-index-news-area] li a", func(element *colly.HTMLElement) {
		title := ""

		link := element.Request.AbsoluteURL(element.Attr("href"))

		element.ForEach("span", func(i int, element *colly.HTMLElement) {
			if strings.Contains(element.Attr("class"), "_news-widget__item__title_") {
				title = element.Text
			}
		})

		preview := structs.NewNewsPreview(title, link)

		// todo: Check if preview was already processed (in cache, ttl 1d) -> send message though broker to telegram bot
		fmt.Println(preview)

		postNewsPreview(preview)
	})

	c.Visit("https://stopgame.ru/")
}

func postNewsPreview(newsPreview structs.NewsPreview) {
	// Kafka broker address
	broker := "kafka:9092" // The Kafka broker address
	// Kafka topic name
	topic := "tg_news_previews"

	// Sarama configuration
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true // Ensure successful sends return confirmations

	// Create a new Kafka producer
	producer, err := sarama.NewSyncProducer([]string{broker}, config)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close() // Close the producer when done

	previewJson, _ := newsPreview.Json()
	// Create a Kafka message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%x", newsPreview.GetHash())), // Optional key for partitioning
		Value: sarama.StringEncoder(previewJson),                              // The message payload
	}

	// Send the message
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("Failed to send message to Kafka: %v", err)
	}

	fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)
}
