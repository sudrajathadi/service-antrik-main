package routes

import (
	"doctor-booking/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterAppointmentRoutes(r *gin.Engine, ctrl *controllers.AppointmentController) {
	appointments := r.Group("/api/appointments")
	{
		appointments.POST("", ctrl.Create)
		appointments.GET("", ctrl.GetAll)
		appointments.GET("/:id", ctrl.GetByID)
		appointments.PUT("/:id", ctrl.Update)
		appointments.DELETE("/:id", ctrl.Delete)
	}
}
