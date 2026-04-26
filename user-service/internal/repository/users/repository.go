package users

import (
	"context"
	"time"
	"user-service/internal/domain"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type userDTO struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
}

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, name string) (*domain.User, error) {
	id := uuid.New().String()
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, name, created_at) VALUES ($1, $2, $3)`,
		id, name, time.Now(),
	)
	if err != nil {
		return nil, err
	}
	return &domain.User{ID: id, Name: name, CreatedAt: time.Now()}, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var dto userDTO
	err := r.db.GetContext(ctx, &dto, `SELECT id, name, created_at FROM users WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return &domain.User{ID: dto.ID, Name: dto.Name, CreatedAt: dto.CreatedAt}, nil
}

func (r *Repository) Search(ctx context.Context, query string) ([]*domain.User, error) {
	var dtos []userDTO
	err := r.db.SelectContext(ctx, &dtos,
		`SELECT id, name, created_at FROM users WHERE name ILIKE $1 ORDER BY name LIMIT 50`,
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	users := make([]*domain.User, 0, len(dtos))
	for _, d := range dtos {
		users = append(users, &domain.User{ID: d.ID, Name: d.Name, CreatedAt: d.CreatedAt})
	}
	return users, nil
}
