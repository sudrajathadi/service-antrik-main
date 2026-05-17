package repository

import (
	"doctor-booking/models"

	"gorm.io/gorm"
)

type AppointmentRepository interface {
	Create(appointment *models.Appointment) error
	FindAll() ([]models.Appointment, error)
	FindByID(id uint) (*models.Appointment, error)
	Update(appointment *models.Appointment) error
	Delete(id uint) error
}

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) AppointmentRepository {
	return &appointmentRepository{db}
}

func (r *appointmentRepository) Create(appointment *models.Appointment) error {
	return r.db.Create(appointment).Error
}

func (r *appointmentRepository) FindAll() ([]models.Appointment, error) {
	var appointments []models.Appointment
	err := r.db.Preload("User").Preload("Doctor").Preload("Hospital").Find(&appointments).Error
	return appointments, err
}

func (r *appointmentRepository) FindByID(id uint) (*models.Appointment, error) {
	var appointment models.Appointment
	err := r.db.Preload("User").Preload("Doctor").Preload("Hospital").First(&appointment, id).Error
	if err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepository) Update(appointment *models.Appointment) error {
	return r.db.Save(appointment).Error
}

func (r *appointmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.Appointment{}, id).Error
}
