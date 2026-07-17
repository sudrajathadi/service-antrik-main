package controllers

import (
	"net/http"

	"service-antrik-chatbot/chatbot"

	"github.com/gin-gonic/gin"
)

type ChatController struct {
	engine *chatbot.Engine
}

func NewChatController(engine *chatbot.Engine) *ChatController {
	return &ChatController{engine: engine}
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

	respondSuccess(ctx, http.StatusOK, "Chat processed successfully", response)
}
