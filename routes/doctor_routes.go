package routes

import (
	"service-antrik-chatbot/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterDoctorRoutes(r *gin.Engine, ctrl *controllers.DoctorController) {
	doctors := r.Group("/api/doctors")
	{
		doctors.POST("", ctrl.Create)
		doctors.GET("", ctrl.GetAll)
		doctors.GET("/:id", ctrl.GetByID)
		doctors.PUT("/:id", ctrl.Update)
		doctors.DELETE("/:id", ctrl.Delete)
	}
}
