package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"newshub/internal/broker"
)

type KafkaBrokerManager struct {
	brokerDns     string // kafka:9092
	topicChannels map[string]<-chan *broker.Message
}

func NewKafkaBrokerManager(brokerDns string) *KafkaBrokerManager {
	return &KafkaBrokerManager{
		brokerDns:     brokerDns,
		topicChannels: make(map[string]<-chan *broker.Message),
	}
}

func (manager *KafkaBrokerManager) PostMessageToTopic(topic string, message *broker.Message) {
	// Sarama configuration
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true // Ensure successful sends return confirmations

	// Create a new Kafka producer
	producer, err := sarama.NewSyncProducer([]string{manager.brokerDns}, config)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close() // Close the producer when done

	// Create a Kafka message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(message.Key),
		Value: sarama.StringEncoder(message.Payload),
	}

	// Send the message
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		log.Fatalf("Failed to send message to Kafka: %v", err)
	}

	fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)
}

func (manager *KafkaBrokerManager) GetTopicChannel(ctx context.Context, topic string) <-chan *broker.Message {
	if _, ok := manager.topicChannels[topic]; !ok {
		// Create Sarama configuration
		config := sarama.NewConfig()
		config.Consumer.Return.Errors = true // Enable error handling

		// Create a Kafka consumer
		consumer, err := sarama.NewConsumer([]string{manager.brokerDns}, config)
		if err != nil {
			log.Fatalf("Failed to create Kafka consumer: %v", err)
		}

		// Get a partition consumer (partition 0 in this example)
		partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetNewest)
		if err != nil {
			log.Fatalf("Failed to create partition consumer: %v", err)
		}

		messagesChannel := make(chan *broker.Message)
		go func() {
			for {
				select {

				case msg := <-partitionConsumer.Messages():
					messagesChannel <- broker.NewMessage(string(msg.Key), string(msg.Value))
				case errMsg := <-partitionConsumer.Errors():
					log.Fatalf(fmt.Sprintf("Some error occured: %s", errMsg.Error()))
				case <-ctx.Done(): // Check if the context has been canceled
					log.Println("Stopping update processing loop.")
					return // Exit the loop if context is canceled
				}
			}
		}()

		manager.topicChannels[topic] = messagesChannel
	}

	return manager.topicChannels[topic]
}
