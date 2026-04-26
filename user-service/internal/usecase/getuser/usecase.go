package getuser

import (
	"context"
	"user-service/internal/domain"
)

type userGetter interface {
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type UseCase struct {
	repo userGetter
}

func NewUseCase(repo userGetter) *UseCase {
	return &UseCase{repo: repo}
}

type Request struct {
	UserID string
}

type Response struct {
	User *domain.User
}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	user, err := uc.repo.GetByID(ctx, req.UserID)
	if err != nil {
		return Response{}, err
	}
	return Response{User: user}, nil
}
