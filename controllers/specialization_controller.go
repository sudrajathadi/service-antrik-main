package controllers

import (
	"net/http"
	"strconv"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SpecializationController struct {
	repo repository.SpecializationRepository
}

func NewSpecializationController(repo repository.SpecializationRepository) *SpecializationController {
	return &SpecializationController{repo}
}

func (c *SpecializationController) Create(ctx *gin.Context) {
	var spec models.Specialization
	if err := ctx.ShouldBindJSON(&spec); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.repo.Create(&spec); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, spec)
}

func (c *SpecializationController) GetAll(ctx *gin.Context) {
	specs, err := c.repo.FindAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, specs)
}

func (c *SpecializationController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}
	spec, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "specialization not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, spec)
}

func (c *SpecializationController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	spec, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "specialization not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := ctx.ShouldBindJSON(spec); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.repo.Update(spec); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, spec)
}

func (c *SpecializationController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "specialization deleted"})
}
