package controllers

import (
	"net/http"
	"strconv"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppointmentController struct {
	repo repository.AppointmentRepository
}

func NewAppointmentController(repo repository.AppointmentRepository) *AppointmentController {
	return &AppointmentController{repo}
}

func (c *AppointmentController) Create(ctx *gin.Context) {
	var appointment models.Appointment
	if err := ctx.ShouldBindJSON(&appointment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if appointment.Status == "" {
		appointment.Status = models.StatusPending
	}
	if err := c.repo.Create(&appointment); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, appointment)
}

func (c *AppointmentController) GetAll(ctx *gin.Context) {
	appointments, err := c.repo.FindAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, appointments)
}

func (c *AppointmentController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid idnya"})
		return
	}
	appointment, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, appointment)
}

func (c *AppointmentController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	appointment, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "appointment not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindJSON(appointment); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.repo.Update(appointment); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, appointment)
}

func (c *AppointmentController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "appointment deleted"})
}
