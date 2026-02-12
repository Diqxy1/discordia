package domain

import "time"

type Message struct {
	Sender    string    `json:"sender"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Interface que o banco de dados deve seguir
type MessageRepository interface {
	Save(msg *Message) error
	GetAll() ([]Message, error)
}

// Interface que a rede P2P deve seguir
type P2PService interface {
	Publish(msg *Message) error
	Subscribe(handler func(*Message))
}
