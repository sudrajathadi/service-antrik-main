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
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid hospital request body", err.Error())
		return
	}
	if err := c.repo.Create(&hospital); err != nil {
		respondError(ctx, http.StatusUnprocessableEntity, "HOSPITAL_CREATE_FAILED", "Hospital could not be created", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusCreated, "Hospital created successfully", hospital)
}

func (c *HospitalController) GetAll(ctx *gin.Context) {
	hospitals, err := c.repo.FindAll()
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "HOSPITALS_FETCH_FAILED", "Hospitals could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Hospitals fetched successfully", hospitals)
}

func (c *HospitalController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_HOSPITAL_ID", "Invalid hospital id", err.Error())
		return
	}
	hospital, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "HOSPITAL_NOT_FOUND", "Hospital not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "HOSPITAL_FETCH_FAILED", "Hospital could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Hospital fetched successfully", hospital)
}

func (c *HospitalController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_HOSPITAL_ID", "Invalid hospital id", err.Error())
		return
	}
	hospital, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "HOSPITAL_NOT_FOUND", "Hospital not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "HOSPITAL_FETCH_FAILED", "Hospital could not be fetched", err.Error())
		return
	}
	if err := ctx.ShouldBindJSON(hospital); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid hospital request body", err.Error())
		return
	}
	if err := c.repo.Update(hospital); err != nil {
		respondError(ctx, http.StatusInternalServerError, "HOSPITAL_UPDATE_FAILED", "Hospital could not be updated", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Hospital updated successfully", hospital)
}

func (c *HospitalController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_HOSPITAL_ID", "Invalid hospital id", err.Error())
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		respondError(ctx, http.StatusInternalServerError, "HOSPITAL_DELETE_FAILED", "Hospital could not be deleted", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Hospital deleted successfully", gin.H{"id": id})
}
