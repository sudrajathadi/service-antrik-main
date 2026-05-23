package controllers

import (
	"net/http"
	"strconv"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HospitalController struct {
	repo repository.HospitalRepository
}

func NewHospitalController(repo repository.HospitalRepository) *HospitalController {
	return &HospitalController{repo}
}

func (c *HospitalController) Create(ctx *gin.Context) {
	var hospital models.Hospital
	if err := ctx.ShouldBindJSON(&hospital); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.repo.Create(&hospital); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, hospital)
}

func (c *HospitalController) GetAll(ctx *gin.Context) {
	hospitals, err := c.repo.FindAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, hospitals)
}

func (c *HospitalController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id when getting by id"})
		return
	}
	hospital, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, hospital)
}

func (c *HospitalController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	hospital, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "hospital not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindJSON(hospital); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.repo.Update(hospital); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, hospital)
}

func (c *HospitalController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "hospital deleted"})
}
