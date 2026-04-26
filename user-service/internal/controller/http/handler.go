package http

import (
	"database/sql"
	"net/http"
	"user-service/internal/domain"
	"user-service/internal/usecase/getuser"
	"user-service/internal/usecase/register"
	"user-service/internal/usecase/search"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	registerUC *register.UseCase
	searchUC   *search.UseCase
	getUserUC  *getuser.UseCase
}

func NewHandler(reg *register.UseCase, srch *search.UseCase, get *getuser.UseCase) *Handler {
	return &Handler{registerUC: reg, searchUC: srch, getUserUC: get}
}

func (h *Handler) Register(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	v1.POST("/users", h.registerUser)
	v1.GET("/users", h.searchUsers)
	v1.GET("/users/:id", h.getUser)
}

func (h *Handler) registerUser(c *gin.Context) {
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.registerUC.Handle(c.Request.Context(), register.Request{Name: body.Name})
	if err != nil {
		if err == domain.ErrEmptyName {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toUserResponse(resp.User))
}

func (h *Handler) searchUsers(c *gin.Context) {
	q := c.Query("q")
	resp, err := h.searchUC.Handle(c.Request.Context(), search.Request{Query: q})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := make([]userResponse, 0, len(resp.Users))
	for _, u := range resp.Users {
		result = append(result, toUserResponse(u))
	}
	c.JSON(http.StatusOK, result)
}

func (h *Handler) getUser(c *gin.Context) {
	id := c.Param("id")
	resp, err := h.getUserUC.Handle(c.Request.Context(), getuser.Request{UserID: id})
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, toUserResponse(resp.User))
}
