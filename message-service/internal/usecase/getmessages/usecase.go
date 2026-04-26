package getmessages

import (
	"context"
	"message-service/internal/domain"
)

type conversationGetter interface {
	GetConversation(ctx context.Context, userA, userB string, afterID string, limit int) ([]*domain.Message, error)
}

type UseCase struct {
	repo conversationGetter
}

func NewUseCase(repo conversationGetter) *UseCase {
	return &UseCase{repo: repo}
}

type Request struct {
	UserA   string
	UserB   string
	AfterID string
	Limit   int
}

type Response struct {
	Messages []*domain.Message
}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	if req.Limit <= 0 {
		req.Limit = 50
	}
	msgs, err := uc.repo.GetConversation(ctx, req.UserA, req.UserB, req.AfterID, req.Limit)
	if err != nil {
		return Response{}, err
	}
	return Response{Messages: msgs}, nil
}
