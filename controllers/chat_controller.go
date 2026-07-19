package controllers

import (
	"net/http"
	"time"

	"service-antrik-chatbot/chatbot"
	"service-antrik-chatbot/repository"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	engine *chatbot.Engine
	users  repository.UserRepository
}

func NewChatController(engine *chatbot.Engine, userRepo repository.UserRepository) *ChatController {
	return &ChatController{
		engine: engine,
		users:  userRepo,
	}
}

func (c *ChatController) Reply(ctx *gin.Context) {
	var request chatbot.ChatRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_CHAT_REQUEST", "Invalid chat request body", err.Error())
		return
	}

	response, err := c.engine.Reply(request)
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "CHAT_PROCESSING_FAILED", "Chat message could not be processed", err.Error())
		return
	}

	if err := c.saveHistory(ctx, request, response); err != nil {
		respondError(ctx, http.StatusInternalServerError, "CHAT_HISTORY_SAVE_FAILED", "Chat history could not be saved", err.Error())
		return
	}

	respondSuccess(ctx, http.StatusOK, "Chat processed successfully", response)
}

func (c *ChatController) saveHistory(ctx *gin.Context, request chatbot.ChatRequest, response chatbot.ChatResponse) error {
	now := time.Now()
	return c.users.AppendChatHistory(
		ctx.Request.Context(),
		request.ChatID,
		repository.FrontendMessage{
			Role:      "user",
			Category:  "chat",
			Content:   request.Message,
			Timestamp: now.Format(time.RFC3339Nano),
		},
		repository.FrontendMessage{
			Role:      "agent",
			Category:  "chat",
			Content:   response.Reply,
			Timestamp: now.Add(time.Nanosecond).Format(time.RFC3339Nano),
		},
	)
}
