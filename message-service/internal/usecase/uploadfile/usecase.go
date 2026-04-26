package uploadfile

import (
	"context"
	"fmt"
	"io"
	"message-service/internal/domain"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type fileSaver interface {
	Save(ctx context.Context, origName, contentType, path string) (*domain.File, error)
}

type UseCase struct {
	repo       fileSaver
	uploadsDir string
}

func NewUseCase(repo fileSaver, uploadsDir string) *UseCase {
	return &UseCase{repo: repo, uploadsDir: uploadsDir}
}

type Request struct {
	OrigName    string
	ContentType string
	Reader      io.Reader
}

type Response struct {
	File *domain.File
}

func (uc *UseCase) Handle(ctx context.Context, req Request) (Response, error) {
	id := uuid.New().String()
	ext := filepath.Ext(req.OrigName)
	filename := id + ext
	path := filepath.Join(uc.uploadsDir, filename)

	if err := os.MkdirAll(uc.uploadsDir, 0755); err != nil {
		return Response{}, fmt.Errorf("mkdir: %w", err)
	}

	dst, err := os.Create(path)
	if err != nil {
		return Response{}, fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, req.Reader); err != nil {
		return Response{}, fmt.Errorf("write file: %w", err)
	}

	file, err := uc.repo.Save(ctx, req.OrigName, req.ContentType, path)
	if err != nil {
		return Response{}, err
	}
	file.ID = id + ext // keep the ID as uuid, override path-based
	file.ID = id
	return Response{File: file}, nil
}
