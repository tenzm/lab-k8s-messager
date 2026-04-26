package getconversations

import (
	"context"
	"message-service/internal/domain"
	"message-service/internal/repository/messages"
)

type conversationsGetter interface {
	GetConversations(ctx context.Context, userID string) ([]messages.ConversationRow, error)
}

type UseCase struct {
	repo conversationsGetter
}

func NewUseCase(repo conversationsGetter) *UseCase {
	return &UseCase{repo: repo}
}

type Request struct {
	UserID string
}

type ConversationItem struct {
	PartnerID   string
	LastMessage *domain.Message
}

type Response struct {
	Conversations []ConversationItem
}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	rows, err := uc.repo.GetConversations(ctx, req.UserID)
	if err != nil {
		return Response{}, err
	}
	items := make([]ConversationItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, ConversationItem{PartnerID: r.PartnerID, LastMessage: r.LastMessage})
	}
	return Response{Conversations: items}, nil
}
