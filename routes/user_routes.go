package routes

import (
	"doctor-booking/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(r *gin.Engine, ctrl *controllers.UserController) {
	users := r.Group("/api/users")
	{
		users.POST("", ctrl.Create)
		users.GET("", ctrl.GetAll)
		users.GET("/:id", ctrl.GetByID)
		users.PUT("/:id", ctrl.Update)
		users.DELETE("/:id", ctrl.Delete)
	}
}
