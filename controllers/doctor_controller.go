package controllers

import (
	"net/http"
	"strconv"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DoctorController struct {
	repo repository.DoctorRepository
}

func NewDoctorController(repo repository.DoctorRepository) *DoctorController {
	return &DoctorController{repo}
}

func (c *DoctorController) Create(ctx *gin.Context) {
	var doctor models.Doctor
	if err := ctx.ShouldBindJSON(&doctor); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid doctor request body", err.Error())
		return
	}
	if err := c.repo.Create(&doctor); err != nil {
		respondError(ctx, http.StatusUnprocessableEntity, "DOCTOR_CREATE_FAILED", "Doctor could not be created", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusCreated, "Doctor created successfully", doctor)
}

func (c *DoctorController) GetAll(ctx *gin.Context) {
	filter := repository.DoctorFilter{
		Specialization: queryAny(ctx, "specialization", "spesialisasi"),
		City:           ctx.Query("city"),
		Location:       queryAny(ctx, "location", "lokasi"),
	}

	doctors, err := c.repo.FindAllFiltered(filter)
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "DOCTORS_FETCH_FAILED", "Doctors could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Doctors fetched successfully", doctors)
}

func queryAny(ctx *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := ctx.Query(key); value != "" {
			return value
		}
	}
	return ""
}

func (c *DoctorController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_DOCTOR_ID", "Invalid doctor id", err.Error())
		return
	}
	doctor, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "DOCTOR_NOT_FOUND", "Doctor not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "DOCTOR_FETCH_FAILED", "Doctor could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Doctor fetched successfully", doctor)
}

func (c *DoctorController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_DOCTOR_ID", "Invalid doctor id", err.Error())
		return
	}
	doctor, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "DOCTOR_NOT_FOUND", "Doctor not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "DOCTOR_FETCH_FAILED", "Doctor could not be fetched", err.Error())
		return
	}
	if err := ctx.ShouldBindJSON(doctor); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid doctor request body", err.Error())
		return
	}
	if err := c.repo.Update(doctor); err != nil {
		respondError(ctx, http.StatusInternalServerError, "DOCTOR_UPDATE_FAILED", "Doctor could not be updated", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Doctor updated successfully", doctor)
}

func (c *DoctorController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_DOCTOR_ID", "Invalid doctor id", err.Error())
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		respondError(ctx, http.StatusInternalServerError, "DOCTOR_DELETE_FAILED", "Doctor could not be deleted", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Doctor deleted successfully", gin.H{"id": id})
}
