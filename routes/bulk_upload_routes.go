package routes

import (
	"service-antrik-chatbot/controllers"

	"github.com/gin-gonic/gin"
)

func RegisterBulkUploadRoutes(r *gin.Engine, ctrl *controllers.BulkUploadController) {
	bulkUpload := r.Group("/api/bulk-upload")
	{
		bulkUpload.POST("/:table", ctrl.UploadCSV)
		bulkUpload.GET("/templates/:table", ctrl.DownloadTemplate)
	}

	r.POST("/api/bulk-upload-url/:table", ctrl.UploadCSVFromURL)
}
