package controllers

import (
	"net/http"
	"strconv"
	"time"

	"service-antrik-chatbot/models"
	"service-antrik-chatbot/repository"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AppointmentController struct {
	repo repository.AppointmentRepository
}

type CreateAppointmentResponse struct {
	ID              uint                     `json:"id"`
	UserID          uint                     `json:"user_id"`
	DoctorID        uint                     `json:"doctor_id"`
	HospitalID      uint                     `json:"hospital_id"`
	AppointmentDate time.Time                `json:"appointment_date"`
	AppointmentTime string                   `json:"appointment_time"`
	SymptomsNote    string                   `json:"symptoms_note"`
	Status          models.AppointmentStatus `json:"status"`
	CreatedAt       time.Time                `json:"created_at"`
}

func NewAppointmentController(repo repository.AppointmentRepository) *AppointmentController {
	return &AppointmentController{repo}
}

func (c *AppointmentController) Create(ctx *gin.Context) {
	var appointment models.Appointment

	if err := ctx.ShouldBindJSON(&appointment); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid appointment request body", err.Error())
		return
	}

	if appointment.Status == "" {
		appointment.Status = models.StatusPending
	}

	if err := c.repo.Create(&appointment); err != nil {
		respondError(ctx, http.StatusUnprocessableEntity, "APPOINTMENT_CREATE_FAILED", "Appointment could not be created", err.Error())
		return
	}

	response := CreateAppointmentResponse{
		ID:              appointment.ID,
		UserID:          appointment.UserID,
		DoctorID:        appointment.DoctorID,
		HospitalID:      appointment.HospitalID,
		AppointmentDate: appointment.AppointmentDate,
		AppointmentTime: appointment.AppointmentTime,
		SymptomsNote:    appointment.SymptomsNote,
		Status:          appointment.Status,
		CreatedAt:       appointment.CreatedAt,
	}

	respondSuccess(ctx, http.StatusCreated, "Appointment created successfully", response)
}

func (c *AppointmentController) GetAll(ctx *gin.Context) {
	appointments, err := c.repo.FindAll()
	if err != nil {
		respondError(ctx, http.StatusInternalServerError, "APPOINTMENTS_FETCH_FAILED", "Appointments could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Appointments fetched successfully", appointments)
}

func (c *AppointmentController) GetByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_APPOINTMENT_ID", "Invalid appointment id", err.Error())
		return
	}
	appointment, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "APPOINTMENT_NOT_FOUND", "Appointment not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "APPOINTMENT_FETCH_FAILED", "Appointment could not be fetched", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Appointment fetched successfully", appointment)
}

func (c *AppointmentController) Update(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_APPOINTMENT_ID", "Invalid appointment id", err.Error())
		return
	}
	appointment, err := c.repo.FindByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			respondError(ctx, http.StatusNotFound, "APPOINTMENT_NOT_FOUND", "Appointment not found", err.Error())
			return
		}
		respondError(ctx, http.StatusInternalServerError, "APPOINTMENT_FETCH_FAILED", "Appointment could not be fetched", err.Error())
		return
	}
	if err := ctx.ShouldBindJSON(appointment); err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid appointment request body", err.Error())
		return
	}
	if err := c.repo.Update(appointment); err != nil {
		respondError(ctx, http.StatusInternalServerError, "APPOINTMENT_UPDATE_FAILED", "Appointment could not be updated", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Appointment updated successfully", appointment)
}

func (c *AppointmentController) Delete(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		respondError(ctx, http.StatusBadRequest, "INVALID_APPOINTMENT_ID", "Invalid appointment id", err.Error())
		return
	}
	if err := c.repo.Delete(uint(id)); err != nil {
		respondError(ctx, http.StatusInternalServerError, "APPOINTMENT_DELETE_FAILED", "Appointment could not be deleted", err.Error())
		return
	}
	respondSuccess(ctx, http.StatusOK, "Appointment deleted successfully", gin.H{"id": id})
}
