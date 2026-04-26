package http

import (
	"bff/internal/external/messageservice"
	"bff/internal/external/userservice"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	users    *userservice.Client
	messages *messageservice.Client
	msgSvcBaseURL string
}

func NewHandler(users *userservice.Client, messages *messageservice.Client, msgSvcBaseURL string) *Handler {
	return &Handler{users: users, messages: messages, msgSvcBaseURL: msgSvcBaseURL}
}

func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")

	// users
	v1.POST("/users", h.registerUser)
	v1.GET("/users", h.searchUsers)
	v1.GET("/users/:id", h.getUser)

	// messages
	v1.POST("/messages", h.sendMessage)
	v1.PUT("/messages/:id", h.editMessage)
	v1.DELETE("/messages/:id", h.deleteMessage)
	v1.GET("/messages", h.getMessages)

	// conversations
	v1.GET("/conversations", h.getConversations)

	// long polling
	v1.GET("/poll", h.poll)

	// file upload
	v1.POST("/files", h.uploadFile)
	// file download proxied via message-service
	v1.GET("/files/:id", h.proxyFile)
}

func (h *Handler) registerUser(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := h.users.Register(c.Request.Context(), body.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *Handler) searchUsers(c *gin.Context) {
	q := c.Query("q")
	users, err := h.users.SearchUsers(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (h *Handler) getUser(c *gin.Context) {
	user, err := h.users.GetUser(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *Handler) sendMessage(c *gin.Context) {
	var body struct {
		SenderID   string `json:"sender_id" binding:"required"`
		ReceiverID string `json:"receiver_id" binding:"required"`
		Text       string `json:"text"`
		FileID     string `json:"file_id"`
		FileName   string `json:"file_name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg, err := h.messages.SendMessage(c.Request.Context(), body.SenderID, body.ReceiverID, body.Text, body.FileID, body.FileName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h *Handler) editMessage(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		UserID  string `json:"user_id" binding:"required"`
		NewText string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	msg, err := h.messages.EditMessage(c.Request.Context(), id, body.UserID, body.NewText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msg)
}

func (h *Handler) deleteMessage(c *gin.Context) {
	id := c.Param("id")
	userID := c.Query("user_id")
	if err := h.messages.DeleteMessage(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) getMessages(c *gin.Context) {
	userA := c.Query("user_a")
	userB := c.Query("user_b")
	afterID := c.Query("after_id")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	msgs, err := h.messages.GetMessages(c.Request.Context(), userA, userB, afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

// poll implements long polling: waits up to 30s for new messages after afterID
func (h *Handler) poll(c *gin.Context) {
	userA := c.Query("user_a")
	userB := c.Query("user_b")
	afterID := c.Query("after_id")

	deadline := time.Now().Add(30 * time.Second)
	for {
		msgs, err := h.messages.GetMessages(c.Request.Context(), userA, userB, afterID, 50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if len(msgs) > 0 {
			c.JSON(http.StatusOK, msgs)
			return
		}
		if time.Now().After(deadline) {
			c.JSON(http.StatusOK, []*messageservice.Message{})
			return
		}
		select {
		case <-c.Request.Context().Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}

func (h *Handler) getConversations(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}
	convs, err := h.messages.GetConversations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type convWithUser struct {
		PartnerID   string                     `json:"partner_id"`
		PartnerName string                     `json:"partner_name"`
		LastMessage *messageservice.Message    `json:"last_message"`
	}

	result := make([]convWithUser, 0, len(convs))
	for _, conv := range convs {
		name := conv.PartnerID
		if u, err := h.users.GetUser(c.Request.Context(), conv.PartnerID); err == nil {
			name = u.Name
		}
		result = append(result, convWithUser{
			PartnerID:   conv.PartnerID,
			PartnerName: name,
			LastMessage: conv.LastMessage,
		})
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) uploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file required"})
		return
	}
	defer file.Close()

	result, err := h.messages.UploadFile(c.Request.Context(), header.Filename, header.Header.Get("Content-Type"), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

func (h *Handler) proxyFile(c *gin.Context) {
	id := c.Param("id")
	target := h.msgSvcBaseURL + "/api/v1/files/" + id
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
}
