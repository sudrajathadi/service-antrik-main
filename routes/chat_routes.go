package routes

import (
	"service-antrik-chatbot/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterChatRoutes(r *gin.Engine, ctrl *controllers.ChatController) {
	chat := r.Group("/api/chat")
	{
		chat.POST("", ctrl.Reply)
	}
}
