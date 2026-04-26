package search

import (
	"context"
	"user-service/internal/domain"
)

type userSearcher interface {
	Search(ctx context.Context, query string) ([]*domain.User, error)
}

type UseCase struct {
	repo userSearcher
}

func NewUseCase(repo userSearcher) *UseCase {
	return &UseCase{repo: repo}
}

type Request struct {
	Query string
}

type Response struct {
	Users []*domain.User
}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	users, err := uc.repo.Search(ctx, req.Query)
	if err != nil {
		return Response{}, err
	}
	return Response{Users: users}, nil
}
