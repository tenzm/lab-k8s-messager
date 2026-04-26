package http

import (
	"message-service/internal/domain"
	"message-service/internal/usecase/deletemessage"
	"message-service/internal/usecase/editmessage"
	"message-service/internal/usecase/getconversations"
	"message-service/internal/usecase/getmessages"
	"message-service/internal/usecase/sendmessage"
	"message-service/internal/usecase/uploadfile"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	sendUC          *sendmessage.UseCase
	editUC          *editmessage.UseCase
	deleteUC        *deletemessage.UseCase
	getUC           *getmessages.UseCase
	getConvsUC      *getconversations.UseCase
	uploadUC        *uploadfile.UseCase
	uploadsDir      string
}

func NewHandler(send *sendmessage.UseCase, edit *editmessage.UseCase, del *deletemessage.UseCase,
	get *getmessages.UseCase, getConvs *getconversations.UseCase, upload *uploadfile.UseCase, uploadsDir string) *Handler {
	return &Handler{sendUC: send, editUC: edit, deleteUC: del, getUC: get, getConvsUC: getConvs, uploadUC: upload, uploadsDir: uploadsDir}
}

func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	v1.POST("/messages", h.sendMessage)
	v1.PUT("/messages/:id", h.editMessage)
	v1.DELETE("/messages/:id", h.deleteMessage)
	v1.GET("/messages", h.getMessages)
	v1.GET("/conversations", h.getConversations)
	v1.POST("/files", h.uploadFile)
	v1.GET("/files/:id", h.serveFile)
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
	resp, err := h.sendUC.Handle(c.Request.Context(), sendmessage.Request{
		SenderID: body.SenderID, ReceiverID: body.ReceiverID, Text: body.Text, FileID: body.FileID, FileName: body.FileName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toMsgResponse(resp.Message))
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
	resp, err := h.editUC.Handle(c.Request.Context(), editmessage.Request{MessageID: id, UserID: body.UserID, NewText: body.NewText})
	if err != nil {
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toMsgResponse(resp.Message))
}

func (h *Handler) deleteMessage(c *gin.Context) {
	id := c.Param("id")
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}
	_, err := h.deleteUC.Handle(c.Request.Context(), deletemessage.Request{MessageID: id, UserID: userID})
	if err != nil {
		if err == domain.ErrForbidden {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		if err == domain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
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

	resp, err := h.getUC.Handle(c.Request.Context(), getmessages.Request{
		UserA: userA, UserB: userB, AfterID: afterID, Limit: limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]msgResponse, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		result = append(result, toMsgResponse(m))
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) getConversations(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id required"})
		return
	}
	resp, err := h.getConvsUC.Handle(c.Request.Context(), getconversations.Request{UserID: userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	type convResponse struct {
		PartnerID   string      `json:"partner_id"`
		LastMessage msgResponse `json:"last_message"`
	}
	result := make([]convResponse, 0, len(resp.Conversations))
	for _, item := range resp.Conversations {
		result = append(result, convResponse{
			PartnerID:   item.PartnerID,
			LastMessage: toMsgResponse(item.LastMessage),
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

	resp, err := h.uploadUC.Handle(c.Request.Context(), uploadfile.Request{
		OrigName:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Reader:      file,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":       resp.File.ID,
		"orig_name": resp.File.OrigName,
		"url":      "/api/v1/files/" + resp.File.ID,
	})
}

func (h *Handler) serveFile(c *gin.Context) {
	id := c.Param("id")
	// Find file by scanning uploads dir with matching uuid prefix
	matches, err := filepath.Glob(filepath.Join(h.uploadsDir, id+"*"))
	if err != nil || len(matches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}
	c.File(matches[0])
}

