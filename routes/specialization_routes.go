package routes

import (
	"doctor-booking/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterSpecializationRoutes(r *gin.Engine, ctrl *controllers.SpecializationController) {
	specs := r.Group("/api/specializations")
	{
		specs.POST("", ctrl.Create)
		specs.GET("", ctrl.GetAll)
		specs.GET("/:id", ctrl.GetByID)
		specs.PUT("/:id", ctrl.Update)
		specs.DELETE("/:id", ctrl.Delete)
	}
}
