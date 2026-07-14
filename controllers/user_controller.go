package controllers

import (
	"net/http"
	"strconv"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	repo repository.UserRepository
}

type CreateUserRequest struct {
	FullName    string `json:"full_name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Email       string `json:"email" binding:"required"`
}

type CreateUserResponse struct {
	ID          uint   `json:"id"`
	ChatID      string `json:"chat_id"`
	FullName    string `json:"full_name"`
	PhoneNumber string `json:"phone_number"`
	Email       string `json:"email"`
}

func NewUserController(repo repository.UserRepository) *UserController {
	return &UserController{repo}
}

func (c *UserController) Create(ctx *gin.Context) {
	var request CreateUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid user request body", err.Error())
		return
	}

	user := models.User{
		ChatID:      request.PhoneNumber,
		FullName:    request.FullName,
		PhoneNumber: request.PhoneNumber,
		Email:       request.Email,
	}

	if err := c.repo.Create(&user); err != nil {
		respondError(ctx, http.StatusUnprocessableEntity, "USER_CREATE_FAILED", "User could not be created", err.Error())
		return
	}

	response := CreateUserResponse{
		ID:          user.ID,
		ChatID:      user.ChatID,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Email:       user.Email,
	}

	respondSuccess(ctx, http.StatusCreated, "User created successfully", response)
}

func (c *UserController) GetAll(ctx *gin.Context) {
	users, err := c.repo.FindAll()
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "USERS_FETCH_FAILED", "Users could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Users fetched successfully", users)
}

func (c *UserController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user id", err.Error())
		return
	}
	user, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "USER_NOT_FOUND", "User not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "USER_FETCH_FAILED", "User could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "User fetched successfully", user)
}

func (c *UserController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user id", err.Error())
		return
	}
	user, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "USER_NOT_FOUND", "User not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "USER_FETCH_FAILED", "User could not be fetched", err.Error())
		return
	}
	if err := ctx.ShouldBindJSON(user); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid user request body", err.Error())
		return
	}
	if err := c.repo.Update(user); err != nil {
		respondError(ctx, http.StatusInternalServerError, "USER_UPDATE_FAILED", "User could not be updated", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "User updated successfully", user)
}

func (c *UserController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user id", err.Error())
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		respondError(ctx, http.StatusInternalServerError, "USER_DELETE_FAILED", "User could not be deleted", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "User deleted successfully", gin.H{"id": id})
}

// GetHistory fetches the Redis chat history for a specific ChatID
func (c *UserController) GetHistory(ctx *gin.Context) {
	chatID := ctx.Param("chat_id")

	if chatID == "" {
		respondError(ctx, http.StatusBadRequest, "CHAT_ID_REQUIRED", "chat_id parameter is required", "")
		return
	}

	// Pass the context from the request so it can be used by the Redis client
	history, err := c.repo.GetChatHistory(ctx.Request.Context(), chatID)
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "CHAT_HISTORY_FETCH_FAILED", "Chat history could not be fetched", err.Error())
		return
	}

	respondSuccess(ctx, http.StatusOK, "Chat history fetched successfully", history)
}

func (c *UserController) ClearHistory(ctx *gin.Context) {
	chatID := ctx.Param("id")

	if chatID == "" {
		respondError(ctx, http.StatusBadRequest, "CHAT_ID_REQUIRED", "id parameter is required", "")
		return
	}

	if err := c.repo.DeleteChatHistory(ctx.Request.Context(), chatID); err != nil {
		respondError(ctx, http.StatusInternalServerError, "CHAT_HISTORY_CLEAR_FAILED", "Chat history could not be cleared", err.Error())
		return
	}

	respondSuccess(ctx, http.StatusOK, "Chat history cleared successfully", gin.H{"chat_id": chatID})
}
