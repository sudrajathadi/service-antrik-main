package repository

import (
	"context"
	"encoding/json"
	"service-antrik-chatbot/models"
	"strings"

	"github.com/redis/go-redis/v9" // 1. Added Redis import
	"gorm.io/gorm"
)

// 2. Added the required message structs
// (You can also move these to your models package and import them)
type LangChainMessage struct {
    Type string `json:"type"`
    Data struct {
        Content string `json:"content"`
    } `json:"data"`
}

type FrontendMessage struct {
    Role     string `json:"role"`     // "user" or "agent"
    Category string `json:"category"` // NEW: "chat" or "system"
    Content  string `json:"content"`
}

// 3. Added GetChatHistory to the interface
type UserRepository interface {
    Create(user *models.User) error
    FindAll() ([]models.User, error)
    FindByID(id uint) (*models.User, error)
    Update(user *models.User) error
    Delete(id uint) error
    GetChatHistory(ctx context.Context, chatID string) ([]FrontendMessage, error)
}

// 4. Added the Redis client to the repository struct
type userRepository struct {
    db    *gorm.DB
    redis *redis.Client 
}

// 5. Updated constructor to require BOTH Gorm and Redis
func NewUserRepository(db *gorm.DB, redisClient *redis.Client) UserRepository {
    return &userRepository{
        db:    db,
        redis: redisClient,
    }
}

func (r *userRepository) GetChatHistory(ctx context.Context, chatID string) ([]FrontendMessage, error) {
    key := chatID 
    
    messagesJSON, err := r.redis.LRange(ctx, key, 0, -1).Result()
    if err != nil {
        return nil, err
    }

    var history []FrontendMessage
    for _, msgStr := range messagesJSON {
        var lcMsg LangChainMessage
        
        if err := json.Unmarshal([]byte(msgStr), &lcMsg); err != nil {
            continue 
        }

        // 1. Map roles (catch "tool" as an agent/system interaction)
        role := "user"
        if lcMsg.Type == "ai" || lcMsg.Type == "tool" {
            role = "agent"
        }

        // 2. Categorize the message
        category := "chat"
        contentStr := strings.TrimSpace(lcMsg.Data.Content)

        // Rule A: Is it the AI announcing a tool call?
        if role == "agent" && strings.HasPrefix(contentStr, "Calling ") {
            category = "system"
        }

        // Rule B: Is it the raw JSON output from the HTTP request?
        // If it starts with JSON brackets, it's database data, not a human chatting.
        if strings.HasPrefix(contentStr, "[{") || strings.HasPrefix(contentStr, "{\"") {
            category = "system"
            role = "agent" // Force it to agent/system so it doesn't look like the user typed it
        }

        history = append(history, FrontendMessage{
            Role:     role,
            Category: category,
            Content:  lcMsg.Data.Content,
        })
    }

    if history == nil {
        history = []FrontendMessage{}
    }

    return history, nil
}

// --- Standard CRUD Methods Below ---

func (r *userRepository) Create(user *models.User) error {
    return r.db.Create(user).Error
}

func (r *userRepository) FindAll() ([]models.User, error) {
    var users []models.User
    err := r.db.Find(&users).Error
    return users, err
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
    var user models.User
    err := r.db.First(&user, id).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
    return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
    return r.db.Delete(&models.User{}, id).Error
}