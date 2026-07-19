package repository

import (
	"context"
	"encoding/json"
	"service-antrik-chatbot/models"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type LangChainMessage struct {
	Type      string `json:"type"`
	Timestamp string `json:"timestamp,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Data      struct {
		Content   string `json:"content"`
		Timestamp string `json:"timestamp,omitempty"`
		CreatedAt string `json:"created_at,omitempty"`
	} `json:"data"`
}

type FrontendMessage struct {
	Role      string `json:"role"`
	Category  string `json:"category"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp,omitempty"`
	Order     int    `json:"order"`
}

type UserRepository interface {
	Create(user *models.User) error
	FindAll() ([]models.User, error)
	FindByID(id uint) (*models.User, error)
	FindByChatID(chatID string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	GetChatHistory(ctx context.Context, chatID string) ([]FrontendMessage, error)
	AppendChatHistory(ctx context.Context, chatID string, messages ...FrontendMessage) error
	DeleteChatHistory(ctx context.Context, chatID string) error
}

type userRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

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
	for index, msgStr := range messagesJSON {
		message, ok := decodeHistoryMessage(msgStr, index)
		if !ok {
			continue
		}
		history = append(history, message)
	}

	if history == nil {
		history = []FrontendMessage{}
	}

	sortFrontendMessages(history)

	return history, nil
}

func (r *userRepository) AppendChatHistory(ctx context.Context, chatID string, messages ...FrontendMessage) error {
	if len(messages) == 0 {
		return nil
	}

	values := make([]interface{}, 0, len(messages))
	for index, message := range messages {
		if message.Category == "" {
			message.Category = "chat"
		}
		if message.Timestamp == "" {
			message.Timestamp = time.Now().Add(time.Duration(index) * time.Nanosecond).Format(time.RFC3339Nano)
		}
		payload, err := json.Marshal(message)
		if err != nil {
			return err
		}
		values = append(values, payload)
	}

	return r.redis.RPush(ctx, chatID, values...).Err()
}

func decodeHistoryMessage(value string, index int) (FrontendMessage, bool) {
	if message, ok := decodeFrontendMessage(value, index); ok {
		return message, true
	}
	return decodeLangChainMessage(value, index)
}

func decodeFrontendMessage(value string, index int) (FrontendMessage, bool) {
	var message FrontendMessage
	if err := json.Unmarshal([]byte(value), &message); err != nil {
		return FrontendMessage{}, false
	}
	if strings.TrimSpace(message.Content) == "" {
		return FrontendMessage{}, false
	}
	if message.Category == "" {
		message.Category = "chat"
	}
	message.Order = index
	return message, true
}

func decodeLangChainMessage(value string, index int) (FrontendMessage, bool) {
	var lcMsg LangChainMessage
	if err := json.Unmarshal([]byte(value), &lcMsg); err != nil {
		return FrontendMessage{}, false
	}
	if strings.TrimSpace(lcMsg.Data.Content) == "" {
		return FrontendMessage{}, false
	}

	role := "user"
	if lcMsg.Type == "ai" || lcMsg.Type == "tool" {
		role = "agent"
	}

	category := "chat"
	contentStr := strings.TrimSpace(lcMsg.Data.Content)
	if role == "agent" && strings.HasPrefix(contentStr, "Calling ") {
		category = "system"
	}
	if strings.HasPrefix(contentStr, "[{") || strings.HasPrefix(contentStr, "{\"") {
		category = "system"
		role = "agent"
	}

	return FrontendMessage{
		Role:      role,
		Category:  category,
		Content:   lcMsg.Data.Content,
		Timestamp: firstNonEmpty(lcMsg.Timestamp, lcMsg.CreatedAt, lcMsg.Data.Timestamp, lcMsg.Data.CreatedAt),
		Order:     index,
	}, true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func sortFrontendMessages(messages []FrontendMessage) {
	sort.SliceStable(messages, func(i, j int) bool {
		leftTime, leftOK := parseMessageTime(messages[i].Timestamp)
		rightTime, rightOK := parseMessageTime(messages[j].Timestamp)

		if leftOK && rightOK && !leftTime.Equal(rightTime) {
			return leftTime.Before(rightTime)
		}

		return messages[i].Order > messages[j].Order
	})
}

func parseMessageTime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}

	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, true
		}
	}

	unixValue, err := strconv.ParseInt(value, 10, 64)
	if err == nil {
		if unixValue > 1_000_000_000_000 {
			return time.UnixMilli(unixValue), true
		}

		return time.Unix(unixValue, 0), true
	}

	return time.Time{}, false
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

func (r *userRepository) FindByChatID(chatID string) (*models.User, error) {
	var user models.User
	err := r.db.Where("chat_id = ?", chatID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
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

func (r *userRepository) DeleteChatHistory(ctx context.Context, chatID string) error {
	// Make sure this matches how you are fetching the key!
	// (e.g., key := chatID  OR  key := fmt.Sprintf("chat_history:%s", chatID))
	key := chatID

	return r.redis.Del(ctx, key).Err()
}
