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

type CreateUserResponse struct {
    ID          uint   `json:"id"`
    FullName    string `json:"full_name"`
    PhoneNumber string `json:"phone_number"`
    Email       string `json:"email"`
}

func NewUserController(repo repository.UserRepository) *UserController {
    return &UserController{repo}
}

func (c *UserController) Create(ctx *gin.Context) {
    var user models.User
    if err := ctx.ShouldBindJSON(&user); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := c.repo.Create(&user); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    response := CreateUserResponse{
        ID:          user.ID,
        FullName:    user.FullName,
        PhoneNumber: user.PhoneNumber,
        Email:       user.Email,
    }

    ctx.JSON(http.StatusCreated, response)
}

func (c *UserController) GetAll(ctx *gin.Context) {
    users, err := c.repo.FindAll()
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, users)
}

func (c *UserController) GetByID(ctx *gin.Context) {
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }
    user, err := c.repo.FindByID(uint(id))
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, user)
}

func (c *UserController) Update(ctx *gin.Context) {
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    user, err := c.repo.FindByID(uint(id))
    if err != nil {
        if err == gorm.ErrRecordNotFound {
            ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
            return
        }
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    if err := ctx.ShouldBindJSON(user); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := c.repo.Update(user); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, user)
}

func (c *UserController) Delete(ctx *gin.Context) {
    id, err := strconv.Atoi(ctx.Param("id"))
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    if err := c.repo.Delete(uint(id)); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

// GetHistory fetches the Redis chat history for a specific ChatID
func (c *UserController) GetHistory(ctx *gin.Context) {
    chatID := ctx.Param("chat_id")
    
    if chatID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "chat_id parameter is required"})
        return
    }

    // Pass the context from the request so it can be used by the Redis client
    history, err := c.repo.GetChatHistory(ctx.Request.Context(), chatID)
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch chat history: " + err.Error()})
        return
    }
    
    ctx.JSON(http.StatusOK, history)
}

func (c *UserController) ClearHistory(ctx *gin.Context) {
    chatID := ctx.Param("id")
    
    if chatID == "" {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "id parameter is required"})
        return
    }

    if err := c.repo.DeleteChatHistory(ctx.Request.Context(), chatID); err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear history: " + err.Error()})
        return
    }
    
    ctx.JSON(http.StatusOK, gin.H{"message": "chat history cleared successfully"})
}