package messageservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

type Message struct {
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

type FileUploadResponse struct {
	ID       string `json:"id"`
	OrigName string `json:"orig_name"`
	URL      string `json:"url"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{baseURL: baseURL, httpClient: &http.Client{Timeout: 30 * time.Second}}
}

func (c *Client) SendMessage(ctx context.Context, senderID, receiverID, text, fileID, fileName string) (*Message, error) {
	body, _ := json.Marshal(map[string]string{
		"sender_id":   senderID,
		"receiver_id": receiverID,
		"text":        text,
		"file_id":     fileID,
		"file_name":   fileName,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("msg-service: status %d", resp.StatusCode)
	}
	var msg Message
	return &msg, json.NewDecoder(resp.Body).Decode(&msg)
}

func (c *Client) EditMessage(ctx context.Context, msgID, userID, text string) (*Message, error) {
	body, _ := json.Marshal(map[string]string{"user_id": userID, "text": text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/api/v1/messages/"+msgID, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var msg Message
	return &msg, json.NewDecoder(resp.Body).Decode(&msg)
}

func (c *Client) DeleteMessage(ctx context.Context, msgID, userID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+"/api/v1/messages/"+msgID+"?user_id="+url.QueryEscape(userID), nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("msg-service: status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) GetMessages(ctx context.Context, userA, userB, afterID string, limit int) ([]*Message, error) {
	u := fmt.Sprintf("%s/api/v1/messages?user_a=%s&user_b=%s&after_id=%s&limit=%d",
		c.baseURL, url.QueryEscape(userA), url.QueryEscape(userB), url.QueryEscape(afterID), limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var msgs []*Message
	return msgs, json.NewDecoder(resp.Body).Decode(&msgs)
}

type ConversationItem struct {
	PartnerID   string   `json:"partner_id"`
	LastMessage *Message `json:"last_message"`
}

func (c *Client) GetConversations(ctx context.Context, userID string) ([]*ConversationItem, error) {
	u := fmt.Sprintf("%s/api/v1/conversations?user_id=%s", c.baseURL, url.QueryEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var items []*ConversationItem
	return items, json.NewDecoder(resp.Body).Decode(&items)
}

func (c *Client) UploadFile(ctx context.Context, filename, contentType string, reader io.Reader) (*FileUploadResponse, error) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, reader); err != nil {
		return nil, err
	}
	w.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/files", &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("msg-service upload: status %d", resp.StatusCode)
	}
	var result FileUploadResponse
	return &result, json.NewDecoder(resp.Body).Decode(&result)
}
