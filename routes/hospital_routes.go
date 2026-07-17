package routes

import (
	"service-antrik-chatbot/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterHospitalRoutes(r *gin.Engine, ctrl *controllers.HospitalController) {
	hospitals := r.Group("/api/hospitals")
	{
		hospitals.POST("", ctrl.Create)
		hospitals.GET("", ctrl.GetAll)
		hospitals.GET("/search", ctrl.SearchByCity)
		hospitals.GET("/:id", ctrl.GetByID)
		hospitals.PUT("/:id", ctrl.Update)
		hospitals.DELETE("/:id", ctrl.Delete)
	}
}
