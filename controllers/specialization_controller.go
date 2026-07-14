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
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid specialization request body", err.Error())
		return
	}
	if err := c.repo.Create(&spec); err != nil {
		respondError(ctx, http.StatusUnprocessableEntity, "SPECIALIZATION_CREATE_FAILED", "Specialization could not be created", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusCreated, "Specialization created successfully", spec)
}

func (c *SpecializationController) GetAll(ctx *gin.Context) {
	specs, err := c.repo.FindAll()
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "SPECIALIZATIONS_FETCH_FAILED", "Specializations could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Specializations fetched successfully", specs)
}

func (c *SpecializationController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_SPECIALIZATION_ID", "Invalid specialization id", err.Error())
		return
	}
	spec, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "SPECIALIZATION_NOT_FOUND", "Specialization not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "SPECIALIZATION_FETCH_FAILED", "Specialization could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Specialization fetched successfully", spec)
}

func (c *SpecializationController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_SPECIALIZATION_ID", "Invalid specialization id", err.Error())
		return
	}
	spec, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "SPECIALIZATION_NOT_FOUND", "Specialization not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "SPECIALIZATION_FETCH_FAILED", "Specialization could not be fetched", err.Error())
		return
	}
	if err := ctx.ShouldBindJSON(spec); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid specialization request body", err.Error())
		return
	}
	if err := c.repo.Update(spec); err != nil {
		respondError(ctx, http.StatusInternalServerError, "SPECIALIZATION_UPDATE_FAILED", "Specialization could not be updated", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Specialization updated successfully", spec)
}

func (c *SpecializationController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_SPECIALIZATION_ID", "Invalid specialization id", err.Error())
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		respondError(ctx, http.StatusInternalServerError, "SPECIALIZATION_DELETE_FAILED", "Specialization could not be deleted", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Specialization deleted successfully", gin.H{"id": id})
}
