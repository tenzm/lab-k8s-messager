package editmessage

import (
	"context"
	"message-service/internal/domain"
)

type messageGetterUpdater interface {
	GetByID(ctx context.Context, id string) (*domain.Message, error)
	Update(ctx context.Context, id, text string) error
}

type UseCase struct {
	repo messageGetterUpdater
}

func NewUseCase(repo messageGetterUpdater) *UseCase {
	return &UseCase{repo: repo}
}

type Request struct {
	MessageID string
	UserID    string
	NewText   string
}

type Response struct {
	Message *domain.Message
}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	msg, err := uc.repo.GetByID(ctx, req.MessageID)
	if err != nil {
		return Response{}, domain.ErrNotFound
	}
	if msg.SenderID != req.UserID {
		return Response{}, domain.ErrForbidden
	}
	if err := uc.repo.Update(ctx, req.MessageID, req.NewText); err != nil {
		return Response{}, err
	}
	msg.Text = req.NewText
	msg.Edited = true
	return Response{Message: msg}, nil
}
