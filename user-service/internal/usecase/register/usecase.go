package register

import (
	"context"
	"user-service/internal/domain"
)

type userCreator interface {
	Create(ctx context.Context, name string) (*domain.User, error)
}

type UseCase struct {
	repo userCreator
}

func NewUseCase(repo userCreator) *UseCase {
	return &UseCase{repo: repo}
}

type Request struct {
	Name string
}

type Response struct {
	User *domain.User
}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	if req.Name == "" {
		return Response{}, domain.ErrEmptyName
	}
	user, err := uc.repo.Create(ctx, req.Name)
	if err != nil {
		return Response{}, err
	}
	return Response{User: user}, nil
}
