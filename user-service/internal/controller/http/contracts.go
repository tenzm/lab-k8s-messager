package http

import "context"

type registerUseCase interface {
	Handle(ctx context.Context, req interface{}) (interface{}, error)
}
