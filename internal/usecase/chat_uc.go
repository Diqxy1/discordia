package usecase

import (
	"discordia/internal/domain"
	"time"
)

type ChatUseCase struct {
	repo domain.MessageRepository
	p2p  domain.P2PService
}

func NewChatUseCase(r domain.MessageRepository, p domain.P2PService) *ChatUseCase {
	return &ChatUseCase{repo: r, p2p: p}
}

func (uc *ChatUseCase) SendMessage(sender, content string) error {
	msg := &domain.Message{
		Sender:    sender,
		Content:   content,
		Timestamp: time.Now(),
	}

	uc.repo.Save(msg)
	return uc.p2p.Publish(msg)
}

func (uc *ChatUseCase) StartReceiving(displayFunc func(domain.Message)) {
	uc.p2p.Subscribe(func(msg *domain.Message) {
		uc.repo.Save(msg)
		displayFunc(*msg)
	})
}
