package broker

type Message struct {
	Key     string
	Payload string
}

func NewMessage(key string, payload string) *Message {
	return &Message{Key: key, Payload: payload}
}
