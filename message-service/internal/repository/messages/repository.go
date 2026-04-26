package messages

import (
	"context"
	"database/sql"
	"message-service/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type msgDTO struct {
	ID         string         `db:"id"`
	SenderID   string         `db:"sender_id"`
	ReceiverID string         `db:"receiver_id"`
	Text       string         `db:"text"`
	FileID     sql.NullString `db:"file_id"`
	FileName   string         `db:"file_name"`
	Edited     bool           `db:"edited"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
}

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, senderID, receiverID, text, fileID, fileName string) (*domain.Message, error) {
	id := uuid.New().String()
	now := time.Now()
	var fid sql.NullString
	if fileID != "" {
		fid = sql.NullString{String: fileID, Valid: true}
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO messages (id, sender_id, receiver_id, text, file_id, file_name, edited, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,false,$7,$8)`,
		id, senderID, receiverID, text, fid, fileName, now, now,
	)
	if err != nil {
		return nil, err
	}
	return &domain.Message{
		ID: id, SenderID: senderID, ReceiverID: receiverID,
		Text: text, FileID: fileID, FileName: fileName, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	var dto msgDTO
	err := r.db.GetContext(ctx, &dto, `SELECT * FROM messages WHERE id = $1`, id)
	if err != nil {
		return nil, err
	}
	return fromDTO(dto), nil
}

func (r *Repository) Update(ctx context.Context, id, text string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE messages SET text=$1, edited=true, updated_at=$2 WHERE id=$3`,
		text, time.Now(), id,
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM messages WHERE id=$1`, id)
	return err
}

func (r *Repository) GetConversation(ctx context.Context, userA, userB string, afterID string, limit int) ([]*domain.Message, error) {
	var dtos []msgDTO
	var err error
	if afterID == "" {
		err = r.db.SelectContext(ctx, &dtos,
			`SELECT * FROM messages WHERE (sender_id=$1 AND receiver_id=$2) OR (sender_id=$2 AND receiver_id=$1) ORDER BY created_at DESC LIMIT $3`,
			userA, userB, limit,
		)
	} else {
		err = r.db.SelectContext(ctx, &dtos,
			`SELECT * FROM messages WHERE ((sender_id=$1 AND receiver_id=$2) OR (sender_id=$2 AND receiver_id=$1)) AND created_at > (SELECT created_at FROM messages WHERE id=$3) ORDER BY created_at ASC LIMIT $4`,
			userA, userB, afterID, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	msgs := make([]*domain.Message, 0, len(dtos))
	for _, d := range dtos {
		msgs = append(msgs, fromDTO(d))
	}
	return msgs, nil
}

type ConversationRow struct {
	PartnerID   string
	LastMessage *domain.Message
}

func (r *Repository) GetConversations(ctx context.Context, userID string) ([]ConversationRow, error) {
	// One row per partner — the most recent message in each conversation.
	var dtos []struct {
		PartnerID string `db:"partner_id"`
		msgDTO
	}
	err := r.db.SelectContext(ctx, &dtos, `
		SELECT DISTINCT ON (partner_id)
			CASE WHEN sender_id = $1 THEN receiver_id ELSE sender_id END AS partner_id,
			id, sender_id, receiver_id, text, file_id, file_name, edited, created_at, updated_at
		FROM messages
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY partner_id, created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	rows := make([]ConversationRow, 0, len(dtos))
	for _, d := range dtos {
		rows = append(rows, ConversationRow{
			PartnerID:   d.PartnerID,
			LastMessage: fromDTO(d.msgDTO),
		})
	}
	return rows, nil
}

func fromDTO(d msgDTO) *domain.Message {
	m := &domain.Message{
		ID:         d.ID,
		SenderID:   d.SenderID,
		ReceiverID: d.ReceiverID,
		Text:       d.Text,
		FileName:   d.FileName,
		Edited:     d.Edited,
		CreatedAt:  d.CreatedAt,
		UpdatedAt:  d.UpdatedAt,
	}
	if d.FileID.Valid {
		m.FileID = d.FileID.String
	}
	return m
}
