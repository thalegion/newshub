package broker

import "context"

type BrokerManager interface {
	PostMessageToTopic(topic string, message *Message)
	GetTopicChannel(ctx context.Context, topic string) <-chan *Message
}
