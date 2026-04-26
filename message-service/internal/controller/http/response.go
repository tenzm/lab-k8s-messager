package http

import (
	"message-service/internal/domain"
	"time"
)

type msgResponse struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Text       string    `json:"text"`
	FileID     string    `json:"file_id,omitempty"`
	FileName   string    `json:"file_name,omitempty"`
	FileURL    string    `json:"file_url,omitempty"`
	Edited     bool      `json:"edited"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func toMsgResponse(m *domain.Message) msgResponse {
	r := msgResponse{
		ID:         m.ID,
		SenderID:   m.SenderID,
		ReceiverID: m.ReceiverID,
		Text:       m.Text,
		FileID:     m.FileID,
		FileName:   m.FileName,
		Edited:     m.Edited,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
	if m.FileID != "" {
		r.FileURL = "/api/v1/files/" + m.FileID
	}
	return r
}
