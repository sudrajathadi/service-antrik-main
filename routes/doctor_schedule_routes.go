package routes

import (
	"doctor-booking/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterDoctorScheduleRoutes(r *gin.Engine, ctrl *controllers.DoctorScheduleController) {
	schedules := r.Group("/api/schedules")
	{
		schedules.POST("", ctrl.Create)
		schedules.GET("", ctrl.GetAll)
		schedules.GET("/:id", ctrl.GetByID)
		schedules.PUT("/:id", ctrl.Update)
		schedules.DELETE("/:id", ctrl.Delete)
	}
}
