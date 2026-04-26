package sendmessage

import (
	"context"
	"message-service/internal/domain"
)

type messageCreator interface {
	Create(ctx context.Context, senderID, receiverID, text, fileID, fileName string) (*domain.Message, error)
}

type UseCase struct {
	repo messageCreator
}

func NewUseCase(repo messageCreator) *UseCase {
	return &UseCase{repo: repo}
}

type Request struct {
	SenderID   string
	ReceiverID string
	Text       string
	FileID     string
	FileName   string
}

type Response struct {
	Message *domain.Message
}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	if req.SenderID == "" || req.ReceiverID == "" {
		return Response{}, domain.ErrBadRequest
	}
	msg, err := uc.repo.Create(ctx, req.SenderID, req.ReceiverID, req.Text, req.FileID, req.FileName)
	if err != nil {
		return Response{}, err
	}
	return Response{Message: msg}, nil
}
