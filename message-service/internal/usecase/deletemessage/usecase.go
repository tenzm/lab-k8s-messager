package deletemessage

import (
	"context"
	"message-service/internal/domain"
)

type messageGetterDeleter interface {
	GetByID(ctx context.Context, id string) (*domain.Message, error)
	Delete(ctx context.Context, id string) error
}

type UseCase struct {
	repo messageGetterDeleter
}

func NewUseCase(repo messageGetterDeleter) *UseCase {
	return &UseCase{repo: repo}
}

type Request struct {
	MessageID string
	UserID    string
}

type Response struct{}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	msg, err := uc.repo.GetByID(ctx, req.MessageID)
	if err != nil {
		return Response{}, domain.ErrNotFound
	}
	if msg.SenderID != req.UserID {
		return Response{}, domain.ErrForbidden
	}
	if err := uc.repo.Delete(ctx, req.MessageID); err != nil {
		return Response{}, err
	}
	return Response{}, nil
}
