package routes

import (
	"service-antrik-chatbot/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine, ctrl *controllers.UserController) {
	users := r.Group("/api/users")
	{
		users.POST("", ctrl.Create)
		users.DELETE("/:id", ctrl.Delete)
		users.GET("", ctrl.GetAll)
		users.GET("/chat/:chat_id", ctrl.GetHistory)
		users.GET("/:id", ctrl.GetByID)
		users.PUT("/:id", ctrl.Update)
	}
}
