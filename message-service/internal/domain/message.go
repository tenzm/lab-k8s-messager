package domain

import "time"

type Message struct {
	ID         string
	SenderID   string
	ReceiverID string
	Text       string
	FileID     string
	FileName   string
	FileURL    string
	Edited     bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type File struct {
	ID          string
	OrigName    string
	ContentType string
	Path        string
	CreatedAt   time.Time
}
