package files

import (
	"context"
	"message-service/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type fileDTO struct {
	ID          string    `db:"id"`
	OrigName    string    `db:"orig_name"`
	ContentType string    `db:"content_type"`
	Path        string    `db:"path"`
	CreatedAt   time.Time `db:"created_at"`
}

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(ctx context.Context, origName, contentType, path string) (*domain.File, error) {
	id := uuid.New().String()
	now := time.Now()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO files (id, orig_name, content_type, path, created_at) VALUES ($1,$2,$3,$4,$5)`,
		id, origName, contentType, path, now,
	)
	if err != nil {
		return nil, err
	}
	return &domain.File{ID: id, OrigName: origName, ContentType: contentType, Path: path, CreatedAt: now}, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.File, error) {
	var dto fileDTO
	err := r.db.GetContext(ctx, &dto, `SELECT * FROM files WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	return &domain.File{ID: dto.ID, OrigName: dto.OrigName, ContentType: dto.ContentType, Path: dto.Path, CreatedAt: dto.CreatedAt}, nil
}
